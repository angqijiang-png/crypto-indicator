package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"crypto-indicator/model"
)

// FetchKlines 从 Binance 拉取 K 线数据
// symbol: 比如 "BTCUSDT"
// interval: 比如 "1d" "4h" "1h"
func FetchKlines(symbol, interval string) ([]model.Kline, error) {
	url := fmt.Sprintf(
		"https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=100",
		symbol, interval,
	)

	// 发 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 Binance 失败: %w", err)
	}
	defer resp.Body.Close()

	// Binance 返回的是 [][]interface{}，每个元素是混合类型的数组
	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 把原始数据转换成我们自己定义的 Kline 结构
	var klines []model.Kline
	for _, item := range raw {
		k := model.Kline{
			OpenTime: int64(item[0].(float64)),
			Open:     parseFloat(item[1]),
			High:     parseFloat(item[2]),
			Low:      parseFloat(item[3]),
			Close:    parseFloat(item[4]),
			Volume:   parseFloat(item[5]),
		}
		klines = append(klines, k)
	}

	return klines, nil
}

// parseFloat 把 interface{} 转成 float64
// Binance 把数字用字符串返回，比如 "71000.00"
func parseFloat(v interface{}) float64 {
	s, ok := v.(string)
	if !ok {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
