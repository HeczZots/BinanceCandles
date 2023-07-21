package models

type Candle struct {
	CandleE string `json:"e"` // Event type
	E       int64  `json:"E"` // Event time
	S       string `json:"s"` // Symbol
	K       K      `json:"k"`
}

type K struct {
	OT   int64  `json:"t"` // Kline start time
	CT   int64  `json:"T"` // Kline close time
	S    string `json:"s"` //symbol
	I    string `json:"i"` //interval
	F    int64  `json:"f"` // First trade ID
	L    int64  `json:"L"` // Last trade ID
	OP   string `json:"o"` // Open price
	CP   string `json:"c"` // Close price
	HP   string `json:"h"` // High price
	LP   string `json:"l"` // Low price
	Vol  string `json:"v"` // Base asset volume
	N    int64  `json:"n"` // Number of trades
	Done bool   `json:"x"` // Is this kline closed?
	KQ   string `json:"q"` // Quote asset volume
	V    string `json:"V"` // Taker buy base asset volume
	Q    string `json:"Q"` // Taker buy quote asset volume
	B    string `json:"B"` // Ignore
}

type MyCandle struct {
	OT   int64
	CT   int64
	OP   float64
	CP   float64
	HP   float64
	LP   float64
	Vol  float64
	Done bool
}
