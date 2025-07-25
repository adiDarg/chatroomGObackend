package main

import (
	gopackages "chatAppWebsocketBackend/go-packages"
	"fmt"
	"net/http"
	"time"
)

const minutesBetweenBackups int = 1

func main() {
	go backup(minutesBetweenBackups)
	http.HandleFunc("/ws", gopackages.WsHandler)
	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
func backup(sleepMinutes int) {
	var fails int = 0
	for {
		fmt.Println("Backing up...")
		err := gopackages.SaveChats()
		if err != nil {
			fmt.Println("Error saving chats:", err)
			fails++
			time.Sleep(time.Second)
			if fails >= 10 {
				fails = 0
				fmt.Println("Failed backup")
				time.Sleep(time.Duration(sleepMinutes) * time.Minute)
			}
			continue
		}
		fmt.Println("Backing up complete!")
		time.Sleep(time.Duration(sleepMinutes) * time.Minute)
	}
}
