package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KEPTANy/plag-check/shared/jwt"
	"github.com/KEPTANy/plag-check/user-service/internal/model"
	"github.com/KEPTANy/plag-check/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *model.RegisterRequest) error
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
}

type userService struct {
	db         repository.UserRepository
	jwtSecret  string
	bCryptCost int
}

func NewUserService(db repository.UserRepository, jwtSecret string, bCryptCost int) UserService {
	return &userService{db: db, jwtSecret: jwtSecret, bCryptCost: bCryptCost}
}

func (u *userService) Register(ctx context.Context, req *model.RegisterRequest) error {
	password_hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), u.bCryptCost)
	if err != nil {
		return fmt.Errorf("Failed to hash password: %w", err)
	}

	_, err = u.db.CreateUser(ctx, req.Username, string(password_hash), req.Role)
	if err != nil {
		return fmt.Errorf("Failed to add user to a repository: %w", err)
	}

	return err
}

func (u *userService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := u.db.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("Failed to find user in the db: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("Invalid password or username")
	}

	token, err := jwt.GenerateToken(time.Minute*time.Duration(req.DurationMin), req.Username, user.Role, u.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate jwt token: %w", err)
	}

	return &model.LoginResponse{Token: token}, nil
}
