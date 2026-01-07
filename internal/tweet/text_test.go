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

func TestHTMLContent(t *testing.T) {
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
			name: "full tweet with name and screen name",
			tweet: &twitterxapi.Tweet{
				URL:  "https://x.com/alice/status/123",
				Text: "Hello world",
				Author: twitterxapi.Author{
					Name:       "Alice Smith",
					ScreenName: "alice",
				},
			},
			want: `<a href="https://x.com/alice/status/123">Tweet</a> from <a href="https://x.com/alice">Alice Smith</a>

Hello world`,
		},
		{
			name: "falls back to @screen name when name is empty",
			tweet: &twitterxapi.Tweet{
				URL:  "https://x.com/bob/status/456",
				Text: "Test tweet",
				Author: twitterxapi.Author{
					ScreenName: "bob",
				},
			},
			want: `<a href="https://x.com/bob/status/456">Tweet</a> from <a href="https://x.com/bob">@bob</a>

Test tweet`,
		},
		{
			name: "no author info",
			tweet: &twitterxapi.Tweet{
				URL:  "https://x.com/user/status/789",
				Text: "No author",
			},
			want: `<a href="https://x.com/user/status/789">Tweet</a>

No author`,
		},
		{
			name: "escapes HTML in text",
			tweet: &twitterxapi.Tweet{
				URL:  "https://x.com/test/status/1",
				Text: "Hello <script>alert('xss')</script> & goodbye",
				Author: twitterxapi.Author{
					Name:       "Test <User>",
					ScreenName: "test",
				},
			},
			want: `<a href="https://x.com/test/status/1">Tweet</a> from <a href="https://x.com/test">Test &lt;User&gt;</a>

Hello &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; &amp; goodbye`,
		},
		{
			name: "no URL still shows Tweet text",
			tweet: &twitterxapi.Tweet{
				Text: "Just text",
				Author: twitterxapi.Author{
					Name:       "User",
					ScreenName: "user",
				},
			},
			want: `Tweet from <a href="https://x.com/user">User</a>

Just text`,
		},
		{
			name: "no text only header",
			tweet: &twitterxapi.Tweet{
				URL: "https://x.com/user/status/1",
				Author: twitterxapi.Author{
					Name:       "User",
					ScreenName: "user",
				},
			},
			want: `<a href="https://x.com/user/status/1">Tweet</a> from <a href="https://x.com/user">User</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := DefaultFormatter()
			if got := f.HTMLContent(tt.tweet); got != tt.want {
				t.Errorf("HTMLContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{
			name:  "max zero",
			input: "<a>hello</a>",
			max:   0,
			want:  "",
		},
		{
			name:  "short input fits",
			input: "<a>hi</a>",
			max:   10,
			want:  "<a>hi</a>",
		},
		{
			name:  "truncates preserving tags",
			input: `<a href="url">Tweet</a> from <a href="profile">Author Name</a>

This is a long tweet text that needs to be truncated`,
			max:  50,
			want: `<a href="url">Tweet</a> from <a href="profile">Author Name</a>

This is a long tweet te...`,
		},
		{
			name:  "closes unclosed tags",
			input: "<a>abcdefghij</a>",
			max:   8,
			want:  "<a>abcde...</a>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateHTML(tt.input, tt.max); got != tt.want {
				t.Errorf("TruncateHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTMLCaptionAndMessageText(t *testing.T) {
	tw := &twitterxapi.Tweet{
		URL:  "https://x.com/user/status/1",
		Text: "Hello world",
		Author: twitterxapi.Author{
			Name:       "Test User",
			ScreenName: "testuser",
		},
	}

	caption := HTMLCaption(tw)
	messageText := HTMLMessageText(tw)

	// Both should contain the formatted content
	expectedContent := `<a href="https://x.com/user/status/1">Tweet</a> from <a href="https://x.com/testuser">Test User</a>

Hello world`

	if caption != expectedContent {
		t.Errorf("HTMLCaption() = %q, want %q", caption, expectedContent)
	}
	if messageText != expectedContent {
		t.Errorf("HTMLMessageText() = %q, want %q", messageText, expectedContent)
	}
}
