package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

func WsBlockInfo(blockHeightchan chan<- string) {

	wss := "wss://juno.kingnodes.com:443/websocket"
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

		// Define a channel to receive the result of the close operation
		closeComplete := make(chan struct{})
		go func() {
			// Block until an interrupt signal is received
			<-interrupt
			fmt.Printf("Interrupt signal received. Performing cleanup...")
			conn.Close()
			// Notify that the close operation is complete
			close(closeComplete)
		}()
		// Define the JSON-RPC message to subscribe to specific events
		// Send the subscribe message to the WebSocket server
		subscribeMessage := []byte(`{
		"jsonrpc": "2.0",
		"method": "subscribe",
		"params": ["tm.event = 'NewBlock'"],
		"id": 1
	}`)
		err = conn.WriteMessage(websocket.TextMessage, subscribeMessage)
		if err != nil {
			log.Fatal("Failed to send subscribe message:", err)
		}
		// Start a goroutine to receive messages from the WebSocket server
		go func() {
			for {
				select {
				case <-closeComplete:
					return
				default:
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
				var parsedResponse map[string]interface{}
				err = json.Unmarshal(message, &parsedResponse)
				if err != nil {
					fmt.Println("Error parsing JSON:", err)
					continue
				}

				result, ok := parsedResponse["result"].(map[string]interface{})
				if !ok {
					fmt.Println("Missing or invalid 'result' field in the JSON response.")
					continue
				}
				data, ok := result["data"].(map[string]interface{})
				if !ok {
					fmt.Printf("Waiting for data about block...\n")
					continue
				}

				value, ok := data["value"].(map[string]interface{})
				if !ok {
					fmt.Println("Missing or invalid 'value' field in the JSON response.")
					continue
				}
				block, ok := value["block"].(map[string]interface{})
				if !ok {
					fmt.Println("Missing or invalid 'block' field in the JSON response.")
					continue
				}
				blockData, ok := block["header"].(map[string]interface{})
				if !ok {
					fmt.Println("Missing or invalid 'blockData' field in the JSON response.")
					continue
				}
				blockHeight, ok := blockData["height"].(string)
				if !ok {
					fmt.Println("Missing or invalid 'txs' field in the JSON response.")
					continue
				}
				blockHeightchan <- blockHeight
			}
		}()
		select {}
	}
}
