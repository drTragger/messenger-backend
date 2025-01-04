package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type ClientManager struct {
	Clients map[uint]*websocket.Conn // Map user ID to WebSocket connection
	mu      sync.RWMutex             // Mutex for thread-safe operations
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients: make(map[uint]*websocket.Conn),
	}
}

func (cm *ClientManager) AddClient(userID uint, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Clients[userID] = conn
	log.Printf("User %d connected", userID)
}

func (cm *ClientManager) RemoveClient(userID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if conn, exists := cm.Clients[userID]; exists {
		conn.Close()
		delete(cm.Clients, userID)
		log.Printf("User %d disconnected", userID)
	}
}

func (cm *ClientManager) SendMessage(userID uint, message *Notification) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, exists := cm.Clients[userID]
	if !exists {
		log.Printf("User %d is not connected", userID)
		return
	}
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Error sending message to user %d: %s", userID, err)
	}
}
