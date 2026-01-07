package tweetsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/twitterxapi"
)

func TestServiceBuildInlineResultVideo(t *testing.T) {
	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "123",
			Text: "video",
			URL:  "https://x.com/user/status/123",
			Media: &twitterxapi.Media{
				Videos: []twitterxapi.Video{{
					URL:          "https://video/1.mp4",
					ThumbnailURL: "https://video/1.jpg",
					Width:        640,
					Height:       480,
				}},
			},
		},
	}
	svc := New(fetcher, nil)

	result, ok, err := svc.BuildInlineResult(context.Background(), "user", "123")
	if err != nil {
		t.Fatalf("BuildInlineResult() error = %v", err)
	}
	if !ok {
		t.Fatalf("BuildInlineResult() ok = false, want true")
	}
	if _, ok := result.(gotgbot.InlineQueryResultVideo); !ok {
		t.Fatalf("result type = %T, want InlineQueryResultVideo", result)
	}
}

func TestServiceBuildInlineResultPhoto(t *testing.T) {
	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "321",
			Text: "photo",
			URL:  "https://x.com/user/status/321",
			Media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{{URL: "https://img/1.jpg", Width: 640, Height: 480}},
			},
		},
	}
	svc := New(fetcher, nil)

	result, ok, err := svc.BuildInlineResult(context.Background(), "user", "321")
	if err != nil {
		t.Fatalf("BuildInlineResult() error = %v", err)
	}
	if !ok {
		t.Fatalf("BuildInlineResult() ok = false, want true")
	}
	if _, ok := result.(gotgbot.InlineQueryResultPhoto); !ok {
		t.Fatalf("result type = %T, want InlineQueryResultPhoto", result)
	}
}

func TestServiceBuildInlineResultText(t *testing.T) {
	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "777",
			Text: "just text",
			URL:  "https://x.com/user/status/777",
			Author: twitterxapi.Author{
				Name:       "User",
				ScreenName: "user",
			},
		},
	}
	svc := New(fetcher, nil)

	result, ok, err := svc.BuildInlineResult(context.Background(), "user", "777")
	if err != nil {
		t.Fatalf("BuildInlineResult() error = %v", err)
	}
	if !ok {
		t.Fatalf("BuildInlineResult() ok = false, want true")
	}
	if _, ok := result.(gotgbot.InlineQueryResultArticle); !ok {
		t.Fatalf("result type = %T, want InlineQueryResultArticle", result)
	}
}

func TestServiceBuildInlineResultEmptyTweet(t *testing.T) {
	fetcher := &fakeFetcher{}
	svc := New(fetcher, nil)

	result, ok, err := svc.BuildInlineResult(context.Background(), "user", "999")
	if err != nil {
		t.Fatalf("BuildInlineResult() error = %v", err)
	}
	if ok {
		t.Fatalf("BuildInlineResult() ok = true, want false")
	}
	if result != nil {
		t.Fatalf("result = %v, want nil", result)
	}
}

func TestServiceBuildInlineResultFetchError(t *testing.T) {
	fetcher := &fakeFetcher{err: errors.New("boom")}
	svc := New(fetcher, nil)

	_, ok, err := svc.BuildInlineResult(context.Background(), "user", "123")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrFetchTweet) {
		t.Fatalf("expected ErrFetchTweet, got %v", err)
	}
	if ok {
		t.Fatalf("ok = true, want false")
	}
}
