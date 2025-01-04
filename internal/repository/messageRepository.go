package repository

import (
	"database/sql"
	"github.com/drTragger/messenger-backend/internal/models"
)

type MessageRepository struct {
	DB *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		DB: db,
	}
}

func (mr *MessageRepository) CreateMessage(msg *models.Message) (*models.Message, error) {
	query := `
		INSERT INTO messages (sender_id, recipient_id, content, message_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := mr.DB.QueryRow(query, msg.SenderID, msg.RecipientID, msg.Content, msg.MessageType).
		Scan(&msg.ID, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
