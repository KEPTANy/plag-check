package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KEPTANy/plag-check/shared/middleware"
	"github.com/KEPTANy/plag-check/user-service/internal/config"
	"github.com/KEPTANy/plag-check/user-service/internal/handler"
	"github.com/KEPTANy/plag-check/user-service/internal/repository"
	"github.com/KEPTANy/plag-check/user-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var cfg config.Config
	err := cfg.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.NewPgRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db.GetPool()); err != nil {
		log.Printf("WARNING! Failed to run migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret, cfg.BCryptCost)

	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler(userService)

	mux := http.NewServeMux()

	mux.Handle("GET /health", http.HandlerFunc(healthHandler.Health))

	mux.Handle("POST /auth/register", http.HandlerFunc(userHandler.Register))
	mux.Handle("POST /auth/login", http.HandlerFunc(userHandler.Login))

	handler := middleware.Chain(
		middleware.RecoveringMiddleware,
		middleware.LoggingMiddleware,
	)(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("User service starting on port: %v", cfg.Port)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}

func runMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	var tableExists bool
	err := pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'users'
        )
    `).Scan(&tableExists)

	if err != nil {
		return err
	}

	if !tableExists {
		log.Println("Running database migrations...")

		migrationSQL, err := os.ReadFile("migrations/001_create_users_table.up.sql")
		if err != nil {
			return err
		}

		_, err = pool.Exec(ctx, string(migrationSQL))
		if err != nil {
			return err
		}

		log.Println("Database migrations completed")
	}

	return nil
}
