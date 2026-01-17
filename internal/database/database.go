package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB with additional methods.
type DB struct {
	*sql.DB
}

// New creates a new database connection.
func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Migrate runs database migrations.
func (db *DB) Migrate(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS chat_settings (
			chat_id BIGINT PRIMARY KEY,
			language VARCHAR(2) NOT NULL DEFAULT 'en',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_chat_settings_language ON chat_settings(language);
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
