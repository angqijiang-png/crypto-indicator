# crypto-indicator

> Real-time cryptocurrency technical analysis engine — Go backend + Binance API + Bloomberg-style dashboard

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat&logo=go)
![Binance](https://img.shields.io/badge/Data-Binance%20API-F0B90B?style=flat)
![Zero Dependencies](https://img.shields.io/badge/Go%20deps-zero-brightgreen?style=flat)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

## Why I built this

I spent 2 years trading A-shares with real money. Every time I wanted to check MA, RSI, or MACD, I had to rely on tools I didn't fully understand or trust. So I built my own — from scratch, in Go — to understand exactly how the math works and verify that the numbers are correct.

This project is not a tutorial clone. Every algorithm (MA, EMA, RSI, MACD) is implemented by hand from the mathematical definition, then validated against known values.

---

## Features

- Fetches real-time OHLCV data from Binance REST API (`/api/v3/klines`)
- Computes **MA(5/20), EMA, RSI(14), MACD(12,26,9)** — all implemented from scratch in pure Go, no libraries
- MACD DEA zero-bias fix: EMA is computed only over the valid DIF segment to avoid seed distortion
- RESTful JSON API with CORS support
- Bloomberg Terminal-style dark dashboard: candlestick chart + RSI panel + MACD panel
- Composite signal scoring: each indicator votes BULL/BEAR/NEUTRAL → aggregated into a final market signal (0–6 score)
- Auto-refresh with configurable countdown timer
- `PORT` env variable support for deployment flexibility
- Zero external Go dependencies — standard library only

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.21, `net/http` |
| Data Source | Binance REST API (`/api/v3/klines`) |
| Charts | Lightweight Charts (TradingView), Chart.js |
| Frontend | Vanilla HTML / CSS / JS |

---

## Project Structure

```
crypto-indicator/
├── main.go                 # HTTP server, routing, CORS middleware
├── calculator/
│   └── calculator.go       # MA, EMA, RSI, MACD — pure math, no libs
├── fetcher/
│   └── fetcher.go          # Binance API client + response parsing
├── model/
│   └── kline.go            # Kline data structure
├── static/
│   └── index.html          # Web dashboard (single file)
└── go.mod
```

---

## API

| Endpoint | Description |
|---|---|
| `GET /ping` | Health check → `{"code":200,"message":"ok"}` |
| `GET /api/kline?symbol=BTCUSDT&interval=1d&limit=200` | Raw K-line data |
| `GET /api/indicator?symbol=BTCUSDT&interval=1d&limit=200` | K-lines + all computed indicators |

**Response shape (`/api/indicator`):**
```json
{
  "symbol": "BTCUSDT",
  "interval": "1d",
  "klines": [...],
  "ma5":   [0, 0, 0, 0, 42000.5, ...],
  "ma20":  [...],
  "rsi14": [...],
  "macd":  { "dif": [...], "dea": [...], "macd": [...] }
}
```

**Supported symbols:** BTCUSDT · ETHUSDT · SOLUSDT · BNBUSDT · XRPUSDT · DOGEUSDT · ADAUSDT · AVAXUSDT

**Supported intervals:** `1h` · `4h` · `1d` · `1w`

---

## Quick Start

```bash
git clone https://github.com/angqijiang-png/crypto-indicator.git
cd crypto-indicator
go run main.go
```

Open `http://localhost:8080` in your browser.

**Custom port:**
```bash
PORT=9000 go run main.go
```

---

## Dashboard

- Candlestick chart with MA5 / MA20 overlay
- RSI(14) panel — overbought (70) / oversold (30) threshold lines
- MACD(12,26,9) panel — DIF / DEA / histogram
- Side panel: live indicator values + per-indicator signal badges
- Composite signal: all indicator votes → final BULL / BEAR / NEUTRAL

---

## Roadmap

- [ ] Bollinger Bands (20, 2σ)
- [ ] WebSocket real-time push (replace polling)
- [ ] Price alert notifications (RSI cross 70/30)
- [ ] Custom symbol input (any Binance pair)
- [ ] Swagger / OpenAPI docs for the REST API

---

## Author

**Angqi Jiang (蒋昂奇)**
- GitHub: [@angqijiang-png](https://github.com/angqijiang-png)
- Telegram: [@angqi_web3](https://t.me/angqi_web3)
- X: [@angsevenJIANG](https://x.com/angsevenJIANG)
