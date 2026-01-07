package inline_test

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/handlers/testutil"
	"twitterx-bot/internal/twitterxapi"
)

func TestIntegration_InlineQuery_ReturnsResult(t *testing.T) {
	fakeAPI := &testutil.FakeTweetAPI{
		Tweets: map[string]*twitterxapi.Tweet{
			"inlineuser/987": {
				ID:   "987",
				URL:  "https://x.com/inlineuser/status/987",
				Text: "Inline response",
				Author: twitterxapi.Author{
					Name:       "Inline",
					ScreenName: "inlineuser",
				},
			},
		},
	}

	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, fakeAPI)

	update := gotgbot.Update{
		UpdateId: 6,
		InlineQuery: &gotgbot.InlineQuery{
			Id:    "inline-ok",
			Query: "https://x.com/inlineuser/status/987",
			From:  gotgbot.User{Id: 2001, FirstName: "Inline", Username: "inlineuser"},
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	calls := mock.GetCalls("answerInlineQuery")
	if len(calls) != 1 {
		t.Fatalf("answerInlineQuery calls = %d, want 1", len(calls))
	}

	call := calls[0]
	results := testutil.DecodeInlineResults(t, call)
	if len(results) != 1 {
		t.Fatalf("results len = %d, want 1", len(results))
	}
	resultObj, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("inline result type = %T, want map", results[0])
	}
	if _, ok := resultObj["id"]; !ok {
		t.Fatalf("inline result missing id: %v", resultObj)
	}
	if gotInlineID, ok := call.JSONString("inline_query_id"); !ok || gotInlineID != "inline-ok" {
		t.Fatalf("inline_query_id = %q, want %q", gotInlineID, "inline-ok")
	}
}

func TestIntegration_InlineQuery_InvalidQuery(t *testing.T) {
	bot, mock, dispatcher := testutil.SetupBotAndDispatcher(t, &testutil.FakeTweetAPI{})

	update := gotgbot.Update{
		UpdateId: 7,
		InlineQuery: &gotgbot.InlineQuery{
			Id:    "inline-empty",
			Query: "not a url",
			From:  gotgbot.User{Id: 2002, FirstName: "Empty"},
		},
	}

	if err := dispatcher.ProcessUpdate(bot, &update, nil); err != nil {
		t.Fatalf("ProcessUpdate() error = %v", err)
	}

	calls := mock.GetCalls("answerInlineQuery")
	if len(calls) != 1 {
		t.Fatalf("answerInlineQuery calls = %d, want 1", len(calls))
	}

	call := calls[0]
	results := testutil.DecodeInlineResults(t, call)
	if len(results) != 0 {
		t.Fatalf("results len = %d, want 0", len(results))
	}
}
