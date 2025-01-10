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

func (mr *MessageRepository) GetUserMessages(senderID uint, recipientID uint, limit int, offset int) ([]*models.Message, error) {
	query := `
		SELECT 
			m.id, 
			m.sender_id, 
			m.recipient_id, 
			m.content, 
			m.read_at, 
			m.message_type, 
			m.created_at, 
			m.updated_at,
			u.id AS user_id,
			u.username AS user_username,
			u.phone AS user_phone
		FROM messages AS m
		JOIN users AS u ON m.sender_id = u.id
		WHERE m.sender_id = $1 AND m.recipient_id = $2
		ORDER BY m.created_at
		LIMIT $3 OFFSET $4
	`

	// Execute the query
	rows, err := mr.DB.Query(query, senderID, recipientID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize a slice to hold the messages
	messages := make([]*models.Message, 0)

	// Iterate through the rows and scan data into the struct
	for rows.Next() {
		var msg models.Message
		var user models.User // Assuming you want user data for the sender

		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.ReadAt,
			&msg.MessageType,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&user.ID,
			&user.Username,
			&user.Phone,
		)
		if err != nil {
			return nil, err // Return if there's a scanning error
		}

		// Assign the user to the message sender
		msg.Sender = &user
		messages = append(messages, &msg)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
