package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"crypto-indicator/model"
)

func FetchKlines(symbol, interval string, limit int) ([]model.Kline, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	url := fmt.Sprintf(
		"https://data-api.binance.vision/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}

	klines := make([]model.Kline, 0, len(raw))
	for _, r := range raw {
		openTime := int64(r[0].(float64))
		open, _ := strconv.ParseFloat(r[1].(string), 64)
		high, _ := strconv.ParseFloat(r[2].(string), 64)
		low, _ := strconv.ParseFloat(r[3].(string), 64)
		close_, _ := strconv.ParseFloat(r[4].(string), 64)
		volume, _ := strconv.ParseFloat(r[5].(string), 64)
		klines = append(klines, model.Kline{
			OpenTime: openTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close_,
			Volume:   volume,
		})
	}
	return klines, nil
}
