package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewBot_UsesConfiguredTelegramAPIURL verifies that when TELEGRAM_API_URL is set,
// the bot client will attempt to use that URL for API calls.
// We set up a mock HTTP server and verify it receives the getMe call.
func TestNewBot_UsesConfiguredTelegramAPIURL(t *testing.T) {
	called := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// gotgbot calls /bot<token>/getMe on NewBot
		if r.URL.Path == "/bot123:ABC/getMe" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":123,"is_bot":true,"first_name":"TestBot","username":"testbot"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	// Set environment variables
	t.Setenv("BOT_TOKEN", "123:ABC")
	t.Setenv("DEBUG", "false")
	t.Setenv("TELEGRAM_API_URL", mockServer.URL)
	t.Setenv("TWITTERX_API_URL", "http://127.0.0.1:8080")

	bot, updater, logger, err := NewBot()
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	if bot == nil {
		t.Fatal("bot is nil")
	}
	if updater == nil {
		t.Fatal("updater is nil")
	}
	if logger == nil {
		t.Fatal("logger is nil")
	}

	if !called {
		t.Fatal("mock server was not called - bot did not use configured TELEGRAM_API_URL")
	}

	// Verify bot identity from mock response
	if bot.User.Username != "testbot" {
		t.Errorf("bot.User.Username = %q, want %q", bot.User.Username, "testbot")
	}
}

// TestNewBot_DefaultTelegramAPIURL verifies that when TELEGRAM_API_URL is empty,
// the bot attempts to use the default Telegram API (which will fail in tests
// without network, but that's expected behavior).
func TestNewBot_DefaultTelegramAPIURL_RequiresBotToken(t *testing.T) {
	t.Setenv("BOT_TOKEN", "")
	t.Setenv("DEBUG", "false")
	t.Setenv("TELEGRAM_API_URL", "")

	_, _, _, err := NewBot()
	if err == nil {
		t.Fatal("expected error when BOT_TOKEN is empty")
	}
}
