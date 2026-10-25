package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	Pool *sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	if databaseURL == "" {
		log.Println("[Database] DATABASE_URL not set. Running in offline/mock mode.")
		return &DB{Pool: nil}, nil
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Supabase connection pool tuning
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Printf("[Database Warning] Could not connect to Supabase: %v", err)
		// Return DB instance so the app continues running on Railway
		return &DB{Pool: db}, nil
	}

	log.Println("[Database] Successfully connected to Supabase PostgreSQL database.")
	return &DB{Pool: db}, nil
}

func (d *DB) IsConnected() bool {
	if d == nil || d.Pool == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return d.Pool.PingContext(ctx) == nil
}
