package models

import "time"

type Chat struct {
	ID            uint      `json:"id"`
	User1ID       uint      `json:"user1Id"`
	User2ID       uint      `json:"user2Id"`
	LastMessageID *uint     `json:"lastMessageId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	User1       *User    `json:"user1"`
	User2       *User    `json:"user2"`
	LastMessage *Message `json:"lastMessage"`
}
