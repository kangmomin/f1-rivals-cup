package database

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// DB holds the database connection pool
type DB struct {
	Pool *sql.DB
}

// New creates a new database connection
func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	slog.Info("Database connected successfully")

	return &DB{Pool: db}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
