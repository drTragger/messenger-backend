package handlers

import (
	"errors"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	ws "github.com/drTragger/messenger-backend/internal/websocket"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	ClientManager *ws.ClientManager
	TokenRepo     *repository.TokenRepository
	Translator    *utils.Translator
	Secret        string
}

func NewWebSocketHandler(clientManager *ws.ClientManager, tokenRepo *repository.TokenRepository, translator *utils.Translator, secret string) *WebSocketHandler {
	return &WebSocketHandler{
		ClientManager: clientManager,
		TokenRepo:     tokenRepo,
		Translator:    translator,
		Secret:        secret,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user
	userID, err := h.authenticate(r)
	if err != nil {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Translator.Translate(r, "errors.unauthorized", nil), err.Error())
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	// Add client to the client manager
	h.ClientManager.AddClient(userID, conn)
	defer h.ClientManager.RemoveClient(userID)

	// Handle incoming WebSocket messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WebSocketHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	onlineUsers := h.ClientManager.GetOnlineUsers()
	response := map[string][]uint{}

	for userID, onlineUser := range onlineUsers {
		if onlineUser.IsOnline {
			response["onlineUsers"] = append(response["onlineUsers"], userID)
		}
	}

	responses.SuccessResponse(w, http.StatusOK, h.Translator.Translate(r, "success.user.get_online_list", nil), response)
}

func (h *WebSocketHandler) GetUserIsOnline(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Translator.Translate(r, "errors.input", nil), err.Error())
		return
	}

	onlineUser := h.ClientManager.UserIsOnline(uint(userID))

	responses.SuccessResponse(w, http.StatusOK, h.Translator.Translate(r, "success.user.get_online_list", nil), onlineUser)
}

func (h *WebSocketHandler) authenticate(r *http.Request) (uint, error) {
	// Get the token from the query parameters
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		return 0, errors.New("token not provided")
	}

	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.Secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Extract user ID from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id not found in token claims")
	}

	// Validate the token in the token repository (e.g., Redis)
	valid, err := h.TokenRepo.IsTokenValid(r.Context(), tokenString, uint(userID))
	if err != nil || !valid {
		return 0, errors.New("token is invalid or expired")
	}

	return uint(userID), nil
}
