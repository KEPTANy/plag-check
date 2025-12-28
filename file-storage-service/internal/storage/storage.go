package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type Storage interface {
	SaveFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (hash, storagePath string, size int64, err error)
	GetFile(ctx context.Context, storagePath string) (rc io.ReadCloser, size int64, err error)
	FileExists(ctx context.Context, storagePath string) bool
}

type storage struct {
	root        string
	maxFileSize int64
}

func NewStorage(root string, maxFileSize int64) (Storage, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("Failed to create root directory of a storage: %w", err)
	}
	return &storage{root: root, maxFileSize: maxFileSize}, nil
}

func storagePathFromHash(hash string) string {
	if len(hash) < 4 {
		return hash + ".dat"
	}
	return filepath.Join(hash[0:2], hash[2:4], hash) + ".dat"
}

func (s *storage) SaveFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (hash, storagePath string, size int64, err error) {
	if header.Size > s.maxFileSize {
		return "", "", 0, errors.New("File size excedes max file size")
	}

	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return "", "", 0, fmt.Errorf("Failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	hashWriter := sha256.New()
	file_ := io.TeeReader(file, hashWriter)

	hash = hex.EncodeToString(hashWriter.Sum(nil))
	storagePath = storagePathFromHash(hash)
	fullPath := filepath.Join(s.root, storagePath)

	if exists := s.FileExists(ctx, storagePath); exists {
		return hash, storagePath, size, nil
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", 0, fmt.Errorf("Failed to create directory: %w", err)
	}

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("Failed to create file: %w", err)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, file_)
	if err != nil {
		os.Remove(fullPath)
		return "", "", 0, fmt.Errorf("Failed to save file: %w", err)
	}

	if written != header.Size {
		return "", "", 0, fmt.Errorf("File size mismatch")
	}

	return hash, storagePath, written, nil
}

func (s *storage) GetFile(ctx context.Context, storagePath string) (rc io.ReadCloser, size int64, err error) {
	fullPath := filepath.Join(s.root, storagePath)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("File not found: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to open file: %w", err)
	}

	return file, fileInfo.Size(), nil
}

func (s *storage) FileExists(ctx context.Context, storagePath string) bool {
	_, err := os.Stat(storagePath)
	return err == nil || !os.IsNotExist(err)
}
