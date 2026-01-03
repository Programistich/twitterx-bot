package tweet

import (
	"testing"

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
			if got := BuildTitle(tt.tweet); got != tt.want {
				t.Errorf("BuildTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContent(t *testing.T) {
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
			if got := Content(tt.tweet); got != tt.want {
				t.Errorf("Content() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCaptionAndMessageText(t *testing.T) {
	tw := &twitterxapi.Tweet{
		Text: "hello",
		URL:  "https://x.com/1",
	}
	want := "hello\n\nhttps://x.com/1"

	if got := Caption(tw); got != want {
		t.Errorf("Caption() = %q, want %q", got, want)
	}
	if got := MessageText(tw); got != want {
		t.Errorf("MessageText() = %q, want %q", got, want)
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
			if got := TruncateText(tt.input, tt.max); got != tt.want {
				t.Errorf("TruncateText() = %q, want %q", got, tt.want)
			}
		})
	}
}
