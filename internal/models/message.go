package models

import "time"

type Message struct {
	ID          uint       `json:"id"`
	SenderID    uint       `json:"senderId"`
	RecipientID uint       `json:"recipientId"`
	Content     *string    `json:"content"`
	ReadAt      *time.Time `json:"readAt"`
	ChatID      uint       `json:"chatId"`
	ParentID    *uint      `json:"parentId"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	Sender      *User         `json:"sender,omitempty"`
	Recipient   *User         `json:"recipient,omitempty"`
	Chat        *Chat         `json:"chats,omitempty"`
	Parent      *Message      `json:"parent,omitempty"`
	Attachments []*Attachment `json:"attachments"`
}
