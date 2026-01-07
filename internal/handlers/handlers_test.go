package handlers

import (
	"strings"
	"testing"

	"twitterx-bot/internal/tweet"
)

func TestChainCallbackFilter(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantMatch bool
	}{
		{
			name:      "valid chain callback",
			data:      "chain:alice:123456789",
			wantMatch: true,
		},
		{
			name:      "valid chain callback with underscore",
			data:      "chain:user_name:987654321",
			wantMatch: true,
		},
		{
			name:      "other callback type",
			data:      "other:data:here",
			wantMatch: false,
		},
		{
			name:      "empty data",
			data:      "",
			wantMatch: false,
		},
		{
			name:      "similar prefix but not chain",
			data:      "chains:alice:123",
			wantMatch: false,
		},
		{
			name:      "just prefix",
			data:      "chain:",
			wantMatch: true, // matches prefix, validation happens in decode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is the same filter used in Register()
			matches := strings.HasPrefix(tt.data, tweet.ChainCallbackPrefix)
			if matches != tt.wantMatch {
				t.Errorf("filter(%q) = %v, want %v", tt.data, matches, tt.wantMatch)
			}
		})
	}
}

func TestReplyDetection(t *testing.T) {
	// Helper to create string pointer
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name             string
		replyingToStatus *string
		shouldHaveButton bool
	}{
		{
			name:             "tweet is a reply",
			replyingToStatus: ptr("123456789"),
			shouldHaveButton: true,
		},
		{
			name:             "tweet is not a reply",
			replyingToStatus: nil,
			shouldHaveButton: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from messageHandler
			hasButton := tt.replyingToStatus != nil
			if hasButton != tt.shouldHaveButton {
				t.Errorf("hasButton = %v, want %v", hasButton, tt.shouldHaveButton)
			}
		})
	}
}
