package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KEPTANy/plag-check/gateway-api/internal/config"
	"github.com/KEPTANy/plag-check/gateway-api/internal/handler"
	"github.com/KEPTANy/plag-check/gateway-api/internal/proxy"
	"github.com/KEPTANy/plag-check/shared/middleware"
)

func main() {
	var cfg config.Config
	err := cfg.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	reverseProxy := proxy.NewReverseProxy(
		cfg.UserServiceURL,
		cfg.FileStorageServiceURL,
		cfg.AnalysisServiceURL,
	)

	healthHandler := handler.NewHealthHandler()

	mux := http.NewServeMux()

	mux.Handle("GET /health", http.HandlerFunc(healthHandler.Health))

	mux.Handle("POST /auth/", http.HandlerFunc(reverseProxy.ProxyRequest))
	mux.Handle("GET /auth/", http.HandlerFunc(reverseProxy.ProxyRequest))

	mux.Handle("POST /files/", http.HandlerFunc(reverseProxy.ProxyRequest))
	mux.Handle("GET /files/", http.HandlerFunc(reverseProxy.ProxyRequest))

	mux.Handle("GET /analysis/", http.HandlerFunc(reverseProxy.ProxyRequest))

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
		log.Printf("Gateway API starting on port: %v", cfg.Port)
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
