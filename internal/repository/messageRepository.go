package repository

import (
	"database/sql"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
	"time"
)

const (
	MessagesLimit  = 20
	MessagesOffset = 0
)

type MessageRepository struct {
	DB *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		DB: db,
	}
}

func (mr *MessageRepository) Create(msg *models.Message) (*models.Message, error) {
	query := `
		INSERT INTO messages (sender_id, recipient_id, content, chat_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := mr.DB.QueryRow(query, msg.SenderID, msg.RecipientID, msg.Content, msg.ChatID).
		Scan(&msg.ID, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (mr *MessageRepository) Edit(id uint, content string) (*models.Message, error) {
	query := `
		UPDATE messages
		SET content = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, sender_id, recipient_id, content, read_at, chat_id, created_at, updated_at
	`

	var m models.Message

	err := mr.DB.QueryRow(query, content, id).Scan(
		&m.ID,
		&m.SenderID,
		&m.RecipientID,
		&m.Content,
		&m.ReadAt,
		&m.ChatID,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (mr *MessageRepository) Delete(id uint) error {
	query := `DELETE FROM messages WHERE id = $1`

	_, err := mr.DB.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (mr *MessageRepository) GetChatMessages(chatID uint, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT 
			m.id, 
			m.sender_id, 
			m.recipient_id, 
			m.content, 
			m.read_at, 
			m.chat_id, 
			m.created_at, 
			m.updated_at,
			u1.id AS sender_id, 
			u1.username AS sender_username,
			u2.id AS recipient_id,
			u2.username AS recipient_username
		FROM messages m
		JOIN chats c ON m.chat_id = c.id
		JOIN users u1 ON m.sender_id = u1.id
		JOIN users u2 ON m.recipient_id = u2.id
		WHERE c.id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	// Query the database
	rows, err := mr.DB.Query(query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize a slice to hold the messages
	messages := make([]*models.Message, 0)

	// Iterate through the rows and scan the data into the struct
	for rows.Next() {
		var msg models.Message
		var sender, recipient models.User

		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.RecipientID, &msg.Content, &msg.ReadAt, &msg.ChatID, &msg.CreatedAt, &msg.UpdatedAt,
			&sender.ID, &sender.Username,
			&recipient.ID, &recipient.Username,
		)
		if err != nil {
			return nil, err // Return if there's a scanning error
		}

		// Assign sender and recipient users to the message
		msg.Sender = &sender
		msg.Recipient = &recipient
		messages = append(messages, &msg)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (mr *MessageRepository) GetUserMessages(senderID uint, recipientID uint, limit int, offset int) ([]*models.Message, error) {
	query := `
		SELECT 
			m.id, 
			m.sender_id, 
			m.recipient_id, 
			m.content, 
			m.read_at, 
			m.chat_id, 
			m.created_at, 
			m.updated_at,
			u.id AS user_id,
			u.username AS user_username,
			u.phone AS user_phone
		FROM messages AS m
		JOIN users AS u ON m.sender_id = u.id
		WHERE m.sender_id = $1 AND m.recipient_id = $2
		ORDER BY m.created_at desc 
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
			&msg.ID, &msg.SenderID, &msg.RecipientID, &msg.Content, &msg.ReadAt, &msg.ChatID, &msg.CreatedAt, &msg.UpdatedAt,
			&user.ID, &user.Username, &user.Phone,
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

func (mr *MessageRepository) GetLastMessageForChat(chatID uint) (*models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, content, read_at, chat_id, created_at, updated_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var message models.Message

	err := mr.DB.QueryRow(query, chatID).Scan(
		&message.ID,
		&message.SenderID,
		&message.RecipientID,
		&message.Content,
		&message.ReadAt,
		&message.ChatID,
		&message.CreatedAt,
		&message.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

func (mr *MessageRepository) GetById(id uint) (*models.Message, error) {
	query := `
		SELECT id, sender_id, recipient_id, content, read_at, chat_id, created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	var m models.Message

	err := mr.DB.QueryRow(query, id).Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.Content, &m.ReadAt, &m.ChatID, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}

func (mr *MessageRepository) MarkAsRead(id uint) (*time.Time, error) {
	query := `
		UPDATE messages SET read_at = NOW() WHERE id = $1
		RETURNING read_at
	`

	var readAt time.Time

	err := mr.DB.QueryRow(query, id).Scan(&readAt)
	return &readAt, err
}
