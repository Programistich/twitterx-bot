package telegram

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestMockServer_RecordsCalls(t *testing.T) {
	ms := NewMockServer()
	t.Cleanup(ms.Close)

	bot, err := gotgbot.NewBot("123:ABC", &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{APIURL: ms.URL()},
		},
	})
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	_, err = bot.SendMessage(42, "hello", nil)
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	calls := ms.GetCalls("sendMessage")
	if len(calls) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(calls))
	}
	if gotChatID, ok := calls[0].JSONInt64("chat_id"); !ok || gotChatID != 42 {
		t.Fatalf("chat_id = (%d, %v), want (42, true). json=%v", gotChatID, ok, calls[0].JSON)
	}
	if gotText, ok := calls[0].JSONString("text"); !ok || gotText != "hello" {
		t.Fatalf("text = (%q, %v), want (%q, true). json=%v", gotText, ok, "hello", calls[0].JSON)
	}
}

func TestMockServer_MethodResponseOverride(t *testing.T) {
	ms := NewMockServer()
	t.Cleanup(ms.Close)

	// Override sendMessage result with a known message_id.
	err := ms.SetResponse("sendMessage", map[string]any{
		"message_id": 999,
		"date":       0,
		"chat": map[string]any{
			"id":   42,
			"type": "private",
		},
		"text": "overridden",
	})
	if err != nil {
		t.Fatalf("SetResponse() error = %v", err)
	}

	bot, err := gotgbot.NewBot("123:ABC", &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{APIURL: ms.URL()},
		},
	})
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	msg, err := bot.SendMessage(42, "hello", nil)
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if msg == nil || msg.MessageId != 999 {
		t.Fatalf("message_id = %v, want 999", msg)
	}
}

func TestMockServer_TelegramError(t *testing.T) {
	ms := NewMockServer()
	t.Cleanup(ms.Close)

	ms.SetTelegramError("sendMessage", 400, "bad request (mock)")

	bot, err := gotgbot.NewBot("123:ABC", &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{APIURL: ms.URL()},
		},
	})
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	_, err = bot.SendMessage(42, "hello", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMockServer_HTTPError(t *testing.T) {
	ms := NewMockServer()
	t.Cleanup(ms.Close)

	ms.SetHTTPError("sendMessage", 500, "server error (mock)")

	bot, err := gotgbot.NewBot("123:ABC", &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{APIURL: ms.URL()},
		},
	})
	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	_, err = bot.SendMessage(42, "hello", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}
