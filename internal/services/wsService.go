package services

import "github.com/drTragger/messenger-backend/internal/websocket"

type WsService struct {
	ClientManager *websocket.ClientManager
}

func NewWsService(clientManager *websocket.ClientManager) *WsService {
	return &WsService{
		ClientManager: clientManager,
	}
}

func (s *WsService) SendMessage(event websocket.EventType, recipientID uint, message interface{}) {
	notification := websocket.NewNotification(event, message)
	s.ClientManager.SendMessage(recipientID, notification)
}
