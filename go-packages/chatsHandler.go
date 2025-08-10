package go_packages

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Message struct {
	Username  string
	ChatRoom  string
	Value     string
	TimeStamp string
}
type chatRoom struct {
	Name     string
	Messages []Message
}

type chatManager struct {
	chatRooms map[string]*chatRoom
}

var (
	instance   *chatManager
	once       sync.Once
	storageDir = "chat-storage"
)

func init() {
	once.Do(func() {
		chatRooms, err := loadChats()
		if err != nil {
			fmt.Println(err)
		}
		instance = &chatManager{chatRooms: chatRooms}
	})
}

var lock sync.RWMutex = sync.RWMutex{}

func loadChats() (map[string]*chatRoom, error) {
	chatRooms := make(map[string]*chatRoom)
	files, err := os.ReadDir(storageDir)
	if err != nil {
		fmt.Println(err)
		instance = &chatManager{chatRooms: chatRooms}
	}
	for _, file := range files {
		openFile, err := os.Open(storageDir + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}
		decoder := json.NewDecoder(openFile)
		var chatRoom chatRoom
		err = decoder.Decode(&chatRoom)
		if err != nil {
			fmt.Println(err)
			continue
		}
		chatRooms[chatRoom.Name] = &chatRoom
		err = openFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
	return chatRooms, err
}
func SaveChats() error {
	chatRooms := instance.chatRooms
	lock.RLock()
	for key, value := range chatRooms {
		path := filepath.Join(storageDir, key+".json")
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		err = os.WriteFile(path, jsonData, 0644)
		if err != nil {
			return err
		}
	}
	lock.RUnlock()
	return nil
}
func CreateChatRoom(name string) {
	lock.Lock()
	instance.chatRooms[name] = &chatRoom{Name: name}
	lock.Unlock()
}
func GetRooms() []string {
	lock.RLock()
	var rooms = make([]string, len(instance.chatRooms))
	for room := range instance.chatRooms {
		rooms = append(rooms, room)
	}
	lock.RUnlock()
	return rooms
}
func WriteMessage(message Message) error {
	chatRoom, ok := instance.chatRooms[message.ChatRoom]
	if !ok {
		return errors.New("chat room " + message.ChatRoom + " does not exist")
	}
	lock.Lock()
	chatRoom.Messages = append(chatRoom.Messages, message)
	lock.Unlock()
	return nil
}
func GetMessages(room string) ([]Message, error) {
	lock.RLock()
	cRoom, ok := instance.chatRooms[room]
	lock.RUnlock()
	if !ok {
		return make([]Message, 0), errors.New("chat room " + room + " does not exist")
	}
	return cRoom.Messages, nil
}
