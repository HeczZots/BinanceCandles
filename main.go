package main

import (
	"TT/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	websocket "github.com/gorilla/websocket"
)

func main() {

	ch := make(chan models.Candle)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	if len(os.Args) < 3 {
		fmt.Println("Run follows: ", os.Args[0], "<symbol> <timeframe>")
		return
	}
	//choice symb in lowercase
	symb := os.Args[1]
	subscribeMessage := []byte(fmt.Sprintf(`{"method": "SUBSCRIBE","params": ["%s@kline_1s"],"id": 1}`, symb))
	//choice tf in seconds
	timeframe := os.Args[2]
	tf, _ := strconv.Atoi(timeframe)
	go func() {
		<-interrupt
		time.Sleep(1 * time.Second)
		fmt.Printf("Interrupt signal received. Performing main exit...")
		os.Exit(0)
	}()
	var customCandle models.MyCandle
	go GetCandleStruct(subscribeMessage, ch)
	tf = tf * 250 * 4
	go GetCustomCandle(tf, ch, customCandle)
	for data := range ch {
		fmt.Println(data)
	}
}
func GetCandleStruct(subscribeMessage []byte, candle chan<- models.Candle) {
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
				var data models.Candle
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
func GetCustomCandle(timeframe int, DefaultCandle <-chan models.Candle, customCandle models.MyCandle) {
	customArr := make([]models.MyCandle, 0)
	for data := range DefaultCandle {
		intOP, _ := strconv.Atoi(data.K.OP)
		intCP, _ := strconv.Atoi(data.K.CP)
		intHP, _ := strconv.Atoi(data.K.HP)
		intLP, _ := strconv.Atoi(data.K.LP)
		intVol, _ := strconv.Atoi(data.K.Vol)
		data.K.CT += 1
		customCandle.CP = intCP
		customCandle.CT += int(data.K.CT) - int(data.K.OT)
		if !customCandle.Done && customCandle.CT-customCandle.OT == 999 {
			customCandle.OT = int(data.K.OT)
			customCandle.OP = intOP
		}
		if intHP > customCandle.HP {
			customCandle.HP = intHP
		}
		if intLP < customCandle.LP {
			customCandle.LP = intLP
		}
		customCandle.Vol += intVol
		if timeframe == customCandle.CT-customCandle.OT {
			customCandle.Done = true
			customArr = append(customArr, customCandle)
			customCandle = models.MyCandle{
				OT:   0,
				CT:   0,
				OP:   0,
				CP:   0,
				HP:   0,
				LP:   0,
				Vol:  0,
				Done: false,
			}
			customArr = append(customArr, customCandle)
		}
		fmt.Printf("%v", customArr[len(customArr)-1])
	}
}
