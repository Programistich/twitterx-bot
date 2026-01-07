package message_test

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/handlers/testutil"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

func TestIntegration_MessageHandler_SendsTweetViaMock(t *testing.T) {
	fakeAPI := &testutil.FakeTweetAPI{
		Tweets: map[string]*twitterxapi.Tweet{
			"testuser/123456789": {
				ID:   "123456789",
				URL:  "https://x.com/testuser/status/123456789",
				Text: "Hello integration test!",
				Author: twitterxapi.Author{
					Name:       "Test User",
					ScreenName: "testuser",
				},
			},
		},
	}

	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, fakeAPI)

	const (
		chatID = int64(424242)
		msgID  = int64(100)
	)

	update := gotgbot.Update{
		UpdateId: 1,
		Message: &gotgbot.Message{
			MessageId: msgID,
			Text:      "https://x.com/testuser/status/123456789",
			Chat: gotgbot.Chat{
				Id:   chatID,
				Type: "private",
			},
			From: &gotgbot.User{
				Id:        1001,
				FirstName: "Alice",
				Username:  "alice",
			},
			Date: 1000000,
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	actionCalls := mock.GetCalls("sendChatAction")
	if len(actionCalls) != 1 {
		t.Errorf("sendChatAction calls = %d, want 1", len(actionCalls))
	} else {
		if gotChatID, ok := actionCalls[0].JSONInt64("chat_id"); !ok || gotChatID != chatID {
			t.Errorf("sendChatAction chat_id = %d, want %d", gotChatID, chatID)
		}
		if action, ok := actionCalls[0].JSONString("action"); !ok || action != "typing" {
			t.Errorf("sendChatAction action = %q, want %q", action, "typing")
		}
	}

	msgCalls := mock.GetCalls("sendMessage")
	if len(msgCalls) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(msgCalls))
	}

	call := msgCalls[0]
	if call.JSON == nil {
		t.Fatalf("expected JSON payload, got nil; raw=%s", string(call.RawBody))
	}

	if gotChatID, ok := call.JSONInt64("chat_id"); !ok || gotChatID != chatID {
		t.Errorf("sendMessage chat_id = %d, want %d", gotChatID, chatID)
	}
	if replyMsgID, ok := call.JSONInt64("reply_parameters.message_id"); !ok || replyMsgID != msgID {
		t.Errorf("sendMessage reply_parameters.message_id = %d, want %d", replyMsgID, msgID)
	}
	if parseMode, ok := call.JSONString("parse_mode"); !ok || parseMode != "HTML" {
		t.Errorf("sendMessage parse_mode = %q, want HTML", parseMode)
	}
	if text, ok := call.JSONString("text"); !ok || text == "" {
		t.Errorf("sendMessage text is empty")
	}
}

func TestIntegration_MessageHandler_TweetWithVideo(t *testing.T) {
	fakeAPI := &testutil.FakeTweetAPI{
		Tweets: map[string]*twitterxapi.Tweet{
			"videouser/987654321": {
				ID:   "987654321",
				URL:  "https://x.com/videouser/status/987654321",
				Text: "Check out this video!",
				Author: twitterxapi.Author{
					Name:       "Video User",
					ScreenName: "videouser",
				},
				Media: &twitterxapi.Media{
					Videos: []twitterxapi.Video{
						{URL: "https://video.twimg.com/test.mp4", Width: 1280, Height: 720},
					},
				},
			},
		},
	}

	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, fakeAPI)

	const (
		chatID = int64(555555)
		msgID  = int64(200)
	)

	update := gotgbot.Update{
		UpdateId: 2,
		Message: &gotgbot.Message{
			MessageId: msgID,
			Text:      "https://x.com/videouser/status/987654321",
			Chat:      gotgbot.Chat{Id: chatID, Type: "private"},
			From:      &gotgbot.User{Id: 1002, FirstName: "Bob", Username: "bob"},
			Date:      1000001,
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	videoCalls := mock.GetCalls("sendVideo")
	if len(videoCalls) != 1 {
		t.Fatalf("sendVideo calls = %d, want 1", len(videoCalls))
	}

	call := videoCalls[0]
	if gotChatID, ok := call.JSONInt64("chat_id"); !ok || gotChatID != chatID {
		t.Errorf("sendVideo chat_id = %d, want %d", gotChatID, chatID)
	}
	if replyMsgID, ok := call.JSONInt64("reply_parameters.message_id"); !ok || replyMsgID != msgID {
		t.Errorf("sendVideo reply_parameters.message_id = %d, want %d", replyMsgID, msgID)
	}

	if len(mock.GetCalls("sendMessage")) != 0 {
		t.Errorf("sendMessage calls = %d, want 0 (should use sendVideo instead)", len(mock.GetCalls("sendMessage")))
	}
}

func TestIntegration_MessageHandler_TweetWithReplyChain(t *testing.T) {
	parentTweetID := "111111"
	fakeAPI := &testutil.FakeTweetAPI{
		Tweets: map[string]*twitterxapi.Tweet{
			"replyuser/222222": {
				ID:               "222222",
				URL:              "https://x.com/replyuser/status/222222",
				Text:             "This is a reply!",
				ReplyingToStatus: &parentTweetID,
				ReplyingTo:       testutil.StrPtr("originaluser"),
				Author: twitterxapi.Author{
					Name:       "Reply User",
					ScreenName: "replyuser",
				},
			},
		},
	}

	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, fakeAPI)

	const (
		chatID = int64(777777)
		msgID  = int64(300)
	)

	update := gotgbot.Update{
		UpdateId: 3,
		Message: &gotgbot.Message{
			MessageId: msgID,
			Text:      "https://x.com/replyuser/status/222222",
			Chat:      gotgbot.Chat{Id: chatID, Type: "private"},
			From:      &gotgbot.User{Id: 1003, FirstName: "Carol", Username: "carol"},
			Date:      1000002,
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	msgCalls := mock.GetCalls("sendMessage")
	if len(msgCalls) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(msgCalls))
	}

	call := msgCalls[0]
	if call.JSON == nil {
		t.Fatalf("expected JSON payload")
	}

	replyMarkup, ok := call.JSON["reply_markup"]
	if !ok {
		t.Fatalf("sendMessage missing reply_markup")
	}

	if replyMarkupStr, isString := replyMarkup.(string); isString && replyMarkupStr != "" {
		if !testutil.ContainsString(replyMarkupStr, tweet.ChainCallbackPrefix) {
			t.Errorf("reply_markup should contain chain button with prefix %q, got: %s", tweet.ChainCallbackPrefix, replyMarkupStr)
		}
	}
}

func TestIntegration_NonTwitterURL_NoResponse(t *testing.T) {
	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, &testutil.FakeTweetAPI{Tweets: map[string]*twitterxapi.Tweet{}})

	update := gotgbot.Update{
		UpdateId: 5,
		Message: &gotgbot.Message{
			MessageId: 500,
			Text:      "Hello! This is just a regular message without any URL.",
			Chat:      gotgbot.Chat{Id: 999999, Type: "private"},
			From:      &gotgbot.User{Id: 1005, FirstName: "Eve", Username: "eve"},
			Date:      1000004,
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	if len(mock.GetCalls("sendMessage")) != 0 {
		t.Errorf("sendMessage should not be called for non-Twitter URL")
	}
	if len(mock.GetCalls("sendPhoto")) != 0 {
		t.Errorf("sendPhoto should not be called for non-Twitter URL")
	}
	if len(mock.GetCalls("sendVideo")) != 0 {
		t.Errorf("sendVideo should not be called for non-Twitter URL")
	}
	if len(mock.GetCalls("sendChatAction")) != 0 {
		t.Errorf("sendChatAction should not be called for non-Twitter URL")
	}
}
