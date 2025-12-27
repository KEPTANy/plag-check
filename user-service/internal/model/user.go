package model

import "github.com/gofrs/uuid/v5"

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
}

func IsValidRole(role string) bool {
	return role == "student" || role == "teacher"
}
