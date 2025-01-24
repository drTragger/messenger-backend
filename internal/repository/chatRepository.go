package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/drTragger/messenger-backend/db"
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
			u1.first_name AS user1_first_name, 
			u1.last_name AS user1_last_name, 
			u1.phone AS user1_phone, 
			u1.last_seen AS user1_last_seen, 
			u1.profile_picture AS user1_profile_picture, 
			u2.id AS user2_id, 
			u2.username AS user2_username, 
			u2.first_name AS user2_first_name, 
			u2.last_name AS user2_last_name, 
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
	var lastMessageCreatedAt, lastMessageUpdatedAt sql.NullTime

	err := cr.DB.QueryRow(query, chatID, LastMessageTrim, LastMessageTrim+3).Scan(
		&chat.ID, &chat.User1ID, &chat.User2ID, &chat.LastMessageID, &chat.CreatedAt, &chat.UpdatedAt,
		&user1.ID, &user1.Username, &user1.FirstName, &user1.LastName, &user1.Phone, &user1.LastSeen, &user1.ProfilePicture,
		&user2.ID, &user2.Username, &user2.FirstName, &user2.LastName, &user2.Phone, &user2.LastSeen, &user2.ProfilePicture,
		&lastMessageID, &lastMessageSenderID, &lastMessageRecipientID, &lastMessageContent, &lastMessageChatID, &lastMessageCreatedAt, &lastMessageUpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Assign fetched users to the chats struct
	chat.User1 = &user1
	chat.User2 = &user2

	// Only assign last message if `last_message_id` is not NULL
	if lastMessageID.Valid {
		lastMessage.ID = uint(lastMessageID.Int64)
		lastMessage.SenderID = uint(lastMessageSenderID.Int64)
		lastMessage.RecipientID = uint(lastMessageRecipientID.Int64)
		lastMessage.Content = &lastMessageContent.String
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
	query, err := db.LoadQuery("get_for_user.sql", "chats")
	if err != nil {
		return nil, fmt.Errorf("failed to load query: %w", err)
	}

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
		var lastAttachment models.Attachment

		// Handle nullable fields for the last message
		var lastMessageID sql.NullInt64
		var lastMessageSenderID sql.NullInt64
		var lastMessageRecipientID sql.NullInt64
		var lastMessageContent sql.NullString
		var lastMessageReadAt sql.NullTime
		var lastMessageChatID sql.NullInt64
		var lastMessageCreatedAt sql.NullTime
		var lastMessageUpdatedAt sql.NullTime

		// Handle nullable fields for the last attachment
		var lastAttachmentID sql.NullInt64
		var lastAttachmentFileName sql.NullString
		var lastAttachmentFilePath sql.NullString
		var lastAttachmentFileType sql.NullString
		var lastAttachmentFileSize sql.NullInt64
		var lastAttachmentCreatedAt sql.NullTime
		var lastAttachmentUpdatedAt sql.NullTime

		err := rows.Scan(
			&chat.ID, &chat.User1ID, &chat.User2ID, &chat.LastMessageID, &chat.CreatedAt, &chat.UpdatedAt,
			&user1.ID, &user1.Username, &user1.FirstName, &user1.LastName, &user1.Phone, &user1.LastSeen, &user1.ProfilePicture, &user1.CreatedAt, &user1.UpdatedAt,
			&user2.ID, &user2.Username, &user2.FirstName, &user2.LastName, &user2.Phone, &user2.LastSeen, &user2.ProfilePicture, &user2.CreatedAt, &user2.UpdatedAt,
			&lastMessageID, &lastMessageSenderID, &lastMessageRecipientID, &lastMessageContent, &lastMessageReadAt, &lastMessageChatID, &lastMessageCreatedAt, &lastMessageUpdatedAt,
			&lastAttachmentID, &lastAttachmentFileName, &lastAttachmentFilePath, &lastAttachmentFileType, &lastAttachmentFileSize, &lastAttachmentCreatedAt, &lastAttachmentUpdatedAt,
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
				lastMessage.Content = &lastMessageContent.String
			}
			if lastMessageReadAt.Valid {
				lastMessage.ReadAt = &lastMessageReadAt.Time
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

			// Handle nullable last attachment
			if lastAttachmentID.Valid {
				lastAttachment.ID = uint(lastAttachmentID.Int64)
				if lastAttachmentFileName.Valid {
					lastAttachment.FileName = lastAttachmentFileName.String
				}
				if lastAttachmentFilePath.Valid {
					lastAttachment.FilePath = lastAttachmentFilePath.String
				}
				if lastAttachmentFileType.Valid {
					lastAttachment.FileType = lastAttachmentFileType.String
				}
				if lastAttachmentFileSize.Valid {
					lastAttachment.FileSize = lastAttachmentFileSize.Int64
				}
				if lastAttachmentCreatedAt.Valid {
					lastAttachment.CreatedAt = lastAttachmentCreatedAt.Time
				}
				if lastAttachmentUpdatedAt.Valid {
					lastAttachment.UpdatedAt = lastAttachmentUpdatedAt.Time
				}
				lastMessage.Attachments = []*models.Attachment{&lastAttachment}
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
