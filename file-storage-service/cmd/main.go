package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KEPTANy/plag-check/file-storage-service/internal/config"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/handler"
	intMiddleware "github.com/KEPTANy/plag-check/file-storage-service/internal/middleware"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/repository"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/service"
	"github.com/KEPTANy/plag-check/file-storage-service/internal/storage"
	"github.com/KEPTANy/plag-check/shared/middleware"
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

	fileRepo := repository.NewFileRepository(db)
	fileStorage, err := storage.NewStorage(cfg.StorageRoot, cfg.MaxFileSize)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}
	fileService := service.NewFileStorageService(fileRepo, fileStorage)

	healthHandler := handler.NewHealthHandler()
	fileHandler := handler.NewFileStorageHandler(fileService)

	mux := http.NewServeMux()

	mux.Handle("GET /health", http.HandlerFunc(healthHandler.Health))

	baseChain := middleware.Chain(
		intMiddleware.AuthMiddleware(cfg.JWTSecret),
	)

	studentChain := middleware.Chain(
		baseChain,
		intMiddleware.RequireRole("student"),
	)

	teacherChain := middleware.Chain(
		baseChain,
		intMiddleware.RequireRole("teacher"),
	)

	studentOrTeacherChain := middleware.Chain(
		baseChain,
		intMiddleware.RequireAnyRole("student", "teacher"),
	)

	mux.Handle("POST /files/upload", studentChain(http.HandlerFunc(fileHandler.Upload)))

	mux.Handle("GET /files/download/{id}", studentOrTeacherChain(http.HandlerFunc(fileHandler.Download)))

	mux.Handle("GET /files/user/{userid}",
		studentOrTeacherChain(http.HandlerFunc(fileHandler.ListUserFiles)))

	mux.Handle("GET /files/hash/{hash}",
		teacherChain(http.HandlerFunc(fileHandler.ListFilesByHash)))

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
            AND table_name = 'files'
        )
    `).Scan(&tableExists)

	if err != nil {
		return err
	}

	if !tableExists {
		log.Println("Running database migrations...")

		migrationSQL, err := os.ReadFile("migrations/001_create_files_table.up.sql")
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
