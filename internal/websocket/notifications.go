package websocket

import "time"

const (
	NewMessageEvent    = "newMessage"
	EditMessageEvent   = "editMessage"
	DeleteMessageEvent = "deleteMessage"
	StatusChangeEvent  = "statusChange"
)

type Notification struct {
	Event   string      `json:"event"`
	Message interface{} `json:"message"`
}

func NewNotification(event string, message interface{}) *Notification {
	return &Notification{
		Event:   event,
		Message: message,
	}
}

type StatusChange struct {
	Event    string    `json:"event"`
	UserID   uint      `json:"userId"`
	IsOnline bool      `json:"isOnline"`
	LastSeen time.Time `json:"lastSeen"`
}

func NewStatusChange(userID uint, isOnline bool, lastSeen time.Time) *StatusChange {
	return &StatusChange{
		Event:    StatusChangeEvent,
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: lastSeen,
	}
}
