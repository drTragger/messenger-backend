package repository

import (
	"database/sql"
	"github.com/drTragger/messenger-backend/internal/models"
)

type AttachmentRepository struct {
	DB *sql.DB
}

func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{
		DB: db,
	}
}

func (ar *AttachmentRepository) Create(attachment *models.Attachment) (*models.Attachment, error) {
	query := `
		INSERT INTO attachments (message_id, file_name, file_path, file_type, file_size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`

	row := ar.DB.QueryRow(query, attachment.MessageID, attachment.FileName, attachment.FilePath, attachment.FileType, attachment.FileSize)
	err := row.Scan(&attachment.ID, &attachment.CreatedAt, &attachment.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return attachment, nil
}
