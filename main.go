package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"crypto-indicator/cache"
	"crypto-indicator/calculator"
	"crypto-indicator/fetcher"
)

var dataCache = cache.New(30 * time.Second)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{"code":200,"message":"ok"}`)
}

func klineHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	interval := r.URL.Query().Get("interval")
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if interval == "" {
		interval = "1d"
	}

	klines, err := fetcher.FetchKlines(symbol, interval, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(klines)
}

func indicatorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	symbol := r.URL.Query().Get("symbol")
	interval := r.URL.Query().Get("interval")
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if interval == "" {
		interval = "1d"
	}

	cacheKey := symbol + ":" + interval

	// Cache hit — return immediately
	if cached, ok := dataCache.Get(cacheKey); ok {
		json.NewEncoder(w).Encode(cached)
		return
	}

	klines, err := fetcher.FetchKlines(symbol, interval, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	ma5 := calculator.MA(closes, 5)
	ma20 := calculator.MA(closes, 20)
	rsi := calculator.RSI(closes, 14)
	macdResult := calculator.MACD(closes, 12, 26, 9)
	bollinger := calculator.CalcBollingerBands(klines, 20, 2.0)
	atr := calculator.CalcATR(klines, 14)
	kdj := calculator.CalcKDJ(klines, 9)
	volData := calculator.CalcVolumeMA(klines, 20)
	obv := calculator.CalcOBV(klines)

	// Signal score for the last candle
	var signal calculator.SignalScore
	if len(closes) > 0 {
		signal = calculator.CalcSignalScore(
			closes,
			ma5,
			ma20,
			rsi,
			macdResult.MACD, // histogram
			bollinger,
			kdj,
			atr,
			volData,
			len(closes)-1,
		)
	}

	result := map[string]interface{}{
		"klines": klines,
		"indicators": map[string]interface{}{
			"ma5":  ma5,
			"ma20": ma20,
			"rsi":  rsi,
			"macd": map[string]interface{}{
				"macd_line":   macdResult.DIF,
				"signal_line": macdResult.DEA,
				"histogram":   macdResult.MACD,
			},
			"bollinger":   bollinger,
			"atr":         atr,
			"kdj":         kdj,
			"volume_data": volData,
			"obv":         obv,
			"signal":      signal,
		},
	}

	dataCache.Set(cacheKey, result)
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/ping", pingHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/api/kline", klineHandler)
	http.HandleFunc("/api/indicator", indicatorHandler)

	fmt.Println("服务启动在 http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
