package sendtweet

import (
	"context"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
	testtelegram "twitterx-bot/pkg/testutil/telegram"
)

// This is an integration-style test:
// - sendtweet.UseCase + tweet.Sender use a real gotgbot.Bot instance
// - gotgbot serializes to HTTP and sends requests to a local httptest-based mock server
// - test asserts outgoing HTTP payloads recorded by the mock
func TestUseCaseSendTweet_UsesTelegramHTTPMock(t *testing.T) {
	mock := testtelegram.NewMockServer()
	t.Cleanup(mock.Close)

	bot, err := gotgbot.NewBot("123:ABC", &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{APIURL: mock.URL()},
		},
	})
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "777",
			URL:  "https://x.com/user/status/777",
			Text: "Hello from integration",
			Author: twitterxapi.Author{
				Name:       "User",
				ScreenName: "user",
			},
		},
	}

	uc := New(fetcher, tweet.Sender{Bot: bot})

	const (
		chatID       int64 = 424242
		replyToMsgID int64 = 101
	)

	if err := uc.SendTweet(context.Background(), chatID, replyToMsgID, "user", "777", "@req"); err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}

	calls := mock.GetCalls("sendMessage")
	if len(calls) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(calls))
	}

	call := calls[0]
	if call.JSON == nil {
		t.Fatalf("expected JSON payload to be decoded; raw=%q", string(call.RawBody))
	}

	// Assert key fields (using minimal checks to avoid coupling to gotgbot internals).
	if gotChatID, ok := call.JSONInt64("chat_id"); !ok || gotChatID != chatID {
		t.Fatalf("chat_id = (%d, %v), want (%d, true). full json=%v", gotChatID, ok, chatID, call.JSON)
	}
	if gotReplyID, ok := call.JSONInt64("reply_parameters.message_id"); !ok || gotReplyID != replyToMsgID {
		t.Fatalf("reply_parameters.message_id = (%d, %v), want (%d, true). full json=%v", gotReplyID, ok, replyToMsgID, call.JSON)
	}
	if gotParseMode, ok := call.JSONString("parse_mode"); !ok || gotParseMode != "HTML" {
		t.Fatalf("parse_mode = (%q, %v), want (%q, true). full json=%v", gotParseMode, ok, "HTML", call.JSON)
	}
	if gotText, ok := call.JSONString("text"); !ok || gotText == "" {
		t.Fatalf("text = (%q, %v), want non-empty. full json=%v", gotText, ok, call.JSON)
	}
}
