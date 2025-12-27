package repository

import (
	"context"
	"fmt"

	"github.com/KEPTANy/plag-check/user-service/internal/model"
	"github.com/gofrs/uuid/v5"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, password_hash, role string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
}

type userRepository struct {
	db *PgRepository
}

func NewUserRepository(db *PgRepository) UserRepository {
	return &userRepository{db: db}
}

func (u *userRepository) CreateUser(ctx context.Context, username, password_hash, role string) (*model.User, error) {
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, username, role
	`

	var user model.User
	err := u.db.pool.QueryRow(ctx, query, username, password_hash, role).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("Failed to create a user: %w", err)
	}

	return &user, nil
}

func (u *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, role
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := u.db.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("Failed to find a user: %w", err)
	}

	return &user, nil
}

func (u *userRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, role
		FROM users
		WHERE username = $1
	`

	var user model.User
	err := u.db.pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("Failed to find a user: %w", err)
	}

	return &user, nil
}
