# crypto-indicator

A backend service written in Go that fetches real-time K-line data from Binance
and calculates technical indicators (MA, RSI, MACD).

## Features

- Fetches K-line data from Binance public API (no API key required)
- Calculates MA (5, 20), RSI (14), and MACD (12, 26, 9)
- RESTful JSON API built with Go standard library

## Quick Start
```bash
git clone https://github.com/YOUR_USERNAME/crypto-indicator.git
cd crypto-indicator
go run main.go
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /ping` | Health check |
| `GET /api/kline?symbol=BTCUSDT&interval=1d` | Raw K-line data |
| `GET /api/indicator?symbol=BTCUSDT&interval=1d` | MA / RSI / MACD |

**Supported intervals:** `1m` `5m` `15m` `1h` `4h` `1d`

## Project Structure
```
crypto-indicator/
├── main.go              # HTTP server & routing
├── fetcher/
│   └── binance.go       # Binance API client
├── calculator/
│   └── indicator.go     # MA, EMA, RSI, MACD logic
└── model/
    └── kline.go         # Data structures
```

## Tech Stack

- **Language:** Go 1.21+
- **Data source:** Binance REST API
- **Dependencies:** Go standard library only