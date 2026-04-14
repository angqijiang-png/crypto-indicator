package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"crypto-indicator/calculator"
	"crypto-indicator/fetcher"
)

// parseLimit reads the "limit" query param; defaults to 100.
func parseLimit(r *http.Request) int {
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 100
}

const wsGUID         = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
const wsPushInterval = 30 * time.Second

// wsHandshake upgrades an HTTP connection to WebSocket (RFC 6455).
func wsHandshake(w http.ResponseWriter, r *http.Request) (net.Conn, *bufio.ReadWriter, error) {
	key := r.Header.Get("Sec-Websocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return nil, nil, fmt.Errorf("missing key")
	}
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		return nil, nil, fmt.Errorf("no hijack")
	}
	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	bufrw.WriteString(resp)
	bufrw.Flush()
	return conn, bufrw, nil
}

// wsSendText sends a WebSocket text frame (server → client, no masking).
func wsSendText(conn net.Conn, data []byte) error {
	l := len(data)
	var header []byte
	switch {
	case l < 126:
		header = []byte{0x81, byte(l)}
	case l < 65536:
		header = []byte{0x81, 126, byte(l >> 8), byte(l)}
	default:
		header = make([]byte, 10)
		header[0], header[1] = 0x81, 127
		binary.BigEndian.PutUint64(header[2:], uint64(l))
	}
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := conn.Write(header); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}

// wsReadFrame reads one WebSocket frame (client → server).
// Returns (opcode, payload, error). Handles masking per RFC 6455.
func wsReadFrame(r *bufio.Reader) (byte, []byte, error) {
	b0, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	opcode := b0 & 0x0f

	b1, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	masked := b1&0x80 != 0
	plen := int(b1 & 0x7f)

	if plen == 126 {
		buf := make([]byte, 2)
		if _, err = r.Read(buf); err != nil {
			return 0, nil, err
		}
		plen = int(binary.BigEndian.Uint16(buf))
	} else if plen == 127 {
		buf := make([]byte, 8)
		if _, err = r.Read(buf); err != nil {
			return 0, nil, err
		}
		plen = int(binary.BigEndian.Uint64(buf))
	}

	var mask [4]byte
	if masked {
		if _, err = r.Read(mask[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, plen)
	if _, err = r.Read(payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return opcode, payload, nil
}

// buildPayload fetches klines and computes all indicators.
func buildPayload(symbol, interval string, limit int) ([]byte, error) {
	klines, err := fetcher.FetchKlines(symbol, interval, limit)
	if err != nil {
		return nil, err
	}
	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}
	result := map[string]interface{}{
		"symbol":   symbol,
		"interval": interval,
		"klines":   klines,
		"ma5":      calculator.MA(closes, 5),
		"ma20":     calculator.MA(closes, 20),
		"rsi14":    calculator.RSI(closes, 14),
		"macd":     calculator.MACD(closes, 12, 26, 9),
		"bb20":     calculator.BollingerBands(closes, 20, 2.0),
	}
	return json.Marshal(result)
}

// wsIndicatorHandler handles WebSocket connections on /ws/indicator.
// It pushes fresh indicator data immediately on connect, then every 30 s.
func wsIndicatorHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	interval := r.URL.Query().Get("interval")
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if interval == "" {
		interval = "1d"
	}
	limit := parseLimit(r)

	conn, bufrw, err := wsHandshake(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	// Push immediately on connect.
	data, err := buildPayload(symbol, interval, limit)
	if err != nil || wsSendText(conn, data) != nil {
		return
	}

	// done is closed when the client disconnects.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn.SetReadDeadline(time.Now().Add(wsPushInterval + 15*time.Second))
			op, payload, err := wsReadFrame(bufrw.Reader)
			if err != nil {
				return
			}
			switch op {
			case 8: // close frame — echo it and exit
				conn.Write([]byte{0x88, 0x00})
				return
			case 9: // ping → pong
				pong := make([]byte, 2+len(payload))
				pong[0], pong[1] = 0x8a, byte(len(payload))
				copy(pong[2:], payload)
				conn.Write(pong)
			}
		}
	}()

	ticker := time.NewTicker(wsPushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			data, err := buildPayload(symbol, interval, limit)
			if err != nil {
				continue // skip this tick if Binance is slow
			}
			if wsSendText(conn, data) != nil {
				return
			}
		}
	}
}
