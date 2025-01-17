package storage

import (
	"errors"
	"io"
)

type Storage interface {
	SaveFile(fileName string, fileData io.Reader) (string, error)
	GetFile(fileName string) (io.ReadCloser, error)
	DeleteFile(fileName string) error
}

type Type string

const (
	LocalStorageType Type = "local"
	LocalStoragePath      = "./uploads/profile_pictures"
)

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
