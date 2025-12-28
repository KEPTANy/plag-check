package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/KEPTANy/plag-check/file-storage-service/internal/model"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/repository"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/storage"
	"github.com/gofrs/uuid/v5"
)

type FileStorageService interface {
	UploadFile(ctx context.Context, studentID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*model.File, error)
	DownloadFile(ctx context.Context, fileID int) (*model.File, io.ReadCloser, error)
	ListFilesByUser(ctx context.Context, studentID uuid.UUID) ([]model.File, error)
	ListFilesByHash(ctx context.Context, hash string) ([]model.File, error)
}

type fileStorageService struct {
	db      repository.FileRepository
	storage storage.Storage
}

func NewFileStorageService(db repository.FileRepository, storage storage.Storage) FileStorageService {
	return &fileStorageService{db: db, storage: storage}
}

func (s *fileStorageService) UploadFile(ctx context.Context, studentID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*model.File, error) {
	hash, storagePath, size, err := s.storage.SaveFile(ctx, file, header)
	if err != nil {
		return nil, fmt.Errorf("Failed to save file to storage: %w", err)
	}

	// file id is gonna be set by db
	fileData := &model.File{
		StudentID:   studentID,
		Filename:    header.Filename,
		FileSize:    size,
		FileHash:    hash,
		StoragePath: storagePath,
	}
	fileData.ID, err = s.db.AddFile(ctx, fileData)
	if err != nil {
		return nil, fmt.Errorf("Failed to add file to db: %w", err)
	}

	return fileData, nil
}

func (s *fileStorageService) DownloadFile(ctx context.Context, fileID int) (*model.File, io.ReadCloser, error) {
	file, err := s.db.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get file info from db: %w", err)
	}

	rc, _, err := s.storage.GetFile(ctx, file.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get file reader: %w", err)
	}

	return file, rc, nil
}

func (s *fileStorageService) ListFilesByUser(ctx context.Context, studentID uuid.UUID) ([]model.File, error) {
	files, err := s.db.GetFilesByStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files from db: %w", err)
	}

	return files, nil
}

func (s *fileStorageService) ListFilesByHash(ctx context.Context, hash string) ([]model.File, error) {
	files, err := s.db.GetFilesByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files from db: %w", err)
	}

	return files, nil
}
