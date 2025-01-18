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
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type MessageHandler struct {
	MsgRepo       *repository.MessageRepository
	UserRepo      *repository.UserRepository
	ChatRepo      *repository.ChatRepository
	ClientManager *websocket.ClientManager
	Trans         *utils.Translator
}

func NewMessageHandler(
	msgRepo *repository.MessageRepository,
	userRepo *repository.UserRepository,
	chatRepo *repository.ChatRepository,
	clientManager *websocket.ClientManager,
	trans *utils.Translator,
) *MessageHandler {
	return &MessageHandler{
		MsgRepo:       msgRepo,
		UserRepo:      userRepo,
		ChatRepo:      chatRepo,
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

	chat, err := h.ChatRepo.GetByID(payload.ChatID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if chat == nil {
		chat, err = h.ChatRepo.Create(recipient.ID, sender.ID, nil)
		if err != nil {
			responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
			return
		}
	}

	message := &models.Message{
		SenderID:    r.Context().Value("user_id").(uint),
		RecipientID: payload.RecipientID,
		Content:     payload.Message,
		MessageType: models.TextMessage,
		ChatID:      chat.ID,
	}

	message, err = h.MsgRepo.Create(message)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	message.Sender = sender
	message.Recipient = recipient
	chat.LastMessageID = &message.ID
	message.Chat = chat

	err = h.ChatRepo.UpdateLastMessage(chat.ID, message.ID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	notification := websocket.NewNotification(websocket.NewMessageEvent, message)
	go h.ClientManager.SendMessage(payload.RecipientID, notification)

	responses.SuccessResponse(w, http.StatusCreated, h.Trans.Translate(r, "success.message.send", nil), message)
}

func (h *MessageHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	var payload requests.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	messageIDStr := mux.Vars(r)["messageId"]
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	message, err := h.MsgRepo.Edit(uint(messageID), payload.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), err.Error())
			return
		}
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	chat, err := h.ChatRepo.GetByID(message.ChatID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	message.Chat = chat

	notification := websocket.NewNotification(websocket.EditMessageEvent, message)
	go h.ClientManager.SendMessage(message.RecipientID, notification)

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.edit", nil), message)
}

func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageIDStr := mux.Vars(r)["messageId"]
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	message, err := h.MsgRepo.GetById(uint(messageID))
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	if message == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Message not found")
		return
	}

	err = h.MsgRepo.Delete(uint(messageID))
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	lastMessage, err := h.MsgRepo.GetLastMessageForChat(message.ChatID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	if lastMessage == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Message not found")
		return
	}

	err = h.ChatRepo.UpdateLastMessage(message.ChatID, lastMessage.ID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	notification := websocket.NewNotification(websocket.DeleteMessageEvent, map[string]*models.Message{"deleted": message, "last": lastMessage})
	go h.ClientManager.SendMessage(message.RecipientID, notification)

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.delete", nil), nil)
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Extract and validate sender ID
	chatIDStr := mux.Vars(r)["chatId"]
	if chatIDStr == "" {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"chatId": h.Trans.Translate(r, "validation.required", nil),
		})
		return
	}
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid sender_id")
		return
	}

	// Extract and validate limit (default to const if not provided)
	limitStr := query.Get("limit")
	limit := repository.MessagesLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid limit")
			return
		}
	}

	// Extract and validate offset (default to const if not provided)
	offsetStr := query.Get("offset")
	offset := repository.MessagesOffset
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid offset.")
			return
		}
	}

	// Fetch messages
	messages, err := h.MsgRepo.GetChatMessages(uint(chatID), limit, offset)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if len(messages) == 0 {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Messages not found.")
		return
	}

	// Respond with the list of messages
	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.get_list", nil), messages)
}

func (h *MessageHandler) MarkMessageRead(w http.ResponseWriter, r *http.Request) {
	messageIDStr := mux.Vars(r)["messageId"]
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), err.Error())
		return
	}

	currentUserID := r.Context().Value("user_id").(uint)

	message, err := h.MsgRepo.GetById(uint(messageID))
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if message == nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Message not found")
		return
	}

	if message.SenderID == currentUserID {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.message.sender_read", nil), "Sender is not allowed")
		return
	}

	readAt, err := h.MsgRepo.MarkAsRead(uint(messageID))
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}
	message.ReadAt = readAt

	notification := websocket.NewNotification(websocket.ReadMessageEvent, message)
	go h.ClientManager.SendMessage(message.SenderID, notification)

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.read", nil), nil)
}
