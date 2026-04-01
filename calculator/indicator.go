package calculator

// MA 计算简单移动平均线
// prices: 收盘价数组，period: 周期（比如 5、10、20）
func MA(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - (period - 1); j <= i; j++ {
			sum += prices[j]
		}
		result[i] = sum / float64(period)
	}
	return result
}

// EMA 计算指数移动平均线（RSI 和 MACD 的基础）
func EMA(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	k := 2.0 / float64(period+1)
	result[period-1] = prices[period-1]
	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*k + result[i-1]*(1-k)
	}
	return result
}

// RSI 计算相对强弱指数
func RSI(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	for i := period; i < len(prices); i++ {
		gains, losses := 0.0, 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := prices[j] - prices[j-1]
			if diff > 0 {
				gains += diff
			} else {
				losses -= diff
			}
		}
		avgGain := gains / float64(period)
		avgLoss := losses / float64(period)
		if avgLoss == 0 {
			result[i] = 100
			continue
		}
		rs := avgGain / avgLoss
		result[i] = 100 - (100 / (1 + rs))
	}
	return result
}

// MACDResult 存储 MACD 的三条线
type MACDResult struct {
	DIF  []float64 `json:"dif"`  // 快线 - 慢线
	DEA  []float64 `json:"dea"`  // DIF 的 EMA（信号线）
	MACD []float64 `json:"macd"` // 柱状图 = DIF - DEA
}

// MACD 计算 MACD 指标
// 标准参数：fast=12, slow=26, signal=9
func MACD(prices []float64, fast, slow, signal int) MACDResult {
	emaFast := EMA(prices, fast)
	emaSlow := EMA(prices, slow)

	dif := make([]float64, len(prices))
	for i := slow - 1; i < len(prices); i++ {
		dif[i] = emaFast[i] - emaSlow[i]
	}

	dea := EMA(dif, signal)

	macd := make([]float64, len(prices))
	for i := range prices {
		macd[i] = (dif[i] - dea[i]) * 2
	}

	return MACDResult{DIF: dif, DEA: dea, MACD: macd}
}
