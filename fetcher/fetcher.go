package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"crypto-indicator/model"
)

// FetchKlines 从 Binance 拉取 K 线数据
func FetchKlines(symbol, interval string, limit int) ([]model.Kline, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	url := fmt.Sprintf(
		"https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 Binance 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Binance 返回异常状态码: %d", resp.StatusCode)
	}

	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	klines := make([]model.Kline, 0, len(raw))
	for _, item := range raw {
		if len(item) < 6 {
			continue
		}
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

func parseFloat(v interface{}) float64 {
	s, ok := v.(string)
	if !ok {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
