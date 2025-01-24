package services

import (
	"fmt"
	"github.com/drTragger/messenger-backend/internal/models"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/storage"
	"log"
	"mime/multipart"
	"net/http"
	"runtime"
	"sync"
)

type MessageService struct {
	AttachmentRepo *repository.AttachmentRepository
	Storage        storage.Storage
}

func NewMessageService(
	attachmentRepo *repository.AttachmentRepository,
	storage storage.Storage,
) *MessageService {
	return &MessageService{
		AttachmentRepo: attachmentRepo,
		Storage:        storage,
	}
}

func (s *MessageService) ProcessAttachments(r *http.Request, message *models.Message) error {
	files := r.MultipartForm.File["attachments"]
	if len(files) == 0 {
		return nil
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(files) {
		numWorkers = len(files)
	}

	jobChan := make(chan struct {
		index int
		file  *multipart.FileHeader
	}, len(files))
	errChan := make(chan error, len(files))
	attachments := make([]*models.Attachment, len(files))

	worker := func() {
		for job := range jobChan {
			attachment, err := s.processSingleAttachment(job.file, message.ID)
			if err != nil {
				errChan <- err
				continue
			}
			attachments[job.index] = attachment
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker()
		}()
	}

	for i, fh := range files {
		jobChan <- struct {
			index int
			file  *multipart.FileHeader
		}{index: i, file: fh}
	}
	close(jobChan)

	wg.Wait()
	close(errChan)

	var combinedErr error
	for err := range errChan {
		if err != nil {
			if combinedErr == nil {
				combinedErr = err
			} else {
				combinedErr = fmt.Errorf("%v; %w", combinedErr, err)
			}
		}
	}

	message.Attachments = attachments
	return combinedErr
}

func (s *MessageService) DeleteAttachments(attachments []*models.Attachment) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(attachments))

	for _, attachment := range attachments {
		wg.Add(1)
		go func(attachment *models.Attachment) {
			defer wg.Done()
			if err := s.Storage.DeleteFile(storage.MessageAttachmentsDir, attachment.FilePath); err != nil {
				log.Printf("Error deleting file %s: %s", attachment.FilePath, err.Error())
				errorChan <- err
			}
		}(attachment)
	}

	wg.Wait()
	close(errorChan)

	var combinedErr error
	for err := range errorChan {
		if combinedErr == nil {
			combinedErr = err
		} else {
			combinedErr = fmt.Errorf("%v; %w", combinedErr, err)
		}
	}
	return combinedErr
}

func (s *MessageService) processSingleAttachment(fh *multipart.FileHeader, messageID uint) (*models.Attachment, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	storagePath, err := s.Storage.SaveFile(storage.MessageAttachmentsDir, fh.Filename, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	attachment := &models.Attachment{
		MessageID: messageID,
		FileName:  fh.Filename,
		FilePath:  storagePath,
		FileType:  fh.Header.Get("Content-Type"),
		FileSize:  fh.Size,
	}

	if _, err := s.AttachmentRepo.Create(attachment); err != nil {
		return nil, fmt.Errorf("failed to save attachment record: %w", err)
	}

	return attachment, nil
}
