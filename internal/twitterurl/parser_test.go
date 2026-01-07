package twitterurl

import "testing"

func TestParseTweetURL(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantMatch    bool
		wantUsername string
		wantTweetID  string
	}{
		// ===========================================
		// Basic cases - twitter.com
		// ===========================================
		{
			name:         "twitter.com with https",
			input:        "https://twitter.com/elonmusk/status/1234567890123456789",
			wantMatch:    true,
			wantUsername: "elonmusk",
			wantTweetID:  "1234567890123456789",
		},
		{
			name:         "twitter.com with http",
			input:        "http://twitter.com/NASA/status/9876543210",
			wantMatch:    true,
			wantUsername: "NASA",
			wantTweetID:  "9876543210",
		},
		{
			name:         "twitter.com without protocol",
			input:        "twitter.com/jack/status/111222333",
			wantMatch:    true,
			wantUsername: "jack",
			wantTweetID:  "111222333",
		},
		{
			name:         "twitter.com with www",
			input:        "https://www.twitter.com/OpenAI/status/555666777888",
			wantMatch:    true,
			wantUsername: "OpenAI",
			wantTweetID:  "555666777888",
		},
		{
			name:         "twitter.com with www without protocol",
			input:        "www.twitter.com/github/status/123123123",
			wantMatch:    true,
			wantUsername: "github",
			wantTweetID:  "123123123",
		},

		// ===========================================
		// Basic cases - x.com
		// ===========================================
		{
			name:         "x.com with https",
			input:        "https://x.com/elonmusk/status/1234567890123456789",
			wantMatch:    true,
			wantUsername: "elonmusk",
			wantTweetID:  "1234567890123456789",
		},
		{
			name:         "x.com with http",
			input:        "http://x.com/TwitterDev/status/999888777666",
			wantMatch:    true,
			wantUsername: "TwitterDev",
			wantTweetID:  "999888777666",
		},
		{
			name:         "x.com without protocol",
			input:        "x.com/verified/status/444555666",
			wantMatch:    true,
			wantUsername: "verified",
			wantTweetID:  "444555666",
		},
		{
			name:         "x.com with www",
			input:        "https://www.x.com/X/status/777888999",
			wantMatch:    true,
			wantUsername: "X",
			wantTweetID:  "777888999",
		},

		// ===========================================
		// Alternative domains (vxtwitter, fxtwitter, etc.)
		// Current regex SUPPORTS these - verify if this is desired behavior
		// ===========================================
		{
			name:         "mobile twitter (m.twitter.com)",
			input:        "https://m.twitter.com/user/status/123456",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456",
		},
		{
			name:         "vxtwitter.com",
			input:        "https://vxtwitter.com/user/status/123456",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456",
		},
		{
			name:         "fxtwitter.com",
			input:        "https://fxtwitter.com/user/status/123456",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456",
		},
		{
			name:         "fixupx.com",
			input:        "https://fixupx.com/user/status/123456",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456",
		},
		{
			name:         "nitter instance (nitter.net) - NOT supported",
			input:        "https://nitter.net/user/status/123456",
			wantMatch:    false,
			wantUsername: "",
			wantTweetID:  "",
		},

		// ===========================================
		// Different username formats
		// ===========================================
		{
			name:         "username with underscore",
			input:        "https://twitter.com/user_name_123/status/111222333",
			wantMatch:    true,
			wantUsername: "user_name_123",
			wantTweetID:  "111222333",
		},
		{
			name:         "username with only digits",
			input:        "https://x.com/12345/status/999000111",
			wantMatch:    true,
			wantUsername: "12345",
			wantTweetID:  "999000111",
		},
		{
			name:         "short username (1 char)",
			input:        "https://twitter.com/a/status/123",
			wantMatch:    true,
			wantUsername: "a",
			wantTweetID:  "123",
		},
		{
			name:         "username uppercase",
			input:        "https://twitter.com/UPPERCASE/status/456789",
			wantMatch:    true,
			wantUsername: "UPPERCASE",
			wantTweetID:  "456789",
		},
		{
			name:         "username mixed case",
			input:        "https://twitter.com/CamelCase_User123/status/789012",
			wantMatch:    true,
			wantUsername: "CamelCase_User123",
			wantTweetID:  "789012",
		},

		// ===========================================
		// URL with query parameters and fragments
		// ===========================================
		{
			name:         "URL with query parameter s",
			input:        "https://twitter.com/user/status/123456789?s=20",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456789",
		},
		{
			name:         "URL with multiple query parameters",
			input:        "https://twitter.com/user/status/123456789?s=20&t=abc123XYZ",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456789",
		},
		{
			name:         "URL with ref_src parameter",
			input:        "https://twitter.com/user/status/123456789?ref_src=twsrc%5Etfw",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456789",
		},

		// ===========================================
		// URL in text (inline)
		// ===========================================
		{
			name:         "URL at the beginning of text",
			input:        "https://twitter.com/user/status/123456 check this out",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123456",
		},
		{
			name:         "URL in the middle of text",
			input:        "Look at this tweet https://x.com/cool_user/status/789012 it's amazing",
			wantMatch:    true,
			wantUsername: "cool_user",
			wantTweetID:  "789012",
		},
		{
			name:         "URL at the end of text",
			input:        "Check out: https://twitter.com/viral/status/999888777",
			wantMatch:    true,
			wantUsername: "viral",
			wantTweetID:  "999888777",
		},
		{
			name:         "URL with non-latin text around",
			input:        "Check this tweet https://x.com/ukraine/status/123456789 cool!",
			wantMatch:    true,
			wantUsername: "ukraine",
			wantTweetID:  "123456789",
		},

		// ===========================================
		// Negative cases - should NOT match
		// ===========================================
		{
			name:      "empty string",
			input:     "",
			wantMatch: false,
		},
		{
			name:      "plain text",
			input:     "hello world",
			wantMatch: false,
		},
		{
			name:      "different domain (facebook)",
			input:     "https://facebook.com/user/status/123",
			wantMatch: false,
		},
		{
			name:      "different domain (instagram)",
			input:     "https://instagram.com/p/ABC123",
			wantMatch: false,
		},
		{
			name:      "twitter profile without status",
			input:     "https://twitter.com/elonmusk",
			wantMatch: false,
		},
		{
			name:      "twitter profile with followers tab",
			input:     "https://twitter.com/elonmusk/followers",
			wantMatch: false,
		},
		{
			name:      "twitter profile with likes tab",
			input:     "https://twitter.com/elonmusk/likes",
			wantMatch: false,
		},
		{
			name:      "twitter search",
			input:     "https://twitter.com/search?q=golang",
			wantMatch: false,
		},
		{
			name:      "twitter explore",
			input:     "https://twitter.com/explore",
			wantMatch: false,
		},
		{
			name:      "invalid format - no ID",
			input:     "https://twitter.com/user/status/",
			wantMatch: false,
		},
		{
			name:      "invalid format - letters instead of ID",
			input:     "https://twitter.com/user/status/abcdef",
			wantMatch: false,
		},
		{
			name:         "mixed letters and digits in ID - extracts only leading digits",
			input:        "https://twitter.com/user/status/123abc456",
			wantMatch:    true,
			wantUsername: "user",
			wantTweetID:  "123",
		},
		{
			name:      "twitter domain but not .com",
			input:     "https://twitter.org/user/status/123456",
			wantMatch: false,
		},
		{
			name:      "x domain but not .com",
			input:     "https://x.org/user/status/123456",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, tweetID, ok := ParseTweetURL(tt.input)

			if tt.wantMatch {
				if !ok {
					t.Errorf("expected match for %q, but not found", tt.input)
					return
				}
				if username != tt.wantUsername {
					t.Errorf("username: expected %q, got %q", tt.wantUsername, username)
				}
				if tweetID != tt.wantTweetID {
					t.Errorf("tweetID: expected %q, got %q", tt.wantTweetID, tweetID)
				}
			} else if ok {
				t.Errorf("unexpected match for %q, found: username=%q, id=%q",
					tt.input, username, tweetID)
			}
		})
	}
}
