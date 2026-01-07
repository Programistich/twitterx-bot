// Package testutil provides shared test helpers for handler integration tests.
package testutil

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/handlers"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/twitterxapi"
	testtelegram "twitterx-bot/pkg/testutil/telegram"
)

// FakeTweetAPI is a minimal fake for testing that returns configured tweets.
type FakeTweetAPI struct {
	Tweets map[string]*twitterxapi.Tweet
}

// GetTweet returns a tweet from the configured map or nil if not found.
func (f *FakeTweetAPI) GetTweet(_ context.Context, username, tweetID string) (*twitterxapi.Tweet, error) {
	key := username + "/" + tweetID
	if tw, ok := f.Tweets[key]; ok {
		return tw, nil
	}
	return nil, nil
}

// NewTestBot creates a gotgbot.Bot that points at the provided mock server.
func NewTestBot(t *testing.T, mock *testtelegram.MockServer) *gotgbot.Bot {
	t.Helper()
	bot, err := gotgbot.NewBot("123:ABC", &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{APIURL: mock.URL()},
		},
	})
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}
	return bot
}

// SetupBotAndDispatcher wires the handlers to a dispatcher and returns the mock bot, mock server, and dispatcher.
func SetupBotAndDispatcher(t *testing.T, fakeAPI *FakeTweetAPI) (*gotgbot.Bot, *testtelegram.MockServer, *ext.Dispatcher) {
	t.Helper()
	if fakeAPI == nil {
		fakeAPI = &FakeTweetAPI{}
	}

	mock := testtelegram.NewMockServer()
	t.Cleanup(mock.Close)

	bot := NewTestBot(t, mock)

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(_ *gotgbot.Bot, _ *ext.Context, err error) ext.DispatcherAction {
			t.Errorf("dispatcher error: %v", err)
			return ext.DispatcherActionNoop
		},
	})

	handlers.RegisterWithFetcher(dispatcher, logger.New(true), fakeAPI)

	return bot, mock, dispatcher
}

// DecodeInlineResults decodes the inline query results from a mock call.
func DecodeInlineResults(t *testing.T, call testtelegram.Call) []any {
	t.Helper()
	if call.JSON == nil {
		t.Fatalf("call JSON missing for %s", call.Method)
	}
	raw, ok := call.JSON["results"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case string:
		var decoded []any
		if err := json.Unmarshal([]byte(v), &decoded); err != nil {
			t.Fatalf("failed to decode inline results: %v", err)
		}
		return decoded
	case []any:
		return v
	default:
		t.Fatalf("unexpected results type %T", raw)
		return nil
	}
}

// StrPtr returns a pointer to the given string.
func StrPtr(s string) *string {
	return &s
}

// ContainsString checks if needle is a substring of haystack.
func ContainsString(haystack, needle string) bool {
	if len(haystack) == 0 || len(needle) == 0 {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
