package calculator

// MA 简单移动平均线
func MA(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	if len(prices) < period {
		return result
	}
	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - (period - 1); j <= i; j++ {
			sum += prices[j]
		}
		result[i] = sum / float64(period)
	}
	return result
}

// EMA 指数移动平均线（用 SMA 做种子，修复偏差问题）
func EMA(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	if len(prices) < period {
		return result
	}
	// 用前 period 个价格的 SMA 作为第一个 EMA（标准做法）
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	result[period-1] = sum / float64(period)

	k := 2.0 / float64(period+1)
	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*k + result[i-1]*(1-k)
	}
	return result
}

// RSI 相对强弱指数
func RSI(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	if len(prices) < period+1 {
		return result
	}
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

// MACDResult 存储 MACD 三条线
type MACDResult struct {
	DIF  []float64 `json:"dif"`
	DEA  []float64 `json:"dea"`
	MACD []float64 `json:"macd"`
}

// MACD 计算 MACD 指标，修复 DEA 零值段偏移问题
func MACD(prices []float64, fast, slow, signal int) MACDResult {
	empty := MACDResult{
		DIF:  make([]float64, len(prices)),
		DEA:  make([]float64, len(prices)),
		MACD: make([]float64, len(prices)),
	}
	if len(prices) < slow+signal {
		return empty
	}

	emaFast := EMA(prices, fast)
	emaSlow := EMA(prices, slow)

	dif := make([]float64, len(prices))
	for i := slow - 1; i < len(prices); i++ {
		dif[i] = emaFast[i] - emaSlow[i]
	}

	// 只对有效段（slow-1 之后）做 EMA，避免零值段拉偏 DEA
	validDif := dif[slow-1:]
	validDea := EMA(validDif, signal)
	dea := make([]float64, len(prices))
	copy(dea[slow-1:], validDea)

	macd := make([]float64, len(prices))
	for i := range prices {
		macd[i] = (dif[i] - dea[i]) * 2
	}

	return MACDResult{DIF: dif, DEA: dea, MACD: macd}
}
