package tweet

import (
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const (
	ChainCallbackPrefix = "chain:"
)

// EncodeChainCallback creates callback data for the "Send full chain" button.
// Format: chain:username:tweetID:replyToMsgID
func EncodeChainCallback(username, tweetID string, replyToMsgID int64) string {
	return ChainCallbackPrefix + username + ":" + tweetID + ":" + strconv.FormatInt(replyToMsgID, 10)
}

// DecodeChainCallback parses callback data and extracts username, tweetID, and replyToMsgID.
// Returns ok=false if the format is invalid.
func DecodeChainCallback(data string) (username, tweetID string, replyToMsgID int64, ok bool) {
	if !strings.HasPrefix(data, ChainCallbackPrefix) {
		return "", "", 0, false
	}

	rest := strings.TrimPrefix(data, ChainCallbackPrefix)
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, false
	}

	msgID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", "", 0, false
	}

	return parts[0], parts[1], msgID, true
}

// BuildChainKeyboard creates an inline keyboard with the "Send full chain" button.
func BuildChainKeyboard(username, tweetID string, replyToMsgID int64) *gotgbot.InlineKeyboardMarkup {
	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Send full chain",
					CallbackData: EncodeChainCallback(username, tweetID, replyToMsgID),
				},
			},
		},
	}
}
