package main

import (
	"TT/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
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
	go GetCandleStruct(subscribeMessage, ch)
	tf = tf * 250 * 4
	var wg sync.WaitGroup
	wg.Add(1)
	go GetCustomCandle(tf, ch, &wg)
	wg.Wait()
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
				if data.K.OT != 0 && data.K.CT != 0 {
					candle <- data
				}
			}
		}()
		select {}
	}
}
func GetCustomCandle(timeframe int, DefaultCandle <-chan models.Candle, wg *sync.WaitGroup) {
	defer wg.Done()
	customArr := make([]models.MyCandle, 1)
	var customCandle models.MyCandle
	for data := range DefaultCandle {
		intOP, err := strconv.ParseFloat(data.K.OP, 64)
		if err != nil {
			fmt.Printf("%v", err)
			continue
		}
		intCP, err := strconv.ParseFloat(data.K.CP, 64)
		if err != nil {
			fmt.Printf("CP")
			continue
		}
		intHP, err := strconv.ParseFloat(data.K.HP, 64)
		if err != nil {
			fmt.Printf("HP")
			continue
		}
		intLP, err := strconv.ParseFloat(data.K.LP, 64)
		if err != nil {
			fmt.Printf("LP")
			continue
		}
		intVol, err := strconv.ParseFloat(data.K.Vol, 64)
		if err != nil {
			fmt.Printf("vol")
			continue
		}
		if customArr[len(customArr)-1].OT == 0 {
			customArr[len(customArr)-1].OT = data.K.OT
			customArr[len(customArr)-1].OP = intOP
			customArr[len(customArr)-1].CT = data.K.OT + 1000
		}
		data.K.CT += 1
		customArr[len(customArr)-1].CP = intCP
		customArr[len(customArr)-1].CT += 1000
		fmt.Printf("\n CT OT dflt: %v : %v\n", data.K.CT, data.K.OT)
		fmt.Printf("\nCT OT : %v : %v\n", customArr[len(customArr)-1].CT, customArr[len(customArr)-1].OT)
		if intHP > customArr[len(customArr)-1].HP {
			customArr[len(customArr)-1].HP = intHP
		}
		if intLP < customArr[len(customArr)-1].LP {
			customArr[len(customArr)-1].LP = intLP
		}
		customArr[len(customArr)-1].Vol += intVol
		if int64(timeframe) == customArr[len(customArr)-1].CT-customArr[len(customArr)-1].OT {
			customArr[len(customArr)-1].Done = true
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
			if len(customArr) != 0 {
				fmt.Printf("%v\n", customArr[len(customArr)-1])
			}
			customArr = append(customArr, customCandle)
		}
	}
}
