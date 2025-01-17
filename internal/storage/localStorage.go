package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

type LocalStorage struct {
	BasePath string
}

func newLocalStorage(basePath string) *LocalStorage {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		err := os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			panic(fmt.Errorf("failed to create storage directory '%s': %w", basePath, err))
		}
	}
	return &LocalStorage{BasePath: basePath}
}

func (l *LocalStorage) SaveFile(fileName string, fileData io.Reader) (string, error) {
	ext := filepath.Ext(fileName)
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext) // Unique timestamp-based name
	filePath := l.buildFilePath(newFileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file '%s': %v", filePath, err)
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, fileData)
	if err != nil {
		log.Printf("Failed to copy data to file '%s': %v", filePath, err)
		return "", err
	}

	return newFileName, nil
}

func (l *LocalStorage) GetFile(fileName string) (io.ReadCloser, error) {
	// Build the absolute path to the file
	filePath := l.buildFilePath(fileName)

	// Attempt to open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	return file, nil
}

func (l *LocalStorage) DeleteFile(fileName string) error {
	// Build the absolute path to the file
	filePath := l.buildFilePath(fileName)

	// Check if the file exists before trying to delete it
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file '%s' does not exist", filePath)
	}

	// Attempt to delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file '%s': %w", filePath, err)
	}

	return nil
}

func (l *LocalStorage) buildFilePath(fileName string) string {
	return filepath.Join(l.BasePath, filepath.FromSlash(fileName))
}
