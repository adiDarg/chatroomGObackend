package go_packages

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type IncomingMessage struct {
	Type      string `json:"type"` // "fetch" or "send"
	Username  string `json:"Username,omitempty"`
	Room      string `json:"room,omitempty"`
	Value     string `json:"Value,omitempty"`
	Timestamp string `json:"TimeStamp,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[string][]*websocket.Conn)
var clientLock = sync.RWMutex{}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	// Listen for incoming Messages
	for {
		// Read Message from the client
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading Message:", err)
			break
		}
		fmt.Printf("Received: %s\\n\n", message)
		var incoming IncomingMessage
		err = json.Unmarshal(message, &incoming)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			continue
		}
		go handleMessages(conn, incoming)
	}
}
func sendMessage(incoming IncomingMessage) error {
	err := WriteMessage(Message{incoming.Username, incoming.Room, incoming.Value, incoming.Timestamp})
	if err != nil {
		return err
	}
	clientLock.RLock()
	for _, client := range clients[incoming.Room] {
		go func(c *websocket.Conn) {
			err := fetchMessages(incoming.Room, c)
			if err != nil {
				fmt.Println("Error sending Messages:", err)
			}
		}(client)
	}
	clientLock.RUnlock()
	return nil
}
func fetchMessages(room string, conn *websocket.Conn) error {
	messages, err := GetMessages(room)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		removeClientFromRoom(room, conn)
		return nil
	}
	return err
}
func getRooms(conn *websocket.Conn) error {
	rooms := GetRooms()
	jsonData, err := json.Marshal(rooms)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, jsonData)
}
func handleMessages(conn *websocket.Conn, incoming IncomingMessage) {
	switch incoming.Type {
	case "fetch":
		err := fetchMessages(incoming.Room, conn)
		if err != nil {
			fmt.Println("Error sending Messages:", err)
		}
	case "send":
		err := sendMessage(incoming)
		if err != nil {
			fmt.Println("Error Sending Message: " + err.Error())
		}
	case "getRooms":
		err := getRooms(conn)
		if err != nil {
			fmt.Println("Error getting Rooms:", err)
		}

	case "create":
		CreateChatRoom(incoming.Room)
		fmt.Println("Chat room created: ", incoming.Room)
	case "joinChat":
		err := joinRoom(incoming, conn)
		if err != nil {
			fmt.Println("Error Joining Chat:", err)
		}
	case "leaveChat":
		removeClientFromRoom(incoming.Room, conn)
	}
}
func containsConn(connections []*websocket.Conn, target *websocket.Conn) bool {
	for _, conn := range connections {
		if conn == target {
			return true
		}
	}
	return false
}
func joinRoom(incoming IncomingMessage, conn *websocket.Conn) error {
	if containsConn(clients[incoming.Room], conn) {
		return errors.New("user already in the chat")
	}
	clientLock.Lock()
	clients[incoming.Room] = append(clients[incoming.Room], conn)
	clientLock.Unlock()
	newMSG := IncomingMessage{
		Room:      incoming.Room,
		Type:      "send",
		Username:  "system",
		Value:     incoming.Username + " has joined the chat",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	go func() {
		err := sendMessage(newMSG)
		if err != nil {
			fmt.Println("Error sending X joined chat message: ", err)
		}
	}()
	return nil
}
func removeClientFromRoom(room string, conn *websocket.Conn) {
	clientLock.Lock()
	defer clientLock.Unlock()

	connections := clients[room]
	for i, c := range connections {
		if c == conn {
			clients[room] = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	if len(clients[room]) == 0 {
		delete(clients, room)
	}
}
