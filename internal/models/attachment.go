package models

import "time"

type Attachment struct {
	ID          uint      `json:"id"`
	MessageID   uint      `json:"messageId"`
	FileName    string    `json:"fileName"`
	FilePath    string    `json:"filePath"`
	FileType    string    `json:"fileType"`
	FileSize    int64     `json:"fileSize"`
	ThumbnailID uint      `json:"thumbnailId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Thumbnail *Thumbnail `json:"thumbnail"`
}
