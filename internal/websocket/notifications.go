package websocket

import "time"

const (
	NewMessageEvent    = EventType("newMessage")
	EditMessageEvent   = EventType("editMessage")
	DeleteMessageEvent = EventType("deleteMessage")
	ReadMessageEvent   = EventType("readMessage")
	StatusChangeEvent  = EventType("statusChange")
)

type EventType string

type Notification struct {
	Event   EventType   `json:"event"`
	Message interface{} `json:"message"`
}

func NewNotification(event EventType, message interface{}) *Notification {
	return &Notification{
		Event:   event,
		Message: message,
	}
}

type StatusChange struct {
	Event    EventType `json:"event"`
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
