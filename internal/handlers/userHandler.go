package handlers

import (
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	"github.com/drTragger/messenger-backend/internal/websocket"
	"net/http"
)

type UserHandler struct {
	UserRepo      *repository.UserRepository
	ClientManager *websocket.ClientManager
	Trans         *utils.Translator
}

func NewUserHandler(userRepo *repository.UserRepository, clientManager *websocket.ClientManager, trans *utils.Translator) *UserHandler {
	return &UserHandler{
		UserRepo:      userRepo,
		ClientManager: clientManager,
		Trans:         trans,
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	if query == "" {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"query": h.Trans.Translate(r, "validation.required", nil),
		})
		return
	}

	users, err := h.UserRepo.GetUsersBySearch(query)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, "errors.server", err.Error())
		return
	}

	if len(users) == 0 {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Users not found")
		return
	}

	onlineUsers := h.ClientManager.GetOnlineUsers()

	falsePtr := new(bool)
	*falsePtr = false

	for _, user := range users {
		if status, exists := onlineUsers[user.ID]; exists {
			user.IsOnline = &status.IsOnline
		} else {
			user.IsOnline = falsePtr
		}
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.user.search", nil), users)
}
