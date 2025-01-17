package repository

import (
	"database/sql"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
)

const (
	ChatsLimit      = 20
	ChatsOffset     = 0
	LastMessageTrim = 100
)

type ChatRepository struct {
	DB *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{
		DB: db,
	}
}

func (cr *ChatRepository) Create(user1ID, user2ID uint, lastMessageID *uint) (*models.Chat, error) {
	query := `
		INSERT INTO chats (user1_id, user2_id, last_message_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, user1_id, user2_id, last_message_id, created_at, updated_at
	`

	chat := &models.Chat{}
	err := cr.DB.QueryRow(query, user1ID, user2ID, lastMessageID).Scan(
		&chat.ID,
		&chat.User1ID,
		&chat.User2ID,
		&chat.LastMessageID,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (cr *ChatRepository) GetByID(chatID uint) (*models.Chat, error) {
	query := `
		SELECT 
			c.id, 
			c.user1_id, 
			c.user2_id, 
			c.last_message_id, 
			c.created_at, 
			c.updated_at,
			u1.id AS user1_id, 
			u1.username AS user1_username, 
			u1.phone AS user1_phone, 
			u1.last_seen AS user1_last_seen, 
			u1.profile_picture AS user1_profile_picture, 
			u2.id AS user2_id, 
			u2.username AS user2_username, 
			u2.phone AS user2_phone,
			u2.last_seen AS user2_last_seen,
			u2.profile_picture AS user2_profile_picture, 
			m.id AS message_id, 
			m.sender_id, 
			m.recipient_id, 
			LEFT(
    			CASE
        			WHEN LENGTH(m.content) > $2 THEN CONCAT(SUBSTRING(m.content, 1, $2), '...')
        			ELSE m.content
    			END,
    			$3
			) AS message_content_trimmed, 
			m.message_type, 
			m.chat_id, 
			m.created_at AS message_created_at, 
			m.updated_at AS message_updated_at
		FROM chats AS c
			LEFT JOIN users AS u1 ON c.user1_id = u1.id
			LEFT JOIN users AS u2 ON c.user2_id = u2.id
			LEFT JOIN messages AS m ON c.last_message_id = m.id
		WHERE c.id = $1
	`

	chat := &models.Chat{}
	var user1, user2 models.User
	var lastMessage models.Message

	// Nullable fields
	var lastMessageID sql.NullInt64
	var lastMessageContent sql.NullString
	var lastMessageSenderID, lastMessageRecipientID, lastMessageChatID sql.NullInt64
	var lastMessageType sql.NullString
	var lastMessageCreatedAt, lastMessageUpdatedAt sql.NullTime

	err := cr.DB.QueryRow(query, chatID, LastMessageTrim, LastMessageTrim+3).Scan(
		&chat.ID, &chat.User1ID, &chat.User2ID, &chat.LastMessageID, &chat.CreatedAt, &chat.UpdatedAt,
		&user1.ID, &user1.Username, &user1.Phone, &user1.LastSeen, &user1.ProfilePicture,
		&user2.ID, &user2.Username, &user2.Phone, &user2.LastSeen, &user2.ProfilePicture,
		&lastMessageID, &lastMessageSenderID, &lastMessageRecipientID, &lastMessageContent, &lastMessageType, &lastMessageChatID, &lastMessageCreatedAt, &lastMessageUpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Assign fetched users to the chat struct
	chat.User1 = &user1
	chat.User2 = &user2

	// Only assign last message if `last_message_id` is not NULL
	if lastMessageID.Valid {
		lastMessage.ID = uint(lastMessageID.Int64)
		lastMessage.SenderID = uint(lastMessageSenderID.Int64)
		lastMessage.RecipientID = uint(lastMessageRecipientID.Int64)
		lastMessage.Content = lastMessageContent.String
		lastMessage.MessageType = models.MessageType(lastMessageType.String)
		lastMessage.ChatID = uint(lastMessageChatID.Int64)
		if lastMessageCreatedAt.Valid {
			lastMessage.CreatedAt = lastMessageCreatedAt.Time
		}
		if lastMessageUpdatedAt.Valid {
			lastMessage.UpdatedAt = lastMessageUpdatedAt.Time
		}
		chat.LastMessage = &lastMessage
	} else {
		chat.LastMessage = nil
	}

	return chat, nil
}

func (cr *ChatRepository) GetForUser(userID uint, limit, offset int) ([]*models.Chat, error) {
	query := `
		SELECT 
		    c.id,
       		c.user1_id,
       		c.user2_id,
       		c.last_message_id,
       		c.created_at,
       		c.updated_at,
       		u1.id        AS user1_id,
       		u1.username  AS user1_username,
       		u1.phone  AS user1_phone,
       		u1.last_seen  AS user1_last_seen,
       		u1.profile_picture  AS user1_profile_picture,
       		u1.created_at  AS user1_created_at,
       		u1.updated_at  AS user1_updated_at,
       		u2.id        AS user2_id,
       		u2.username  AS user2_username,
       		u2.phone  AS user2_phone,
       		u2.last_seen  AS user2_last_seen,
       		u2.profile_picture  AS user2_profile_picture,
       		u2.created_at  AS user2_created_at,
       		u2.updated_at  AS user2_updated_at,
       		m.id         AS message_id,
       		m.sender_id         AS last_message_sender_id,
       		m.recipient_id         AS last_message_recipient_id,
       		LEFT(
    			CASE
        			WHEN LENGTH(m.content) > $4 THEN CONCAT(SUBSTRING(m.content, 1, $4), '...')
        			ELSE m.content
    			END,
    			$5
			) AS message_content_trimmed,
		    m.read_at AS last_message_read_at,
		    m.message_type AS last_message_type,
		    m.chat_id AS last_message_chat_id,
       		m.created_at AS last_message_created_at,
       		m.updated_at AS last_message_updated_at
		FROM chats c
         	LEFT JOIN users u1 ON c.user1_id = u1.id
         	LEFT JOIN users u2 ON c.user2_id = u2.id
         	LEFT JOIN messages m ON c.last_message_id = m.id
		WHERE c.user1_id = $1 OR c.user2_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := cr.DB.Query(query, userID, limit, offset, LastMessageTrim, LastMessageTrim+3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*models.Chat
	for rows.Next() {
		var chat models.Chat
		var user1, user2 models.User
		var lastMessage models.Message

		// Handle nullable fields
		var lastMessageID sql.NullInt64
		var lastMessageSenderID sql.NullInt64
		var lastMessageRecipientID sql.NullInt64
		var lastMessageContent sql.NullString
		var lastMessageReadAt sql.NullTime
		var lastMessageType sql.NullString
		var lastMessageChatID sql.NullInt64
		var lastMessageCreatedAt sql.NullTime
		var lastMessageUpdatedAt sql.NullTime

		err := rows.Scan(
			&chat.ID, &chat.User1ID, &chat.User2ID, &chat.LastMessageID, &chat.CreatedAt, &chat.UpdatedAt,
			&user1.ID, &user1.Username, &user1.Phone, &user1.LastSeen, &user1.ProfilePicture, &user1.CreatedAt, &user1.UpdatedAt,
			&user2.ID, &user2.Username, &user2.Phone, &user2.LastSeen, &user2.ProfilePicture, &user2.CreatedAt, &user2.UpdatedAt,
			&lastMessageID, &lastMessageSenderID, &lastMessageRecipientID, &lastMessageContent, &lastMessageReadAt, &lastMessageType, &lastMessageChatID, &lastMessageCreatedAt, &lastMessageUpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		chat.User1 = &user1
		chat.User2 = &user2

		// Handle nullable last message
		if lastMessageID.Valid {
			lastMessage.ID = uint(lastMessageID.Int64)
			if lastMessageSenderID.Valid {
				lastMessage.SenderID = uint(lastMessageSenderID.Int64)
			}
			if lastMessageRecipientID.Valid {
				lastMessage.RecipientID = uint(lastMessageRecipientID.Int64)
			}
			if lastMessageContent.Valid {
				lastMessage.Content = lastMessageContent.String
			}
			if lastMessageReadAt.Valid {
				lastMessage.ReadAt = &lastMessageReadAt.Time
			}
			if lastMessageType.Valid {
				lastMessage.MessageType = models.MessageType(lastMessageType.String)
			}
			if lastMessageChatID.Valid {
				lastMessage.ChatID = uint(lastMessageChatID.Int64)
			}
			if lastMessageCreatedAt.Valid {
				lastMessage.CreatedAt = lastMessageCreatedAt.Time
			}
			if lastMessageUpdatedAt.Valid {
				lastMessage.UpdatedAt = lastMessageUpdatedAt.Time
			}
			chat.LastMessage = &lastMessage
		} else {
			chat.LastMessage = nil
		}

		chats = append(chats, &chat)
	}

	return chats, nil
}

func (cr *ChatRepository) UpdateLastMessage(chatID, lastMessageID uint) error {
	query := `
		UPDATE chats
		SET last_message_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := cr.DB.Exec(query, lastMessageID, chatID)
	return err
}

func (cr *ChatRepository) DeleteChat(chatID uint) error {
	query := `
		DELETE FROM chats
		WHERE id = $1
	`

	_, err := cr.DB.Exec(query, chatID)
	return err
}
