package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	GetFile(ctx context.Context, storagePath string) (rc io.ReadCloser, size int64, err error)
}

type storage struct {
	root string
}

func NewStorage(root string) (Storage, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("Failed to create root directory of a storage: %w", err)
	}
	return &storage{root: root}, nil
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
