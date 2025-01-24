package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/requests"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/services"
	"github.com/drTragger/messenger-backend/internal/storage"
	"github.com/drTragger/messenger-backend/internal/utils"
	"github.com/drTragger/messenger-backend/internal/websocket"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

const (
	MaxMessageSize = 200 // MegaBytes
)

type MessageHandler struct {
	MsgService     *services.MessageService
	WsService      *services.WsService
	MsgRepo        *repository.MessageRepository
	UserRepo       *repository.UserRepository
	ChatRepo       *repository.ChatRepository
	AttachmentRepo *repository.AttachmentRepository
	Storage        storage.Storage
	Trans          *utils.Translator
}

func NewMessageHandler(
	msgService *services.MessageService,
	wsService *services.WsService,
	msgRepo *repository.MessageRepository,
	userRepo *repository.UserRepository,
	chatRepo *repository.ChatRepository,
	attachmentRepo *repository.AttachmentRepository,
	storage storage.Storage,
	trans *utils.Translator,
) *MessageHandler {
	return &MessageHandler{
		MsgService:     msgService,
		WsService:      wsService,
		MsgRepo:        msgRepo,
		UserRepo:       userRepo,
		ChatRepo:       chatRepo,
		AttachmentRepo: attachmentRepo,
		Storage:        storage,
		Trans:          trans,
	}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var payload requests.SendMessageRequest
	if err := r.ParseMultipartForm(MaxMessageSize << 20); err != nil {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Failed to parse form data")
		return
	}

	content := r.FormValue("content")
	payload.Content = &content

	recipientID, err := strconv.Atoi(r.FormValue("recipientId"))
	if err != nil || recipientID < 0 {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid recipient ID")
		return
	}
	payload.RecipientID = uint(recipientID)

	chatID, err := strconv.Atoi(mux.Vars(r)["chatId"])
	if err != nil || chatID < 0 {
		responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid chat ID")
		return
	}

	parentIDStr := r.FormValue("parentId")
	if parentIDStr != "0" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil || parentID < 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid parent ID")
			return
		}
		parentIDUint := uint(parentID)
		payload.ParentID = &parentIDUint
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), utils.FormatValidationError(r, err, h.Trans))
		return
	}

	senderID := r.Context().Value("user_id").(uint)

	sender, err := h.UserRepo.GetUserByID(senderID)
	if err != nil || sender == nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"senderId": h.Trans.Translate(r, "validation.exists", nil),
		})
		return
	}

	recipient, err := h.UserRepo.GetUserByID(payload.RecipientID)
	if err != nil || recipient == nil {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"recipientId": h.Trans.Translate(r, "validation.exists", nil),
		})
		return
	}

	chat, err := h.ChatRepo.GetByID(uint(chatID))
	if err != nil || chat == nil {
		chat, err = h.ChatRepo.Create(payload.RecipientID, senderID, nil)
		if err != nil {
			responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
			return
		}
	}

	if payload.ParentID != nil {
		parent, err := h.MsgRepo.GetById(*payload.ParentID)
		if err != nil || parent == nil {
			responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
				"parentId": h.Trans.Translate(r, "validation.exists", nil),
			})
			return
		}
	}

	message := &models.Message{
		SenderID:    senderID,
		RecipientID: payload.RecipientID,
		Content:     payload.Content,
		ChatID:      chat.ID,
		ParentID:    payload.ParentID,
	}
	message, err = h.MsgRepo.Create(message)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if err := h.MsgService.ProcessAttachments(r, message); err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	err = h.ChatRepo.UpdateLastMessage(chat.ID, message.ID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	go h.WsService.SendMessage(websocket.NewMessageEvent, payload.RecipientID, message)

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

	go h.WsService.SendMessage(websocket.EditMessageEvent, message.RecipientID, message)

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

	err = h.MsgService.DeleteAttachments(message.Attachments)
	if err != nil {
		log.Printf("Error deleting attachments: %s", err.Error())
	}

	if err := h.MsgRepo.Delete(uint(messageID)); err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	lastMessage, err := h.MsgRepo.GetLastMessageForChat(message.ChatID)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	lastMessageID := uint(0)
	if lastMessage != nil {
		lastMessageID = lastMessage.ID
	}
	if err := h.ChatRepo.UpdateLastMessage(message.ChatID, lastMessageID); err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	go h.WsService.SendMessage(websocket.DeleteMessageEvent, message.RecipientID, map[string]*models.Message{"deleted": message, "last": lastMessage})

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.delete", nil), nil)
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
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

	limitStr := query.Get("limit")
	limit := repository.MessagesLimit
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid limit")
			return
		}
	}

	offsetStr := query.Get("offset")
	offset := repository.MessagesOffset
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			responses.ErrorResponse(w, http.StatusBadRequest, h.Trans.Translate(r, "errors.input", nil), "Invalid offset.")
			return
		}
	}

	messages, err := h.MsgRepo.GetChatMessages(uint(chatID), limit, offset)
	if err != nil {
		responses.ErrorResponse(w, http.StatusInternalServerError, h.Trans.Translate(r, "errors.server", nil), err.Error())
		return
	}

	if len(messages) == 0 {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), "Messages not found.")
		return
	}

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

	go h.WsService.SendMessage(websocket.ReadMessageEvent, message.SenderID, message)

	responses.SuccessResponse(w, http.StatusOK, h.Trans.Translate(r, "success.message.read", nil), nil)
}

func (h *MessageHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	fileName := mux.Vars(r)["filename"]
	if fileName == "" {
		responses.ValidationResponse(w, h.Trans.Translate(r, "errors.validation", nil), map[string]string{
			"file": h.Trans.Translate(r, "validation.required", nil),
		})
		return
	}

	filePath, err := h.Storage.GetFile(storage.MessageAttachmentsDir, fileName)
	if err != nil {
		responses.ErrorResponse(w, http.StatusNotFound, h.Trans.Translate(r, "errors.not_found", nil), fmt.Sprintf("File not found: %v", err))
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	responses.ServeFileResponse(w, r, filePath)
}
