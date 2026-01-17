package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Language constants for chat settings.
const (
	LangEnglish   = "en"
	LangUkrainian = "uk"
	LangRussian   = "ru"

	DefaultLanguage = LangEnglish
)

// ValidLanguages contains all supported language codes.
var ValidLanguages = []string{LangEnglish, LangUkrainian, LangRussian}

// IsValidLanguage checks if the language code is supported.
func IsValidLanguage(lang string) bool {
	for _, l := range ValidLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

// ChatSettings represents chat settings stored in the database.
type ChatSettings struct {
	ChatID    int64
	Language  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ChatSettingsRepository handles chat settings database operations.
type ChatSettingsRepository struct {
	db *DB
}

// NewChatSettingsRepository creates a new ChatSettingsRepository.
func NewChatSettingsRepository(db *DB) *ChatSettingsRepository {
	return &ChatSettingsRepository{db: db}
}

// GetOrCreate returns chat settings for the given chat ID.
// If settings don't exist, creates them with default language.
func (r *ChatSettingsRepository) GetOrCreate(ctx context.Context, chatID int64) (*ChatSettings, error) {
	settings, err := r.Get(ctx, chatID)
	if err == nil {
		return settings, nil
	}

	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	// Create new settings with default language
	return r.Create(ctx, chatID, DefaultLanguage)
}

// Get returns chat settings for the given chat ID.
func (r *ChatSettingsRepository) Get(ctx context.Context, chatID int64) (*ChatSettings, error) {
	query := `
		SELECT chat_id, language, created_at, updated_at
		FROM chat_settings
		WHERE chat_id = $1
	`

	var settings ChatSettings
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&settings.ChatID,
		&settings.Language,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chat settings: %w", err)
	}

	return &settings, nil
}

// Create creates new chat settings.
func (r *ChatSettingsRepository) Create(ctx context.Context, chatID int64, language string) (*ChatSettings, error) {
	if !IsValidLanguage(language) {
		return nil, fmt.Errorf("invalid language: %s", language)
	}

	query := `
		INSERT INTO chat_settings (chat_id, language)
		VALUES ($1, $2)
		ON CONFLICT (chat_id) DO UPDATE SET language = EXCLUDED.language, updated_at = CURRENT_TIMESTAMP
		RETURNING chat_id, language, created_at, updated_at
	`

	var settings ChatSettings
	err := r.db.QueryRowContext(ctx, query, chatID, language).Scan(
		&settings.ChatID,
		&settings.Language,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create chat settings: %w", err)
	}

	return &settings, nil
}

// UpdateLanguage updates the language for a chat.
func (r *ChatSettingsRepository) UpdateLanguage(ctx context.Context, chatID int64, language string) (*ChatSettings, error) {
	if !IsValidLanguage(language) {
		return nil, fmt.Errorf("invalid language: %s", language)
	}

	query := `
		UPDATE chat_settings
		SET language = $2, updated_at = CURRENT_TIMESTAMP
		WHERE chat_id = $1
		RETURNING chat_id, language, created_at, updated_at
	`

	var settings ChatSettings
	err := r.db.QueryRowContext(ctx, query, chatID, language).Scan(
		&settings.ChatID,
		&settings.Language,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		// If not exists, create with the specified language
		return r.Create(ctx, chatID, language)
	}
	if err != nil {
		return nil, fmt.Errorf("update chat language: %w", err)
	}

	return &settings, nil
}

// GetLanguage returns the language for a chat, or default if not set.
func (r *ChatSettingsRepository) GetLanguage(ctx context.Context, chatID int64) (string, error) {
	settings, err := r.GetOrCreate(ctx, chatID)
	if err != nil {
		return DefaultLanguage, err
	}
	return settings.Language, nil
}

// ErrNotFound is returned when chat settings are not found.
var ErrNotFound = errors.New("chat settings not found")
