package websocket

const (
	NewMessageEvent = "newMessage"
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
