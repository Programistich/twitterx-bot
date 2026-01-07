package tweet

import (
	"testing"
)

func TestEncodeChainCallback(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		tweetID      string
		replyToMsgID int64
		want         string
	}{
		{
			name:         "simple username and tweet ID",
			username:     "alice",
			tweetID:      "123456789",
			replyToMsgID: 100,
			want:         "chain:alice:123456789:100",
		},
		{
			name:         "username with underscore",
			username:     "user_name",
			tweetID:      "987654321",
			replyToMsgID: 200,
			want:         "chain:user_name:987654321:200",
		},
		{
			name:         "long tweet ID",
			username:     "bob",
			tweetID:      "1234567890123456789",
			replyToMsgID: 999999999,
			want:         "chain:bob:1234567890123456789:999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeChainCallback(tt.username, tt.tweetID, tt.replyToMsgID)
			if got != tt.want {
				t.Errorf("EncodeChainCallback() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeChainCallback(t *testing.T) {
	tests := []struct {
		name             string
		data             string
		wantUsername     string
		wantTweetID      string
		wantReplyToMsgID int64
		wantOk           bool
	}{
		{
			name:             "valid callback data",
			data:             "chain:alice:123456789:100",
			wantUsername:     "alice",
			wantTweetID:      "123456789",
			wantReplyToMsgID: 100,
			wantOk:           true,
		},
		{
			name:             "username with underscore",
			data:             "chain:user_name:987654321:200",
			wantUsername:     "user_name",
			wantTweetID:      "987654321",
			wantReplyToMsgID: 200,
			wantOk:           true,
		},
		{
			name:             "invalid prefix",
			data:             "other:alice:123:100",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "missing replyToMsgID",
			data:             "chain:alice:123456789",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "empty username",
			data:             "chain::123:100",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "empty tweet ID",
			data:             "chain:alice::100",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "empty data",
			data:             "",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "just prefix",
			data:             "chain:",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "invalid replyToMsgID (not a number)",
			data:             "chain:alice:123456789:abc",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
		{
			name:             "empty replyToMsgID",
			data:             "chain:alice:123456789:",
			wantUsername:     "",
			wantTweetID:      "",
			wantReplyToMsgID: 0,
			wantOk:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, tweetID, replyToMsgID, ok := DecodeChainCallback(tt.data)
			if ok != tt.wantOk {
				t.Errorf("DecodeChainCallback() ok = %v, want %v", ok, tt.wantOk)
			}
			if username != tt.wantUsername {
				t.Errorf("DecodeChainCallback() username = %q, want %q", username, tt.wantUsername)
			}
			if tweetID != tt.wantTweetID {
				t.Errorf("DecodeChainCallback() tweetID = %q, want %q", tweetID, tt.wantTweetID)
			}
			if replyToMsgID != tt.wantReplyToMsgID {
				t.Errorf("DecodeChainCallback() replyToMsgID = %d, want %d", replyToMsgID, tt.wantReplyToMsgID)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		username     string
		tweetID      string
		replyToMsgID int64
	}{
		{"alice", "123456789", 100},
		{"user_name", "987654321", 200},
		{"bob", "1234567890123456789", 999999999},
	}

	for _, tt := range tests {
		t.Run(tt.username+"/"+tt.tweetID, func(t *testing.T) {
			encoded := EncodeChainCallback(tt.username, tt.tweetID, tt.replyToMsgID)
			username, tweetID, replyToMsgID, ok := DecodeChainCallback(encoded)
			if !ok {
				t.Fatalf("DecodeChainCallback failed for encoded data: %s", encoded)
			}
			if username != tt.username {
				t.Errorf("Round-trip username = %q, want %q", username, tt.username)
			}
			if tweetID != tt.tweetID {
				t.Errorf("Round-trip tweetID = %q, want %q", tweetID, tt.tweetID)
			}
			if replyToMsgID != tt.replyToMsgID {
				t.Errorf("Round-trip replyToMsgID = %d, want %d", replyToMsgID, tt.replyToMsgID)
			}
		})
	}
}

func TestEncodeDeleteCallback(t *testing.T) {
	tests := []struct {
		name  string
		msgID int64
		opts  *KeyboardOpts
		want  string
	}{
		{
			name:  "simple message ID without chain",
			msgID: 100,
			opts:  nil,
			want:  "del:100",
		},
		{
			name:  "large message ID without chain",
			msgID: 999999999,
			opts:  nil,
			want:  "del:999999999",
		},
		{
			name:  "with chain info",
			msgID: 100,
			opts: &KeyboardOpts{
				ShowChainButton: true,
				ChainUsername:   "alice",
				ChainTweetID:    "123456789",
			},
			want: "del:100|alice|123456789",
		},
		{
			name:  "opts without chain button",
			msgID: 100,
			opts: &KeyboardOpts{
				ShowChainButton: false,
				ChainUsername:   "alice",
				ChainTweetID:    "123456789",
			},
			want: "del:100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeDeleteCallback(tt.msgID, tt.opts)
			if got != tt.want {
				t.Errorf("EncodeDeleteCallback() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeDeleteCallback(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		want   DeleteCallbackData
		wantOk bool
	}{
		{
			name:   "valid callback data without chain",
			data:   "del:100",
			want:   DeleteCallbackData{MsgID: 100, HasChain: false},
			wantOk: true,
		},
		{
			name:   "large message ID without chain",
			data:   "del:999999999",
			want:   DeleteCallbackData{MsgID: 999999999, HasChain: false},
			wantOk: true,
		},
		{
			name: "valid callback data with chain",
			data: "del:100|alice|123456789",
			want: DeleteCallbackData{
				MsgID:         100,
				HasChain:      true,
				ChainUsername: "alice",
				ChainTweetID:  "123456789",
			},
			wantOk: true,
		},
		{
			name:   "invalid prefix",
			data:   "other:100",
			want:   DeleteCallbackData{},
			wantOk: false,
		},
		{
			name:   "empty data",
			data:   "",
			want:   DeleteCallbackData{},
			wantOk: false,
		},
		{
			name:   "just prefix",
			data:   "del:",
			want:   DeleteCallbackData{},
			wantOk: false,
		},
		{
			name:   "invalid msgID (not a number)",
			data:   "del:abc",
			want:   DeleteCallbackData{},
			wantOk: false,
		},
		{
			name:   "incomplete chain info (missing tweetID)",
			data:   "del:100|alice",
			want:   DeleteCallbackData{MsgID: 100, HasChain: false},
			wantOk: true,
		},
		{
			name:   "empty chain username",
			data:   "del:100||123456789",
			want:   DeleteCallbackData{MsgID: 100, HasChain: false},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := DecodeDeleteCallback(tt.data)
			if ok != tt.wantOk {
				t.Errorf("DecodeDeleteCallback() ok = %v, want %v", ok, tt.wantOk)
			}
			if got != tt.want {
				t.Errorf("DecodeDeleteCallback() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBuildKeyboard(t *testing.T) {
	t.Run("with chain button", func(t *testing.T) {
		keyboard := BuildKeyboard(100, &KeyboardOpts{
			ShowChainButton: true,
			ChainUsername:   "alice",
			ChainTweetID:    "123456789",
		})

		if keyboard == nil {
			t.Fatal("BuildKeyboard returned nil")
		}

		if len(keyboard.InlineKeyboard) != 1 {
			t.Fatalf("expected 1 row, got %d", len(keyboard.InlineKeyboard))
		}

		row := keyboard.InlineKeyboard[0]
		if len(row) != 2 {
			t.Fatalf("expected 2 buttons in row, got %d", len(row))
		}

		chainBtn := row[0]
		if chainBtn.Text != "Send full chain" {
			t.Errorf("chain button text = %q, want %q", chainBtn.Text, "Send full chain")
		}
		if chainBtn.CallbackData != "chain:alice:123456789:100" {
			t.Errorf("chain button callback data = %q, want %q", chainBtn.CallbackData, "chain:alice:123456789:100")
		}

		deleteBtn := row[1]
		if deleteBtn.Text != "Delete original" {
			t.Errorf("delete button text = %q, want %q", deleteBtn.Text, "Delete original")
		}
		// Delete button should contain chain info when chain button is present
		if deleteBtn.CallbackData != "del:100|alice|123456789" {
			t.Errorf("delete button callback data = %q, want %q", deleteBtn.CallbackData, "del:100|alice|123456789")
		}
	})

	t.Run("without chain button", func(t *testing.T) {
		keyboard := BuildKeyboard(100, nil)

		if keyboard == nil {
			t.Fatal("BuildKeyboard returned nil")
		}

		if len(keyboard.InlineKeyboard) != 1 {
			t.Fatalf("expected 1 row, got %d", len(keyboard.InlineKeyboard))
		}

		row := keyboard.InlineKeyboard[0]
		if len(row) != 1 {
			t.Fatalf("expected 1 button in row, got %d", len(row))
		}

		deleteBtn := row[0]
		if deleteBtn.Text != "Delete original" {
			t.Errorf("delete button text = %q, want %q", deleteBtn.Text, "Delete original")
		}
		if deleteBtn.CallbackData != "del:100" {
			t.Errorf("delete button callback data = %q, want %q", deleteBtn.CallbackData, "del:100")
		}
	})
}

func TestCallbackDataLength(t *testing.T) {
	// Telegram callback data has a 64 byte limit
	// Test with realistic long values to ensure we stay under the limit

	// Longest reasonable username is 15 characters
	// Longest tweet ID is 19 digits
	// Longest message ID is ~10 digits (realistic max)
	username := "longestusername"
	tweetID := "1234567890123456789"
	replyToMsgID := int64(9999999999)

	t.Run("chain callback", func(t *testing.T) {
		data := EncodeChainCallback(username, tweetID, replyToMsgID)
		if len(data) > 64 {
			t.Errorf("chain callback data too long: %d bytes (max 64), data: %s", len(data), data)
		}
	})

	t.Run("delete callback with chain info", func(t *testing.T) {
		opts := &KeyboardOpts{
			ShowChainButton: true,
			ChainUsername:   username,
			ChainTweetID:    tweetID,
		}
		data := EncodeDeleteCallback(replyToMsgID, opts)
		if len(data) > 64 {
			t.Errorf("delete callback data too long: %d bytes (max 64), data: %s", len(data), data)
		}
	})
}
