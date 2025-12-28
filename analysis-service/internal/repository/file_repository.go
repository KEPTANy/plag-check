package repository

import (
	"context"
	"fmt"

	"github.com/KEPTANy/plag-check/analysis-service/internal/model"
)

type FileRepository interface {
	GetFileByID(ctx context.Context, id int) (*model.File, error)
	GetFilesByHash(ctx context.Context, hash string) ([]model.File, error)
	GetPlagiarismGroups(ctx context.Context) ([]model.PlagiarismResult, error)
}

type fileRepository struct {
	db *PgRepository
}

func NewFileRepository(db *PgRepository) FileRepository {
	return &fileRepository{db: db}
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

func (r *fileRepository) GetPlagiarismGroups(ctx context.Context) ([]model.PlagiarismResult, error) {
	query := `
		SELECT file_hash, COUNT(DISTINCT student_id) as count
		FROM files
		GROUP BY file_hash
		HAVING COUNT(DISTINCT student_id) > 1
		ORDER BY count DESC
	`

	rows, err := r.db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Failed to get plagiarism groups: %w", err)
	}
	defer rows.Close()

	var results []model.PlagiarismResult
	for rows.Next() {
		var result model.PlagiarismResult
		err := rows.Scan(&result.Hash, &result.Count)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan plagiarism result: %w", err)
		}

		files, err := r.GetFilesByHash(ctx, result.Hash)
		if err != nil {
			return nil, fmt.Errorf("Failed to get files for hash: %w", err)
		}

		result.Files = files
		results = append(results, result)
	}

	return results, nil
}
