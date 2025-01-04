package models

import "time"

type Message struct {
	ID          uint        `json:"id"`
	SenderID    uint        `json:"senderId"`
	RecipientID uint        `json:"recipientId"`
	Content     string      `json:"content"`
	ReadAt      *time.Time  `json:"readAt"`
	MessageType MessageType `json:"messageType"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`

	Sender    *User `json:"sender"`
	Recipient *User `json:"recipient"`
}

type MessageType string

const (
	TextMessage   MessageType = "text"
	ImageMessage  MessageType = "image"
	VideoMessage  MessageType = "video"
	FileMessage   MessageType = "file"
	SystemMessage MessageType = "system"
)
