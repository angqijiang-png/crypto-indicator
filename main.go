package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"crypto-indicator/calculator"
	"crypto-indicator/fetcher"
)

// ---- CORS & 通用中间件 ----

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON 编码失败: %v", err)
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// ---- 参数解析 ----

func getQueryParams(r *http.Request) (symbol, interval string, limit int) {
	symbol = strings.ToUpper(r.URL.Query().Get("symbol"))
	interval = r.URL.Query().Get("interval")
	limitStr := r.URL.Query().Get("limit")

	if symbol == "" {
		symbol = "BTCUSDT"
	}
	if interval == "" {
		interval = "1d"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 200
	}
	return symbol, interval, limit
}

// ---- Handlers ----

func pingHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "ok",
	})
}

func klineHandler(w http.ResponseWriter, r *http.Request) {
	symbol, interval, limit := getQueryParams(r)

	klines, err := fetcher.FetchKlines(symbol, interval, limit)
	if err != nil {
		log.Printf("[kline] 拉取失败 %s/%s: %v", symbol, interval, err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	log.Printf("[kline] %s %s %d 条", symbol, interval, len(klines))
	writeJSON(w, http.StatusOK, klines)
}

func indicatorHandler(w http.ResponseWriter, r *http.Request) {
	symbol, interval, limit := getQueryParams(r)

	klines, err := fetcher.FetchKlines(symbol, interval, limit)
	if err != nil {
		log.Printf("[indicator] 拉取失败 %s/%s: %v", symbol, interval, err)
		writeError(w, http.StatusBadGateway, err.Error())
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
	}

	log.Printf("[indicator] %s %s %d 条 ok", symbol, interval, len(klines))
	writeJSON(w, http.StatusOK, result)
}

// ---- Main ----

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// 静态文件（前端 Dashboard）
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// API 路由
	mux.HandleFunc("/ping", withCORS(pingHandler))
	mux.HandleFunc("/api/kline", withCORS(klineHandler))
	mux.HandleFunc("/api/indicator", withCORS(indicatorHandler))

	addr := ":" + port
	fmt.Printf("🚀 crypto-indicator 启动在 http://localhost%s\n", addr)
	fmt.Printf("📊 Dashboard: http://localhost%s\n", addr)
	fmt.Printf("📡 API:       http://localhost%s/api/indicator?symbol=BTCUSDT&interval=1d\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
