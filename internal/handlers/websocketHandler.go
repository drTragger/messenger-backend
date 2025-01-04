package handlers

import (
	ws "github.com/drTragger/messenger-backend/internal/websocket"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for simplicity, adjust for production
		return true
	},
}

type WebSocketHandler struct {
	ClientManager *ws.ClientManager
}

func NewWebSocketHandler(clientManager *ws.ClientManager) *WebSocketHandler {
	return &WebSocketHandler{ClientManager: clientManager}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	userID := r.Context().Value("user_id").(uint) // Extract user ID from context
	h.ClientManager.AddClient(userID, conn)

	defer h.ClientManager.RemoveClient(userID)
	for {
		// Keep the connection alive or handle pings/pongs if necessary
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
