package start_test

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/handlers/testutil"
)

func TestIntegration_StartCommand_RepliesWithGreeting(t *testing.T) {
	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, &testutil.FakeTweetAPI{})

	const (
		chatID = int64(111222)
		msgID  = int64(999)
	)

	update := gotgbot.Update{
		UpdateId: 10,
		Message: &gotgbot.Message{
			MessageId: msgID,
			Text:      "/start",
			Chat:      gotgbot.Chat{Id: chatID, Type: "private"},
			From:      &gotgbot.User{Id: 3003, FirstName: "Greet"},
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	sendCalls := mock.GetCalls("sendMessage")
	if len(sendCalls) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(sendCalls))
	}
	call := sendCalls[0]
	text, ok := call.JSONString("text")
	if !ok {
		t.Fatalf("sendMessage missing text: %v", call.JSON)
	}
	if text != "Hi! Send me any text and I will echo it back." {
		t.Fatalf("start message = %q, want greeting", text)
	}
	if replyID, ok := call.JSONInt64("reply_parameters.message_id"); !ok || replyID != msgID {
		t.Fatalf("reply_parameters.message_id = %d, want %d", replyID, msgID)
	}
}
