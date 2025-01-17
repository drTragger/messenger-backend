package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/requests"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	"github.com/drTragger/messenger-backend/internal/websocket"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ChatHandler struct {
	ChatRepo      *repository.ChatRepository
	UserRepo      *repository.UserRepository
	ClientManager *websocket.ClientManager
	Trans         *utils.Translator
}

func NewChatHandler(
	chatRepo *repository.ChatRepository,
	userRepo *repository.UserRepository,
	clientManager *websocket.ClientManager,
	trans *utils.Translator,
) *ChatHandler {
	return &ChatHandler{
		ChatRepo:      chatRepo,
		UserRepo:      userRepo,
		ClientManager: clientManager,
		Trans:         trans,
	}
}

func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload requests.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	user2, err := h.UserRepo.GetUserByID(payload.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
				"userId": h.Trans.Translate(r, "validation.exists", nil),
			})
			return
		}
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	user1ID := r.Context().Value("user_id").(uint)
	user1, err := h.UserRepo.GetUserByID(user1ID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	chat, err := h.ChatRepo.Create(user1.ID, user2.ID, nil)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	chat.User1 = user1
	chat.User2 = user2

	responses.SuccessResponse(w, http.StatusCreated, h.Trans.Translate(r, "success.chat.create", nil), chat)
}

func (h *ChatHandler) GetForUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userID := r.Context().Value("user_id").(uint)
	var err error

	// Extract and validate limit (default to const if not provided)
	limitStr := query.Get("limit")
	limit := repository.ChatsLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid limit")
			return
		}
	}

	// Extract and validate offset (default to const if not provided)
	offsetStr := query.Get("offset")
	offset := repository.ChatsOffset
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid offset.")
			return
		}
	}

	chats, err := h.ChatRepo.GetForUser(userID, limit, offset)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	onlineUsers := h.ClientManager.GetOnlineUsers()

	falsePtr := new(bool)
	*falsePtr = false

	for _, chat := range chats {
		if status, exists := onlineUsers[chat.User1ID]; exists {
			chat.User1.IsOnline = &status.IsOnline
		} else {
			chat.User1.IsOnline = falsePtr
		}

		if status, exists := onlineUsers[chat.User2ID]; exists {
			chat.User2.IsOnline = &status.IsOnline
		} else {
			chat.User2.IsOnline = falsePtr
		}
	}

	if len(chats) == 0 {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Chats not found.")
		return
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.chat.get_list", nil), chats)
}

func (h *ChatHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	chatIdStr := mux.Vars(r)["id"]
	chatId, err := strconv.Atoi(chatIdStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid limit")
		return
	}

	chat, err := h.ChatRepo.GetByID(uint(chatId))
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if chat == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Chat not found.")
		return
	}

	userID := r.Context().Value("user_id").(uint)

	if chat.User1ID != userID && chat.User2ID != userID {
		responses.ErrorResponse(w, http.StatusForbidden, h.Trans.Translate(r, "errors.forbidden", nil), "Forbidden")
		return
	}

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.chat.show", nil), chat)
}
