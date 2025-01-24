package storage

import (
	"errors"
	"io"
)

type Storage interface {
	SaveFile(baseDir, fileName string, fileData io.Reader) (string, error)
	GetFile(baseDir, fileName string) (string, error)
	DeleteFile(baseDir, fileName string) error
}

type Type string

const (
	LocalStorageType Type = "local"
	LocalStoragePath      = "./uploads"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".mov":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".txt":  true,
}

type Config struct {
	Type      Type
	LocalPath string
}

func NewStorage(config *Config) (Storage, error) {
	switch config.Type {
	case LocalStorageType:
		return newLocalStorage(config.LocalPath), nil
	default:
		return nil, errors.New("unsupported storage type")
	}
}
