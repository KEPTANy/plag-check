package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		writer := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK, written: false}
		defer func() {
			log.Printf("[%s] %s %s - %d (%s)",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				writer.statusCode,
				time.Since(start),
			)
		}()

		next.ServeHTTP(writer, r)
	})
}

func RecoveringMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC RECOVERED: %v", err)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"error": "internal server error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
