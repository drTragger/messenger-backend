package websocket

import (
	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type ClientManager struct {
	Clients     map[uint]*websocket.Conn    // Map user ID to WebSocket connection
	OnlineUsers map[uint]*models.OnlineUser // Track online status
	mu          sync.RWMutex                // Mutex for thread-safe operations
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:     make(map[uint]*websocket.Conn),
		OnlineUsers: make(map[uint]*models.OnlineUser),
	}
}

func (cm *ClientManager) AddClient(userID uint, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Clients[userID] = conn
	cm.OnlineUsers[userID] = &models.OnlineUser{
		IsOnline: true,
		LastSeen: time.Now(),
	}
	log.Printf("User %d connected", userID)

	statusChange := NewStatusChange(userID, true, cm.OnlineUsers[userID].LastSeen)
	cm.broadcastStatusChange(statusChange)
}

func (cm *ClientManager) RemoveClient(userID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if conn, exists := cm.Clients[userID]; exists {
		conn.Close()
		delete(cm.Clients, userID)
		cm.OnlineUsers[userID] = &models.OnlineUser{
			IsOnline: false,
			LastSeen: time.Now(),
		}
		log.Printf("User %d disconnected", userID)

		statusChange := NewStatusChange(userID, false, cm.OnlineUsers[userID].LastSeen)
		cm.broadcastStatusChange(statusChange)
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

func (cm *ClientManager) broadcastStatusChange(change *StatusChange) {
	for uid, conn := range cm.Clients {
		if uid == change.UserID {
			continue
		}
		if err := conn.WriteJSON(change); err != nil {
			log.Printf("Error broadcasting status change to user %d: %s", uid, err)
		}
	}
}

func (cm *ClientManager) GetOnlineUsers() map[uint]*models.OnlineUser {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.OnlineUsers
}

func (cm *ClientManager) UserIsOnline(userID uint) *models.OnlineUser {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.OnlineUsers[userID]
}
