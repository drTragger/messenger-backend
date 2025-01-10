package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/requests"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	"github.com/drTragger/messenger-backend/internal/websocket"
	"net/http"
	"strconv"
)

const (
	MessagesLimit  = 20
	MessagesOffset = 0
)

type MessageHandler struct {
	MsgRepo       *repository.MessageRepository
	UserRepo      *repository.UserRepository
	ClientManager *websocket.ClientManager
	Trans         *utils.Translator
}

func NewMessageHandler(
	msgRepo *repository.MessageRepository,
	userRepo *repository.UserRepository,
	clientManager *websocket.ClientManager,
	trans *utils.Translator,
) *MessageHandler {
	return &MessageHandler{
		MsgRepo:       msgRepo,
		UserRepo:      userRepo,
		ClientManager: clientManager,
		Trans:         trans,
	}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var payload requests.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	senderID := r.Context().Value("user_id").(uint)

	sender, err := h.UserRepo.GetUserByID(senderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
				"senderId": h.Trans.Translate(r, "validation.exists", nil),
			})
			return
		}
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	recipient, err := h.UserRepo.GetUserByID(payload.RecipientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
				"receiverId": h.Trans.Translate(r, "validation.exists", nil),
			})
			return
		}
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	message := &models.Message{
		SenderID:    r.Context().Value("user_id").(uint),
		RecipientID: payload.RecipientID,
		Content:     payload.Message,
		MessageType: models.TextMessage,
	}

	message, err = h.MsgRepo.CreateMessage(message)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	message.Sender = sender
	message.Recipient = recipient

	notification := websocket.NewNotification(websocket.NewMessageEvent, message)
	go h.ClientManager.SendMessage(payload.RecipientID, notification)

	responses.SuccessResponse(w, http.StatusCreated, h.Trans.Translate(r, "success.message.send", nil), message)
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Extract and validate sender ID
	senderIDStr := query.Get("senderId")
	if senderIDStr == "" {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"senderId": h.Trans.Translate(r, "validation.required", nil),
		})
		return
	}
	senderID, err := strconv.Atoi(senderIDStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid sender_id")
		return
	}

	// Extract and validate limit (default to 20 if not provided)
	limitStr := query.Get("limit")
	limit := MessagesLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid limit")
			return
		}
	}

	// Extract and validate offset (default to 0 if not provided)
	offsetStr := query.Get("offset")
	offset := MessagesOffset
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid offset.")
			return
		}
	}

	// Extract receiver ID from context
	receiverID, ok := r.Context().Value("user_id").(uint)
	if !ok {
		responses.ErrorResponse(w, http.StatusUnauthorized, h.Trans.Translate(r, "errors.unauthorized", nil), "User not authenticated.")
		return
	}

	// Fetch messages
	messages, err := h.MsgRepo.GetUserMessages(uint(senderID), receiverID, limit, offset)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if len(messages) == 0 {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.messages.not_found", nil), "Messages not found.")
		return
	}

	// Respond with the list of messages
	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.get_list", nil), messages)
}
