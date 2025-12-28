package model

import "github.com/gofrs/uuid/v5"

type File struct {
	ID          int       `json:"id"`
	StudentID   uuid.UUID `json:"student_id"`
	Filename    string    `json:"filename"`
	FileSize    int64     `json:"file_size"`
	FileHash    string    `json:"file_hash"`
	StoragePath string    `json:"-"`
}

type PlagiarismResult struct {
	Hash  string `json:"hash"`
	Files []File `json:"files"`
	Count int    `json:"count"`
}
