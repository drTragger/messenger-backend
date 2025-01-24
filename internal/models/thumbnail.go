package models

type Thumbnail struct {
	ID        uint   `json:"id"`
	FilePath  string `json:"filePath"`
	FileType  string `json:"fileType"`
	FileSize  int64  `json:"fileSize"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
