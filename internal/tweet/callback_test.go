package tweet

import (
	"testing"
)

func TestEncodeChainCallback(t *testing.T) {
	tests := []struct {
		name     string
		username string
		tweetID  string
		want     string
	}{
		{
			name:     "simple username and tweet ID",
			username: "alice",
			tweetID:  "123456789",
			want:     "chain:alice:123456789",
		},
		{
			name:     "username with underscore",
			username: "user_name",
			tweetID:  "987654321",
			want:     "chain:user_name:987654321",
		},
		{
			name:     "long tweet ID",
			username: "bob",
			tweetID:  "1234567890123456789",
			want:     "chain:bob:1234567890123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeChainCallback(tt.username, tt.tweetID)
			if got != tt.want {
				t.Errorf("EncodeChainCallback() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeChainCallback(t *testing.T) {
	tests := []struct {
		name         string
		data         string
		wantUsername string
		wantTweetID  string
		wantOk       bool
	}{
		{
			name:         "valid callback data",
			data:         "chain:alice:123456789",
			wantUsername: "alice",
			wantTweetID:  "123456789",
			wantOk:       true,
		},
		{
			name:         "username with underscore",
			data:         "chain:user_name:987654321",
			wantUsername: "user_name",
			wantTweetID:  "987654321",
			wantOk:       true,
		},
		{
			name:         "invalid prefix",
			data:         "other:alice:123",
			wantUsername: "",
			wantTweetID:  "",
			wantOk:       false,
		},
		{
			name:         "missing tweet ID",
			data:         "chain:alice",
			wantUsername: "",
			wantTweetID:  "",
			wantOk:       false,
		},
		{
			name:         "empty username",
			data:         "chain::123",
			wantUsername: "",
			wantTweetID:  "",
			wantOk:       false,
		},
		{
			name:         "empty tweet ID",
			data:         "chain:alice:",
			wantUsername: "",
			wantTweetID:  "",
			wantOk:       false,
		},
		{
			name:         "empty data",
			data:         "",
			wantUsername: "",
			wantTweetID:  "",
			wantOk:       false,
		},
		{
			name:         "just prefix",
			data:         "chain:",
			wantUsername: "",
			wantTweetID:  "",
			wantOk:       false,
		},
		{
			name:         "extra colons in tweet ID preserved",
			data:         "chain:alice:123:456",
			wantUsername: "alice",
			wantTweetID:  "123:456",
			wantOk:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, tweetID, ok := DecodeChainCallback(tt.data)
			if ok != tt.wantOk {
				t.Errorf("DecodeChainCallback() ok = %v, want %v", ok, tt.wantOk)
			}
			if username != tt.wantUsername {
				t.Errorf("DecodeChainCallback() username = %q, want %q", username, tt.wantUsername)
			}
			if tweetID != tt.wantTweetID {
				t.Errorf("DecodeChainCallback() tweetID = %q, want %q", tweetID, tt.wantTweetID)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		username string
		tweetID  string
	}{
		{"alice", "123456789"},
		{"user_name", "987654321"},
		{"bob", "1234567890123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.username+"/"+tt.tweetID, func(t *testing.T) {
			encoded := EncodeChainCallback(tt.username, tt.tweetID)
			username, tweetID, ok := DecodeChainCallback(encoded)
			if !ok {
				t.Fatalf("DecodeChainCallback failed for encoded data: %s", encoded)
			}
			if username != tt.username {
				t.Errorf("Round-trip username = %q, want %q", username, tt.username)
			}
			if tweetID != tt.tweetID {
				t.Errorf("Round-trip tweetID = %q, want %q", tweetID, tt.tweetID)
			}
		})
	}
}

func TestBuildChainKeyboard(t *testing.T) {
	keyboard := BuildChainKeyboard("alice", "123456789")

	if keyboard == nil {
		t.Fatal("BuildChainKeyboard returned nil")
	}

	if len(keyboard.InlineKeyboard) != 1 {
		t.Fatalf("expected 1 row, got %d", len(keyboard.InlineKeyboard))
	}

	row := keyboard.InlineKeyboard[0]
	if len(row) != 1 {
		t.Fatalf("expected 1 button in row, got %d", len(row))
	}

	button := row[0]
	if button.Text != "Send full chain" {
		t.Errorf("button text = %q, want %q", button.Text, "Send full chain")
	}

	expectedData := "chain:alice:123456789"
	if button.CallbackData != expectedData {
		t.Errorf("button callback data = %q, want %q", button.CallbackData, expectedData)
	}
}

func TestCallbackDataLength(t *testing.T) {
	// Telegram callback data has a 64 byte limit
	// Test with realistic long values to ensure we stay under the limit

	// Longest reasonable username is 15 characters
	// Longest tweet ID is 19 digits
	username := "longestusername"
	tweetID := "1234567890123456789"

	data := EncodeChainCallback(username, tweetID)

	if len(data) > 64 {
		t.Errorf("callback data too long: %d bytes (max 64), data: %s", len(data), data)
	}
}
