package repository

import (
	"context"
	"fmt"

	"github.com/KEPTANy/plag-check/file-storage-service/internal/model"
	"github.com/gofrs/uuid/v5"
)

type FileRepository interface {
	AddFile(ctx context.Context, file *model.File) (int, error)
	GetFileByID(ctx context.Context, id int) (*model.File, error)
	GetFilesByStudent(ctx context.Context, studentID uuid.UUID) ([]model.File, error)
	GetFilesByHash(ctx context.Context, hash string) ([]model.File, error)
}

type fileRepository struct {
	db *PgRepository
}

func NewFileRepository(db *PgRepository) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) AddFile(ctx context.Context, file *model.File) (int, error) {
	query := `
		INSERT INTO files (
			student_id, file_hash, file_size, storage_path, original_filename
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.pool.QueryRow(
		ctx, query, file.StudentID, file.FileHash, file.FileSize, file.StoragePath, file.Filename,
	).Scan(&file.ID)

	if err != nil {
		return 0, fmt.Errorf("Failed to add file to database: %w", err)
	}

	return file.ID, nil
}

func (r *fileRepository) GetFileByID(ctx context.Context, id int) (*model.File, error) {
	query := `
		SELECT id, student_id, file_hash, file_size, storage_path, original_filename
		FROM files
		WHERE id = $1
	`

	var file model.File
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&file.ID, &file.StudentID, &file.FileHash, &file.FileSize, &file.StoragePath, &file.Filename,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to get file info: %w", err)
	}

	return &file, nil
}

func (r *fileRepository) GetFilesByStudent(ctx context.Context, studentID uuid.UUID) ([]model.File, error) {
	query := `
		SELECT id, student_id, file_hash, file_size, storage_path, original_filename
		FROM files
		WHERE student_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files' info by student_id: %w", err)
	}
	defer rows.Close()

	var files []model.File
	for rows.Next() {
		var file model.File
		err := rows.Scan(
			&file.ID, &file.StudentID, &file.FileHash, &file.FileSize, &file.StoragePath, &file.Filename,
		)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan file info: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}

func (r *fileRepository) GetFilesByHash(ctx context.Context, hash string) ([]model.File, error) {
	query := `
		SELECT id, student_id, file_hash, file_size, storage_path, original_filename
		FROM files
		WHERE file_hash = $1
		ORDER BY id ASC
	`

	rows, err := r.db.pool.Query(ctx, query, hash)
	if err != nil {
		return nil, fmt.Errorf("Failed to get files' info by hash: %w", err)
	}
	defer rows.Close()

	var files []model.File
	for rows.Next() {
		var file model.File
		err := rows.Scan(
			&file.ID, &file.StudentID, &file.FileHash, &file.FileSize, &file.StoragePath, &file.Filename,
		)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan file info: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}
