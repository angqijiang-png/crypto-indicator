# crypto-indicator

> Real-time cryptocurrency technical analysis engine built with Go + Binance API

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat&logo=go)
![Binance](https://img.shields.io/badge/Data-Binance%20API-F0B90B?style=flat)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

## Overview

A lightweight backend service that fetches live K-line data from Binance and computes key technical indicators in real time. Comes with a Bloomberg Terminal-style web dashboard for visualization.

**Live indicators:** MA · EMA · RSI · MACD

---

## Features

- Fetches real-time OHLCV data from Binance REST API
- Calculates MA(5/20), RSI(14), MACD(12,26,9) from scratch in pure Go
- RESTful JSON API with CORS support
- Bloomberg Terminal-style dark dashboard (candlestick + indicator charts)
- Composite signal scoring: aggregates all indicators into a single BULL/BEAR/NEUTRAL signal
- Zero external Go dependencies — standard library only

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.21, `net/http` |
| Data Source | Binance REST API (`/api/v3/klines`) |
| Charts | Lightweight Charts (TradingView), Chart.js |
| Frontend | Vanilla HTML/CSS/JS |

---

## Project Structure

```
crypto-indicator/
├── main.go                 # HTTP server, routing, middleware
├── calculator/
│   └── calculator.go       # MA, EMA, RSI, MACD algorithms
├── fetcher/
│   └── fetcher.go          # Binance API client
├── model/
│   └── kline.go            # Kline data structure
├── static/
│   └── index.html          # Web dashboard
└── go.mod
```

---

## API

| Endpoint | Description |
|---|---|
| `GET /ping` | Health check |
| `GET /api/kline?symbol=BTCUSDT&interval=1d&limit=200` | Raw K-line data |
| `GET /api/indicator?symbol=BTCUSDT&interval=1d&limit=200` | K-lines + all indicators |

**Supported symbols:** BTCUSDT, ETHUSDT, SOLUSDT, BNBUSDT, XRPUSDT, DOGEUSDT, ADAUSDT, AVAXUSDT

**Supported intervals:** `1h` `4h` `1d` `1w`

---

## Quick Start

```bash
git clone https://github.com/angqijiang-png/crypto-indicator.git
cd crypto-indicator
go run main.go
```

Open `http://localhost:8080` in your browser.

---

## Dashboard

- Candlestick chart with MA5 / MA20 overlay
- RSI(14) panel with overbought / oversold zones
- MACD(12,26,9) panel with DIF / DEA / histogram
- Real-time indicator signals (BULL / BEAR / NEUTRAL per indicator)
- Composite signal: votes across all 6 indicators → final market signal

---

## Author

**Angqi Jiang**
- GitHub: [@angqijiang-png](https://github.com/angqijiang-png)
- Telegram: [@angqi_web3](https://t.me/angqi_web3)
- X: [@angsevenJIANG](https://x.com/angsevenJIANG)
