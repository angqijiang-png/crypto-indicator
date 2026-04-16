package calculator

import (
	"math"

	"crypto-indicator/model"
)

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

// BBResult 存储布林带三条线
type BBResult struct {
	Upper []float64 `json:"upper"`
	Mid   []float64 `json:"mid"`
	Lower []float64 `json:"lower"`
}

// BollingerBands 计算布林带（默认参数：period=20, mult=2.0）
func BollingerBands(prices []float64, period int, mult float64) BBResult {
	n := len(prices)
	res := BBResult{Upper: make([]float64, n), Mid: make([]float64, n), Lower: make([]float64, n)}
	if n < period {
		return res
	}
	mid := MA(prices, period)
	for i := period - 1; i < n; i++ {
		mean := mid[i]
		variance := 0.0
		for j := i - (period - 1); j <= i; j++ {
			d := prices[j] - mean
			variance += d * d
		}
		std := math.Sqrt(variance / float64(period))
		res.Upper[i] = mean + mult*std
		res.Mid[i] = mean
		res.Lower[i] = mean - mult*std
	}
	return res
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

// ── helpers ───────────────────────────────────────────────────────────────────

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// ── new structs ───────────────────────────────────────────────────────────────

// BollingerBand stores one candle's Bollinger Band values
type BollingerBand struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
	Width  float64 `json:"width"`
}

// ATRData stores True Range and ATR for one candle
type ATRData struct {
	TR  float64 `json:"tr"`
	ATR float64 `json:"atr"`
}

// KDJData stores K, D, J for one candle
type KDJData struct {
	K float64 `json:"k"`
	D float64 `json:"d"`
	J float64 `json:"j"`
}

// VolumeData stores volume indicator values for one candle
type VolumeData struct {
	Volume   float64 `json:"volume"`
	VolMA    float64 `json:"vol_ma"`
	VolRatio float64 `json:"vol_ratio"`
}

// SignalScore is the result of the composite signal scoring function
type SignalScore struct {
	Total     float64            `json:"total"`
	Verdict   string             `json:"verdict"`
	Breakdown map[string]float64 `json:"breakdown"`
}

// ── new functions ─────────────────────────────────────────────────────────────

// CalcBollingerBands computes Bollinger Bands from klines.
// Middle = SMA(period), Upper/Lower = Middle ± multiplier*stddev,
// Width = (Upper-Lower)/Middle. Zero-value for insufficient data.
func CalcBollingerBands(klines []model.Kline, period int, multiplier float64) []BollingerBand {
	n := len(klines)
	result := make([]BollingerBand, n)
	if n < period {
		return result
	}
	for i := period - 1; i < n; i++ {
		sum := 0.0
		for j := i - (period - 1); j <= i; j++ {
			sum += klines[j].Close
		}
		mean := sum / float64(period)
		variance := 0.0
		for j := i - (period - 1); j <= i; j++ {
			d := klines[j].Close - mean
			variance += d * d
		}
		std := math.Sqrt(variance / float64(period))
		upper := mean + multiplier*std
		lower := mean - multiplier*std
		width := 0.0
		if mean != 0 {
			width = (upper - lower) / mean
		}
		result[i] = BollingerBand{
			Upper:  round2(upper),
			Middle: round2(mean),
			Lower:  round2(lower),
			Width:  round2(width),
		}
	}
	return result
}

// CalcATR computes Average True Range using Wilder smoothing.
// First candle: TR = High-Low. Seed ATR = SMA of first period TRs.
// Subsequent: ATR = (prevATR*(period-1) + TR) / period.
func CalcATR(klines []model.Kline, period int) []ATRData {
	n := len(klines)
	result := make([]ATRData, n)
	if n == 0 {
		return result
	}
	trs := make([]float64, n)
	trs[0] = klines[0].High - klines[0].Low
	for i := 1; i < n; i++ {
		hl := klines[i].High - klines[i].Low
		hpc := math.Abs(klines[i].High - klines[i-1].Close)
		lpc := math.Abs(klines[i].Low - klines[i-1].Close)
		trs[i] = math.Max(hl, math.Max(hpc, lpc))
	}
	if n < period {
		for i := 0; i < n; i++ {
			result[i].TR = round2(trs[i])
		}
		return result
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += trs[i]
		result[i].TR = round2(trs[i])
	}
	atr := sum / float64(period)
	result[period-1].ATR = round2(atr)
	for i := period; i < n; i++ {
		result[i].TR = round2(trs[i])
		atr = (atr*float64(period-1) + trs[i]) / float64(period)
		result[i].ATR = round2(atr)
	}
	return result
}

// CalcKDJ computes KDJ stochastic indicator.
// RSV = (Close - LowestLow_N) / (HighestHigh_N - LowestLow_N) * 100
// K = 2/3*prevK + 1/3*RSV, D = 2/3*prevD + 1/3*K, J = 3K - 2D
// K and D start at 50.
func CalcKDJ(klines []model.Kline, period int) []KDJData {
	n := len(klines)
	result := make([]KDJData, n)
	if n == 0 || period <= 0 {
		return result
	}
	k, d := 50.0, 50.0
	for i := 0; i < n; i++ {
		start := i - period + 1
		if start < 0 {
			start = 0
		}
		lo, hi := klines[start].Low, klines[start].High
		for j := start + 1; j <= i; j++ {
			if klines[j].Low < lo {
				lo = klines[j].Low
			}
			if klines[j].High > hi {
				hi = klines[j].High
			}
		}
		rsv := 50.0
		if hi != lo {
			rsv = (klines[i].Close - lo) / (hi - lo) * 100
		}
		k = 2.0/3.0*k + 1.0/3.0*rsv
		d = 2.0/3.0*d + 1.0/3.0*k
		j := 3*k - 2*d
		result[i] = KDJData{K: round2(k), D: round2(d), J: round2(j)}
	}
	return result
}

// CalcVolumeMA computes Volume SMA and the ratio of current volume to that average.
func CalcVolumeMA(klines []model.Kline, period int) []VolumeData {
	n := len(klines)
	result := make([]VolumeData, n)
	for i := 0; i < n; i++ {
		result[i].Volume = round2(klines[i].Volume)
		if i >= period-1 {
			sum := 0.0
			for j := i - period + 1; j <= i; j++ {
				sum += klines[j].Volume
			}
			volMA := sum / float64(period)
			ratio := 0.0
			if volMA != 0 {
				ratio = klines[i].Volume / volMA
			}
			result[i].VolMA = round2(volMA)
			result[i].VolRatio = round2(ratio)
		}
	}
	return result
}

// CalcOBV computes On-Balance Volume.
// Price up → add volume, price down → subtract volume, unchanged → carry forward.
func CalcOBV(klines []model.Kline) []float64 {
	n := len(klines)
	result := make([]float64, n)
	if n == 0 {
		return result
	}
	result[0] = klines[0].Volume
	for i := 1; i < n; i++ {
		switch {
		case klines[i].Close > klines[i-1].Close:
			result[i] = result[i-1] + klines[i].Volume
		case klines[i].Close < klines[i-1].Close:
			result[i] = result[i-1] - klines[i].Volume
		default:
			result[i] = result[i-1]
		}
	}
	return result
}

// CalcSignalScore computes a 6-factor weighted composite signal score at position idx.
//
// Weights: MA trend 20%, RSI 20%, MACD histogram 20%, Bollinger 15%, KDJ 15%, Volume 10%.
// Total is in the range roughly -100 to +100.
// Verdict: STRONG_BUY(≥60), BUY(≥25), STRONG_SELL(≤-60), SELL(≤-25), else NEUTRAL.
func CalcSignalScore(
	closes []float64,
	ma5 []float64,
	ma20 []float64,
	rsi []float64,
	macdHist []float64,
	bb []BollingerBand,
	kdj []KDJData,
	atr []ATRData,
	volData []VolumeData,
	idx int,
) SignalScore {
	breakdown := make(map[string]float64)

	inBounds := func(slice int) bool { return idx >= 0 && idx < slice }
	hasPrev := func(slice int) bool { return idx > 0 && idx < slice }

	// ── MA Trend (weight 20%) ────────────────────────────────────────────────
	maScore := 0.0
	if inBounds(len(ma5)) && inBounds(len(ma20)) && ma5[idx] != 0 && ma20[idx] != 0 {
		if hasPrev(len(ma5)) && hasPrev(len(ma20)) && ma5[idx-1] != 0 && ma20[idx-1] != 0 {
			switch {
			case ma5[idx] > ma20[idx] && ma5[idx-1] <= ma20[idx-1]:
				maScore = 1.0 // golden cross
			case ma5[idx] < ma20[idx] && ma5[idx-1] >= ma20[idx-1]:
				maScore = -1.0 // death cross
			case ma5[idx] > ma20[idx]:
				maScore = 0.6
			default:
				maScore = -0.6
			}
		} else if ma5[idx] > ma20[idx] {
			maScore = 0.6
		} else if ma5[idx] < ma20[idx] {
			maScore = -0.6
		}
	}
	breakdown["ma"] = round2(maScore)

	// ── RSI (weight 20%) ─────────────────────────────────────────────────────
	rsiScore := 0.0
	if inBounds(len(rsi)) && rsi[idx] != 0 {
		v := rsi[idx]
		switch {
		case v <= 30:
			rsiScore = 0.8
		case v <= 40:
			rsiScore = 0.4
		case v >= 70:
			rsiScore = -0.8
		case v >= 60:
			rsiScore = -0.4
		}
	}
	breakdown["rsi"] = round2(rsiScore)

	// ── MACD Histogram (weight 20%) ──────────────────────────────────────────
	macdScore := 0.0
	if inBounds(len(macdHist)) && macdHist[idx] != 0 {
		h := macdHist[idx]
		growing := hasPrev(len(macdHist)) && macdHist[idx] > macdHist[idx-1]
		falling := hasPrev(len(macdHist)) && macdHist[idx] < macdHist[idx-1]
		switch {
		case h > 0 && growing:
			macdScore = 0.8
		case h > 0:
			macdScore = 0.6
		case h < 0 && falling:
			macdScore = -0.8
		default:
			macdScore = -0.6
		}
	}
	breakdown["macd"] = round2(macdScore)

	// ── Bollinger Bands (weight 15%) ─────────────────────────────────────────
	bbScore := 0.0
	if inBounds(len(bb)) && inBounds(len(closes)) && bb[idx].Middle != 0 {
		b := bb[idx]
		c := closes[idx]
		switch {
		case c <= b.Lower:
			bbScore = 0.8
		case c >= b.Upper:
			bbScore = -0.8
		case c < b.Middle:
			bbScore = 0.2
		default:
			bbScore = -0.2
		}
	}
	breakdown["bb"] = round2(bbScore)

	// ── KDJ (weight 15%) ─────────────────────────────────────────────────────
	kdjScore := 0.0
	if inBounds(len(kdj)) {
		cur := kdj[idx]
		switch {
		case cur.K <= 20 || cur.J <= 0:
			kdjScore = 0.8 // oversold
		case cur.K >= 80 || cur.J >= 100:
			kdjScore = -0.8 // overbought
		case cur.K <= 30:
			kdjScore = 0.4 // leaning low
		case cur.K >= 70:
			kdjScore = -0.4 // leaning high
		default:
			if hasPrev(len(kdj)) {
				prev := kdj[idx-1]
				switch {
				case cur.K > cur.D && prev.K <= prev.D:
					kdjScore = 0.6 // K crossed above D
				case cur.K < cur.D && prev.K >= prev.D:
					kdjScore = -0.6 // K crossed below D
				case cur.K > cur.D:
					kdjScore = 0.2 // K above D
				case cur.K < cur.D:
					kdjScore = -0.2 // K below D
				}
			}
		}
	}
	breakdown["kdj"] = round2(kdjScore)

	// ── Volume (weight 10%) ──────────────────────────────────────────────────
	volScore := 0.0
	if inBounds(len(volData)) && inBounds(len(closes)) && volData[idx].VolRatio > 0 {
		ratio := volData[idx].VolRatio
		priceUp := hasPrev(len(closes)) && closes[idx] > closes[idx-1]
		priceDown := hasPrev(len(closes)) && closes[idx] < closes[idx-1]
		switch {
		case ratio > 2.0 && priceUp:
			volScore = 1.0 // surge up
		case ratio > 1.5 && priceUp:
			volScore = 0.8 // strong up
		case ratio > 1.2 && priceUp:
			volScore = 0.4 // moderate up
		case ratio > 2.0 && priceDown:
			volScore = -1.0 // surge down
		case ratio > 1.5 && priceDown:
			volScore = -0.8 // strong down
		case ratio > 1.2 && priceDown:
			volScore = -0.4 // moderate down
		case ratio < 0.5:
			volScore = 0 // extreme low volume, no signal
		case ratio < 0.8 && priceUp:
			volScore = -0.2 // shrinking up, weakening
		case ratio < 0.8 && priceDown:
			volScore = 0.2 // shrinking down, weakening
		}
	}
	breakdown["volume"] = round2(volScore)

	// ── Weighted Total ───────────────────────────────────────────────────────
	total := round2((maScore*0.20 + rsiScore*0.20 + macdScore*0.20 +
		bbScore*0.15 + kdjScore*0.15 + volScore*0.10) * 100)

	verdict := "NEUTRAL"
	switch {
	case total >= 60:
		verdict = "STRONG_BUY"
	case total >= 25:
		verdict = "BUY"
	case total <= -60:
		verdict = "STRONG_SELL"
	case total <= -25:
		verdict = "SELL"
	}

	return SignalScore{Total: total, Verdict: verdict, Breakdown: breakdown}
}
