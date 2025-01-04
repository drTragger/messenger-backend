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
