package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nurdiansyah0/arff-os-backend/internal/config"
	"github.com/Nurdiansyah0/arff-os-backend/internal/database"
	"github.com/Nurdiansyah0/arff-os-backend/internal/handler"
)

// Middleware CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging Middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s - %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
	})
}

func main() {
	cfg := config.Load()

	// Connect to Supabase PostgreSQL database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Printf("[Warning] Database initialization error: %v", err)
	}
	if db != nil && db.Pool != nil {
		defer db.Pool.Close()
	}

	mux := http.NewServeMux()

	healthH := &handler.HealthHandler{
		DBCheck: func() bool {
			return db != nil && db.IsConnected()
		},
	}

	// Railway Healthchecks & API Status
	mux.Handle("GET /health", healthH)
	mux.Handle("GET /api/health", healthH)

	// API v1 Endpoints
	mux.HandleFunc("GET /api/v1/dashboard/stats", handler.GetDashboardStats)

	// Apply Middlewares
	var rootHandler http.Handler = mux
	rootHandler = corsMiddleware(rootHandler)
	rootHandler = loggingMiddleware(rootHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel for graceful shutdown signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("ARFF-OS Golang Backend listening on port %s (env: %s)", cfg.Port, cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down ARFF-OS server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	log.Println("ARFF-OS Server exited successfully.")
}
