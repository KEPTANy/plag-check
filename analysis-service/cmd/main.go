package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KEPTANy/plag-check/analysis-service/internal/config"
	"github.com/KEPTANy/plag-check/analysis-service/internal/handler"
	intMiddleware "github.com/KEPTANy/plag-check/analysis-service/internal/middleware"
	"github.com/KEPTANy/plag-check/analysis-service/internal/repository"
	"github.com/KEPTANy/plag-check/analysis-service/internal/service"
	"github.com/KEPTANy/plag-check/analysis-service/internal/storage"
	"github.com/KEPTANy/plag-check/shared/middleware"
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

	fileRepo := repository.NewFileRepository(db)
	fileStorage, err := storage.NewStorage(cfg.StorageRoot)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}
	analysisService := service.NewAnalysisService(fileRepo, fileStorage)

	healthHandler := handler.NewHealthHandler()
	analysisHandler := handler.NewAnalysisHandler(analysisService)

	mux := http.NewServeMux()

	mux.Handle("GET /health", http.HandlerFunc(healthHandler.Health))

	baseChain := middleware.Chain(
		intMiddleware.AuthMiddleware(cfg.JWTSecret),
	)

	teacherChain := middleware.Chain(
		baseChain,
		intMiddleware.RequireRole("teacher"),
	)

	mux.Handle("GET /analysis/plagiarism", teacherChain(http.HandlerFunc(analysisHandler.CheckPlagiarism)))

	mux.Handle("GET /analysis/wordcloud/{id}", teacherChain(http.HandlerFunc(analysisHandler.GetWordCloud)))

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
		log.Printf("Analysis service starting on port: %v", cfg.Port)
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
