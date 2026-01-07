package callback_test

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/handlers/testutil"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

func TestIntegration_ChainCallback_SendsMultipleMessages(t *testing.T) {
	parentID := "111111"
	grandparentID := "000000"
	fakeAPI := &testutil.FakeTweetAPI{
		Tweets: map[string]*twitterxapi.Tweet{
			"user/222222": {
				ID:               "222222",
				URL:              "https://x.com/user/status/222222",
				Text:             "Reply to parent",
				ReplyingToStatus: &parentID,
				ReplyingTo:       testutil.StrPtr("user"),
				Author:           twitterxapi.Author{Name: "User", ScreenName: "user"},
			},
			"user/111111": {
				ID:               "111111",
				URL:              "https://x.com/user/status/111111",
				Text:             "Parent tweet",
				ReplyingToStatus: &grandparentID,
				ReplyingTo:       testutil.StrPtr("user"),
				Author:           twitterxapi.Author{Name: "User", ScreenName: "user"},
			},
			"user/000000": {
				ID:     "000000",
				URL:    "https://x.com/user/status/000000",
				Text:   "Original grandparent tweet",
				Author: twitterxapi.Author{Name: "User", ScreenName: "user"},
			},
		},
	}

	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, fakeAPI)

	const (
		chatID = int64(888888)
		msgID  = int64(400)
	)

	callbackData := tweet.EncodeChainCallback("user", "222222", msgID)
	callbackMessage := &gotgbot.Message{
		MessageId: 401,
		Date:      1000003,
		Chat:      gotgbot.Chat{Id: chatID, Type: "private"},
	}

	update := gotgbot.Update{
		UpdateId: 4,
		CallbackQuery: &gotgbot.CallbackQuery{
			Id:      "cb123",
			Data:    callbackData,
			From:    gotgbot.User{Id: 1004, FirstName: "Dave", Username: "dave"},
			Message: callbackMessage,
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	if len(mock.GetCalls("answerCallbackQuery")) < 1 {
		t.Errorf("answerCallbackQuery calls = %d, want >= 1", len(mock.GetCalls("answerCallbackQuery")))
	}

	msgCalls := mock.GetCalls("sendMessage")
	if len(msgCalls) != 3 {
		t.Fatalf("sendMessage calls = %d, want 3 (for 3-tweet chain)", len(msgCalls))
	}

	for i, call := range msgCalls {
		if gotChatID, ok := call.JSONInt64("chat_id"); !ok || gotChatID != chatID {
			t.Errorf("message %d: chat_id = %d, want %d", i, gotChatID, chatID)
		}
	}

	if replyMsgID, ok := msgCalls[0].JSONInt64("reply_parameters.message_id"); !ok || replyMsgID != msgID {
		t.Errorf("first chain message should reply to original message %d, got %d", msgID, replyMsgID)
	}

	if len(mock.GetCalls("deleteMessage")) != 1 {
		t.Errorf("deleteMessage calls = %d, want 1", len(mock.GetCalls("deleteMessage")))
	}
}

func TestIntegration_DeleteCallback_WithChainButton(t *testing.T) {
	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, &testutil.FakeTweetAPI{})

	const (
		chatID        = int64(123123)
		originalMsgID = int64(321321)
	)

	callbackData := tweet.EncodeDeleteCallback(originalMsgID, &tweet.KeyboardOpts{
		ShowChainButton: true,
		ChainUsername:   "chainuser",
		ChainTweetID:    "chain123",
	})

	update := gotgbot.Update{
		UpdateId: 8,
		CallbackQuery: &gotgbot.CallbackQuery{
			Id:   "cb-delete-chain",
			Data: callbackData,
			From: gotgbot.User{Id: 2020, FirstName: "Del"},
			Message: &gotgbot.Message{
				MessageId: 777,
				Chat:      gotgbot.Chat{Id: chatID, Type: "private"},
			},
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	deleteCalls := mock.GetCalls("deleteMessage")
	if len(deleteCalls) != 1 {
		t.Fatalf("deleteMessage calls = %d, want 1", len(deleteCalls))
	}
	if gotChatID, ok := deleteCalls[0].JSONInt64("chat_id"); !ok || gotChatID != chatID {
		t.Fatalf("deleteMessage chat_id = %d, want %d", gotChatID, chatID)
	}
	if gotMsgID, ok := deleteCalls[0].JSONInt64("message_id"); !ok || gotMsgID != originalMsgID {
		t.Fatalf("deleteMessage message_id = %d, want %d", gotMsgID, originalMsgID)
	}

	editCalls := mock.GetCalls("editMessageReplyMarkup")
	if len(editCalls) != 1 {
		t.Fatalf("editMessageReplyMarkup calls = %d, want 1", len(editCalls))
	}
	markup, ok := editCalls[0].JSONString("reply_markup")
	if !ok {
		t.Fatalf("editMessageReplyMarkup missing reply_markup: %v", editCalls[0].JSON)
	}
	if !testutil.ContainsString(markup, tweet.ChainCallbackPrefix) {
		t.Fatalf("reply_markup should keep chain button, got: %s", markup)
	}

	answerCalls := mock.GetCalls("answerCallbackQuery")
	if len(answerCalls) != 1 {
		t.Fatalf("answerCallbackQuery calls = %d, want 1", len(answerCalls))
	}
	if text, ok := answerCalls[0].JSONString("text"); !ok || text != "Deleted" {
		t.Fatalf("answerCallbackQuery text = %q, want %q", text, "Deleted")
	}
}

func TestIntegration_DeleteCallback_RemovesKeyboardWithoutChain(t *testing.T) {
	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, &testutil.FakeTweetAPI{})

	const (
		chatID        = int64(456456)
		originalMsgID = int64(654654)
	)

	callbackData := tweet.EncodeDeleteCallback(originalMsgID, nil)

	update := gotgbot.Update{
		UpdateId: 9,
		CallbackQuery: &gotgbot.CallbackQuery{
			Id:   "cb-delete-nochain",
			Data: callbackData,
			From: gotgbot.User{Id: 2021, FirstName: "Del"},
			Message: &gotgbot.Message{
				MessageId: 888,
				Chat:      gotgbot.Chat{Id: chatID, Type: "private"},
			},
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	deleteCalls := mock.GetCalls("deleteMessage")
	if len(deleteCalls) != 1 {
		t.Fatalf("deleteMessage calls = %d, want 1", len(deleteCalls))
	}

	editCalls := mock.GetCalls("editMessageReplyMarkup")
	if len(editCalls) != 1 {
		t.Fatalf("editMessageReplyMarkup calls = %d, want 1", len(editCalls))
	}
	markup, ok := editCalls[0].JSONString("reply_markup")
	if !ok {
		t.Fatalf("editMessageReplyMarkup missing reply_markup: %v", editCalls[0].JSON)
	}
	if testutil.ContainsString(markup, tweet.ChainCallbackPrefix) {
		t.Fatalf("reply_markup should not contain chain prefix, got: %s", markup)
	}
}
