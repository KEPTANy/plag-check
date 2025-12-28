package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(connString string) (*PgRepository, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse connection string: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5

	ctx, close := context.WithTimeout(context.Background(), 10*time.Second)
	defer close()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	return &PgRepository{pool: pool}, err
}

func (repo *PgRepository) Close() {
	repo.pool.Close()
}

func (repo *PgRepository) GetPool() *pgxpool.Pool {
	return repo.pool
}
