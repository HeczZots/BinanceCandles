package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	websocket "github.com/gorilla/websocket"
)

type Candle struct {
	CandleE string `json:"e"` // Event type
	E       int64  `json:"E"` // Event time
	S       string `json:"s"` // Symbol
	K       K      `json:"k"`
}

type K struct {
	KT int64  `json:"t"` // Kline start time
	T  int64  `json:"T"` // Kline close time
	S  string `json:"s"` //symbol
	I  string `json:"i"` //interval
	F  int64  `json:"f"` // First trade ID
	L  int64  `json:"L"` // Last trade ID
	O  string `json:"o"` // Open price
	C  string `json:"c"` // Close price
	H  string `json:"h"` // High price
	KL string `json:"l"` // Low price
	KV string `json:"v"` // Base asset volume
	N  int64  `json:"n"` // Number of trades
	X  bool   `json:"x"` // Is this kline closed?
	KQ string `json:"q"` // Quote asset volume
	V  string `json:"V"` // Taker buy base asset volume
	Q  string `json:"Q"` // Taker buy quote asset volume
	B  string `json:"B"` // Ignore
}

func main() {

	ch := make(chan Candle)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	if len(os.Args) < 3 {
		fmt.Println("Run follows: ", os.Args[0], "<symbol> <timeframe>")
		return
	}
	symb := os.Args[1]
	subscribeMessage := []byte(fmt.Sprintf(`{"method": "SUBSCRIBE","params": ["%s@kline_1s"],"id": 1}`, symb))
	candle := os.Args[2]
	go func() {
		<-interrupt
		time.Sleep(1 * time.Second)
		fmt.Printf("Interrupt signal received. Performing main exit...")
		os.Exit(0)
	}()
	go GetCandleStruct(subscribeMessage, ch)
	go GetCoustumCandle(candle)
	for data := range ch {
		fmt.Println(data)
	}
}
func GetCandleStruct(subscribeMessage []byte, candle chan<- Candle) {
	wss := "wss://stream.binance.com:9443/ws"
	for {
		conn, _, err := websocket.DefaultDialer.Dial(wss, nil)
		if err != nil {
			time.Sleep(12 * time.Second)
			log.Fatal("WebSocket connection failed:", err)
			continue
		}
		// defer conn.Close()
		fmt.Printf("connection to the: %v  - done\n", wss)
		// Create a channel to receive interrupt signals
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		closeComplete := make(chan struct{})
		ticker := time.NewTicker(500 * time.Millisecond)
		go func() {
			<-interrupt
			fmt.Printf("Interrupt signal received. Performing cleanup...")
			conn.Close()
			ticker.Stop()
			fmt.Println("Ticker stopped")
			close(closeComplete)
		}()
		err = conn.WriteMessage(websocket.TextMessage, subscribeMessage)
		if err != nil {
			log.Fatal("Failed to send subscribe message:", err)
			conn.Close()
			continue
		}
		go func() {
			for {
				select {
				case <-closeComplete:
					return
				case t := <-ticker.C:
					fmt.Println("Tick at", t)
				}
				_, message, err := conn.ReadMessage()
				if err != nil {
					if interrupt != nil {
						conn.Close()
						return
					}
					fmt.Println("Failed to read message from WebSocket:", err)
					fmt.Printf("Trying to reconnect. Performing cleanup...")
					conn.Close()
					time.Sleep(12 * time.Second)
					break
				}
				var data Candle
				err = json.Unmarshal(message, &data)
				if err != nil {
					fmt.Println("Error parsing JSON:", err)
					continue
				}
				candle <- data
			}
		}()
		select {}
	}

}
func GetCoustumCandle(candle string) {

}
