package config

import "testing"

func TestLoad_TelegramAPIURL_DefaultEmpty(t *testing.T) {
	t.Setenv("BOT_TOKEN", "123:ABC")
	t.Setenv("DEBUG", "false")
	t.Setenv("TELEGRAM_API_URL", "")
	t.Setenv("TWITTERX_API_URL", "http://localhost:8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TelegramAPIURL != "" {
		t.Fatalf("TelegramAPIURL = %q, want empty", cfg.TelegramAPIURL)
	}
}

func TestLoad_TelegramAPIURL_TrimSpacesAndSlash(t *testing.T) {
	t.Setenv("BOT_TOKEN", "123:ABC")
	t.Setenv("DEBUG", "false")
	t.Setenv("TELEGRAM_API_URL", "  http://127.0.0.1:9999/  ")
	t.Setenv("TWITTERX_API_URL", "http://localhost:8080")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TelegramAPIURL != "http://127.0.0.1:9999" {
		t.Fatalf("TelegramAPIURL = %q, want %q", cfg.TelegramAPIURL, "http://127.0.0.1:9999")
	}
}

func TestLoad_BotTokenRequired(t *testing.T) {
	t.Setenv("BOT_TOKEN", "")
	t.Setenv("DEBUG", "false")
	t.Setenv("TWITTERX_API_URL", "http://localhost:8080")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}
