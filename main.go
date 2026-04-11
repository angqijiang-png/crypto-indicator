package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"crypto-indicator/calculator"
	"crypto-indicator/fetcher"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{"code":200,"message":"ok"}`)
}

func parseLimit(r *http.Request) int {
	v := r.URL.Query().Get("limit")
	if v == "" {
		return 200
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 || n > 1000 {
		return 200
	}
	return n
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

	klines, err := fetcher.FetchKlines(symbol, interval, parseLimit(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(klines)
}

func indicatorHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	interval := r.URL.Query().Get("interval")
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if interval == "" {
		interval = "1d"
	}

	klines, err := fetcher.FetchKlines(symbol, interval, parseLimit(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
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
