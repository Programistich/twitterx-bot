package handlers

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/twitterxapi"
)

func TestBuildTitle(t *testing.T) {
	tests := []struct {
		name  string
		tweet *twitterxapi.Tweet
		want  string
	}{
		{
			name:  "screen name preferred",
			tweet: &twitterxapi.Tweet{Author: twitterxapi.Author{ScreenName: "alice", Name: "Alice"}},
			want:  "Tweet by @alice",
		},
		{
			name:  "falls back to name",
			tweet: &twitterxapi.Tweet{Author: twitterxapi.Author{Name: "Alice"}},
			want:  "Tweet by Alice",
		},
		{
			name:  "default title",
			tweet: &twitterxapi.Tweet{},
			want:  "Tweet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTitle(tt.tweet); got != tt.want {
				t.Errorf("buildTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTweetContent(t *testing.T) {
	tests := []struct {
		name  string
		tweet *twitterxapi.Tweet
		want  string
	}{
		{
			name:  "nil tweet",
			tweet: nil,
			want:  "",
		},
		{
			name:  "text only",
			tweet: &twitterxapi.Tweet{Text: "  hello  "},
			want:  "hello",
		},
		{
			name:  "url only",
			tweet: &twitterxapi.Tweet{URL: "  https://x.com/1  "},
			want:  "https://x.com/1",
		},
		{
			name:  "text and url",
			tweet: &twitterxapi.Tweet{Text: " hello ", URL: " https://x.com/2 "},
			want:  "hello\n\nhttps://x.com/2",
		},
		{
			name:  "blank text uses url",
			tweet: &twitterxapi.Tweet{Text: "   ", URL: "https://x.com/3"},
			want:  "https://x.com/3",
		},
		{
			name:  "blank url uses text",
			tweet: &twitterxapi.Tweet{Text: "hello", URL: "   "},
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tweetContent(tt.tweet); got != tt.want {
				t.Errorf("tweetContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTweetCaptionAndMessageText(t *testing.T) {
	tweet := &twitterxapi.Tweet{
		Text: "hello",
		URL:  "https://x.com/1",
	}
	want := "hello\n\nhttps://x.com/1"

	if got := tweetCaption(tweet); got != want {
		t.Errorf("tweetCaption() = %q, want %q", got, want)
	}
	if got := tweetMessageText(tweet); got != want {
		t.Errorf("tweetMessageText() = %q, want %q", got, want)
	}
}

func TestSelectPhoto(t *testing.T) {
	tests := []struct {
		name       string
		media      *twitterxapi.Media
		wantURL    string
		wantThumb  string
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "nil media",
			media:      nil,
			wantURL:    "",
			wantThumb:  "",
			wantWidth:  0,
			wantHeight: 0,
		},
		{
			name:       "no photos",
			media:      &twitterxapi.Media{},
			wantURL:    "",
			wantThumb:  "",
			wantWidth:  0,
			wantHeight: 0,
		},
		{
			name: "single photo",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/1.jpg", Width: 640, Height: 480},
				},
			},
			wantURL:    "https://img/1.jpg",
			wantThumb:  "https://img/1.jpg",
			wantWidth:  640,
			wantHeight: 480,
		},
		{
			name: "multiple photos with mosaic",
			media: &twitterxapi.Media{
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
			wantURL:    "https://img/mosaic.jpg",
			wantThumb:  "https://img/mosaic.jpg",
			wantWidth:  1200,
			wantHeight: 800,
		},
		{
			name: "multiple photos with unusable mosaic",
			media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{
					{URL: "https://img/1.jpg", Width: 640, Height: 480},
					{URL: "https://img/2.jpg", Width: 800, Height: 600},
				},
				Mosaic: &twitterxapi.Mosaic{
					Formats: map[string]string{"webp": "https://img/mosaic.webp"},
				},
			},
			wantURL:    "https://img/1.jpg",
			wantThumb:  "https://img/1.jpg",
			wantWidth:  640,
			wantHeight: 480,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotThumb, gotWidth, gotHeight := selectPhoto(tt.media)
			if gotURL != tt.wantURL || gotThumb != tt.wantThumb || gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Errorf("selectPhoto() = (%q, %q, %d, %d), want (%q, %q, %d, %d)",
					gotURL, gotThumb, gotWidth, gotHeight,
					tt.wantURL, tt.wantThumb, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestPickMosaicURL(t *testing.T) {
	tests := []struct {
		name   string
		mosaic *twitterxapi.Mosaic
		want   string
	}{
		{
			name:   "nil mosaic",
			mosaic: nil,
			want:   "",
		},
		{
			name:   "empty formats",
			mosaic: &twitterxapi.Mosaic{},
			want:   "",
		},
		{
			name: "jpeg preferred",
			mosaic: &twitterxapi.Mosaic{
				Formats: map[string]string{
					"jpg":  "https://img/one.jpg",
					"jpeg": "https://img/two.jpg",
				},
			},
			want: "https://img/two.jpg",
		},
		{
			name: "jpg fallback",
			mosaic: &twitterxapi.Mosaic{
				Formats: map[string]string{
					"jpg": "https://img/one.jpg",
				},
			},
			want: "https://img/one.jpg",
		},
		{
			name: "unknown formats",
			mosaic: &twitterxapi.Mosaic{
				Formats: map[string]string{
					"png": "https://img/one.png",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickMosaicURL(tt.mosaic); got != tt.want {
				t.Errorf("pickMosaicURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMosaicDimensions(t *testing.T) {
	tests := []struct {
		name   string
		mosaic *twitterxapi.Mosaic
		wantW  int
		wantH  int
	}{
		{
			name:   "nil mosaic",
			mosaic: nil,
			wantW:  0,
			wantH:  0,
		},
		{
			name:   "no sizes",
			mosaic: &twitterxapi.Mosaic{},
			wantW:  0,
			wantH:  0,
		},
		{
			name:   "width only",
			mosaic: &twitterxapi.Mosaic{Width: intPtr(640)},
			wantW:  640,
			wantH:  0,
		},
		{
			name:   "width and height",
			mosaic: &twitterxapi.Mosaic{Width: intPtr(640), Height: intPtr(480)},
			wantW:  640,
			wantH:  480,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := mosaicDimensions(tt.mosaic)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Errorf("mosaicDimensions() = (%d, %d), want (%d, %d)", gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestMimeTypeForVideo(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "empty format",
			format: "",
			want:   "video/mp4",
		},
		{
			name:   "has mime type",
			format: "video/quicktime",
			want:   "video/quicktime",
		},
		{
			name:   "simple format",
			format: "webm",
			want:   "video/webm",
		},
		{
			name:   "trimmed format",
			format: "  mp4  ",
			want:   "video/mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mimeTypeForVideo(tt.format); got != tt.want {
				t.Errorf("mimeTypeForVideo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	wide := "\u4e16\u754c\u4e16\u754c"

	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{
			name:  "max zero",
			input: "hello",
			max:   0,
			want:  "",
		},
		{
			name:  "short input",
			input: "hello",
			max:   10,
			want:  "hello",
		},
		{
			name:  "wide runes fit",
			input: wide,
			max:   6,
			want:  wide,
		},
		{
			name:  "max three or less",
			input: "abcdef",
			max:   3,
			want:  "abc",
		},
		{
			name:  "ellipsis branch",
			input: "abcdefghijklmnopqrstuvwxyz",
			max:   10,
			want:  "abcdefg...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateText(tt.input, tt.max); got != tt.want {
				t.Errorf("truncateText() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
			gotResult, gotOK := buildInlineResult(tt.tweet, tt.fallbackID)
			if gotOK != tt.wantOK {
				t.Fatalf("buildInlineResult() ok = %v, want %v", gotOK, tt.wantOK)
			}
			tt.assert(t, gotResult)
		})
	}
}

func intPtr(v int) *int {
	return &v
}
