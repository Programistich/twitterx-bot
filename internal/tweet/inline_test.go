package tweet

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/twitterxapi"
)

func TestBuildInlineResult(t *testing.T) {
	tests := []struct {
		name       string
		tweet      *twitterxapi.Tweet
		fallbackID string
		wantOK     bool
		assert     func(t *testing.T, result gotgbot.InlineQueryResult)
	}{
		{
			name:       "nil tweet",
			tweet:      nil,
			fallbackID: "",
			wantOK:     false,
			assert: func(t *testing.T, result gotgbot.InlineQueryResult) {
				if result != nil {
					t.Fatalf("expected nil result, got %#v", result)
				}
			},
		},
		{
			name: "video result",
			tweet: &twitterxapi.Tweet{
				ID:   " 123 ",
				Text: "  Hello  ",
				URL:  "https://x.com/1",
				Author: twitterxapi.Author{
					ScreenName: "alice",
				},
				Media: &twitterxapi.Media{
					Videos: []twitterxapi.Video{
						{
							URL:          "https://video/1.mp4",
							ThumbnailURL: "https://video/1.jpg",
							Width:        1280,
							Height:       720,
							Format:       "mp4",
						},
					},
				},
			},
			wantOK: true,
			assert: func(t *testing.T, result gotgbot.InlineQueryResult) {
				video, ok := result.(gotgbot.InlineQueryResultVideo)
				if !ok {
					t.Fatalf("expected video result, got %#v", result)
				}
				if video.Id != "123:video" {
					t.Errorf("video.Id = %q, want %q", video.Id, "123:video")
				}
				if video.Title != "Tweet by @alice" {
					t.Errorf("video.Title = %q, want %q", video.Title, "Tweet by @alice")
				}
				if video.VideoUrl != "https://video/1.mp4" {
					t.Errorf("video.VideoUrl = %q, want %q", video.VideoUrl, "https://video/1.mp4")
				}
				if video.ThumbnailUrl != "https://video/1.jpg" {
					t.Errorf("video.ThumbnailUrl = %q, want %q", video.ThumbnailUrl, "https://video/1.jpg")
				}
				if video.MimeType != "video/mp4" {
					t.Errorf("video.MimeType = %q, want %q", video.MimeType, "video/mp4")
				}
				if video.Caption != "Hello\n\nhttps://x.com/1" {
					t.Errorf("video.Caption = %q, want %q", video.Caption, "Hello\n\nhttps://x.com/1")
				}
				if video.Description != "Hello" {
					t.Errorf("video.Description = %q, want %q", video.Description, "Hello")
				}
				if video.VideoWidth != 1280 || video.VideoHeight != 720 {
					t.Errorf("video dimensions = (%d, %d), want (1280, 720)", video.VideoWidth, video.VideoHeight)
				}
			},
		},
		{
			name: "photo result with mosaic",
			tweet: &twitterxapi.Tweet{
				ID:   "",
				Text: "Photo",
				Author: twitterxapi.Author{
					Name: "Bob",
				},
				Media: &twitterxapi.Media{
					Photos: []twitterxapi.Photo{
						{URL: "https://img/1.jpg", Width: 640, Height: 480},
						{URL: "https://img/2.jpg", Width: 800, Height: 600},
					},
					Mosaic: &twitterxapi.Mosaic{
						Formats: map[string]string{"jpeg": "https://img/mosaic.jpg"},
						Width:   intPtr(1200),
						Height:  intPtr(800),
					},
				},
			},
			fallbackID: "fallback",
			wantOK:     true,
			assert: func(t *testing.T, result gotgbot.InlineQueryResult) {
				photo, ok := result.(gotgbot.InlineQueryResultPhoto)
				if !ok {
					t.Fatalf("expected photo result, got %#v", result)
				}
				if photo.Id != "fallback:photo" {
					t.Errorf("photo.Id = %q, want %q", photo.Id, "fallback:photo")
				}
				if photo.PhotoUrl != "https://img/mosaic.jpg" {
					t.Errorf("photo.PhotoUrl = %q, want %q", photo.PhotoUrl, "https://img/mosaic.jpg")
				}
				if photo.ThumbnailUrl != "https://img/mosaic.jpg" {
					t.Errorf("photo.ThumbnailUrl = %q, want %q", photo.ThumbnailUrl, "https://img/mosaic.jpg")
				}
				if photo.PhotoWidth != 1200 || photo.PhotoHeight != 800 {
					t.Errorf("photo dimensions = (%d, %d), want (1200, 800)", photo.PhotoWidth, photo.PhotoHeight)
				}
				if photo.Title != "Tweet by Bob" {
					t.Errorf("photo.Title = %q, want %q", photo.Title, "Tweet by Bob")
				}
				if photo.Description != "Photo" {
					t.Errorf("photo.Description = %q, want %q", photo.Description, "Photo")
				}
				if photo.Caption != "Photo" {
					t.Errorf("photo.Caption = %q, want %q", photo.Caption, "Photo")
				}
			},
		},
		{
			name: "article fallback id",
			tweet: &twitterxapi.Tweet{
				ID:   " ",
				Text: "Hello",
				URL:  "https://x.com/1",
				Author: twitterxapi.Author{
					ScreenName: "user",
					AvatarURL:  "https://avatar/1.jpg",
				},
			},
			wantOK: true,
			assert: func(t *testing.T, result gotgbot.InlineQueryResult) {
				article, ok := result.(gotgbot.InlineQueryResultArticle)
				if !ok {
					t.Fatalf("expected article result, got %#v", result)
				}
				if article.Id != "tweet:text" {
					t.Errorf("article.Id = %q, want %q", article.Id, "tweet:text")
				}
				if article.Title != "Tweet by @user" {
					t.Errorf("article.Title = %q, want %q", article.Title, "Tweet by @user")
				}
				textContent, ok := article.InputMessageContent.(gotgbot.InputTextMessageContent)
				if !ok {
					t.Fatalf("expected InputTextMessageContent, got %#v", article.InputMessageContent)
				}
				if textContent.MessageText != "Hello\n\nhttps://x.com/1" {
					t.Errorf("article.MessageText = %q, want %q", textContent.MessageText, "Hello\n\nhttps://x.com/1")
				}
				if article.Url != "https://x.com/1" {
					t.Errorf("article.Url = %q, want %q", article.Url, "https://x.com/1")
				}
				if article.Description != "Hello" {
					t.Errorf("article.Description = %q, want %q", article.Description, "Hello")
				}
				if article.ThumbnailUrl != "https://avatar/1.jpg" {
					t.Errorf("article.ThumbnailUrl = %q, want %q", article.ThumbnailUrl, "https://avatar/1.jpg")
				}
			},
		},
		{
			name: "empty message yields no result",
			tweet: &twitterxapi.Tweet{
				Text:   "   ",
				URL:    " ",
				Author: twitterxapi.Author{ScreenName: "user"},
			},
			wantOK: false,
			assert: func(t *testing.T, result gotgbot.InlineQueryResult) {
				if result != nil {
					t.Fatalf("expected nil result, got %#v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotOK := BuildInlineResult(tt.tweet, tt.fallbackID)
			if gotOK != tt.wantOK {
				t.Fatalf("BuildInlineResult() ok = %v, want %v", gotOK, tt.wantOK)
			}
			tt.assert(t, gotResult)
		})
	}
}
