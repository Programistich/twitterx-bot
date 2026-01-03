package tweet

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const (
	ChainCallbackPrefix = "chain:"
)

// EncodeChainCallback creates callback data for the "Send full chain" button.
// Format: chain:username:tweetID
func EncodeChainCallback(username, tweetID string) string {
	return ChainCallbackPrefix + username + ":" + tweetID
}

// DecodeChainCallback parses callback data and extracts username and tweetID.
// Returns ok=false if the format is invalid.
func DecodeChainCallback(data string) (username, tweetID string, ok bool) {
	if !strings.HasPrefix(data, ChainCallbackPrefix) {
		return "", "", false
	}

	rest := strings.TrimPrefix(data, ChainCallbackPrefix)
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// BuildChainKeyboard creates an inline keyboard with the "Send full chain" button.
func BuildChainKeyboard(username, tweetID string) *gotgbot.InlineKeyboardMarkup {
	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Send full chain",
					CallbackData: EncodeChainCallback(username, tweetID),
				},
			},
		},
	}
}
