package tweet

import (
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

const (
	ChainCallbackPrefix  = "chain:"
	DeleteCallbackPrefix = "del:"
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

// EncodeDeleteCallback creates callback data for the "Delete original" button.
// Format: del:msgID or del:msgID|username|tweetID (if chain info is provided)
func EncodeDeleteCallback(msgID int64, chainOpts *KeyboardOpts) string {
	base := DeleteCallbackPrefix + strconv.FormatInt(msgID, 10)
	if chainOpts != nil && chainOpts.ShowChainButton {
		return base + "|" + chainOpts.ChainUsername + "|" + chainOpts.ChainTweetID
	}
	return base
}

// DeleteCallbackData holds the decoded delete callback information.
type DeleteCallbackData struct {
	MsgID         int64
	HasChain      bool
	ChainUsername string
	ChainTweetID  string
}

// DecodeDeleteCallback parses callback data and extracts msgID and optional chain info.
// Returns ok=false if the format is invalid.
func DecodeDeleteCallback(data string) (result DeleteCallbackData, ok bool) {
	if !strings.HasPrefix(data, DeleteCallbackPrefix) {
		return DeleteCallbackData{}, false
	}

	rest := strings.TrimPrefix(data, DeleteCallbackPrefix)
	if rest == "" {
		return DeleteCallbackData{}, false
	}

	// Check if there's chain info (separated by |)
	parts := strings.SplitN(rest, "|", 3)

	msgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return DeleteCallbackData{}, false
	}

	result.MsgID = msgID

	// If we have chain info (username|tweetID)
	if len(parts) == 3 && parts[1] != "" && parts[2] != "" {
		result.HasChain = true
		result.ChainUsername = parts[1]
		result.ChainTweetID = parts[2]
	}

	return result, true
}

// KeyboardOpts configures the inline keyboard buttons.
type KeyboardOpts struct {
	ShowChainButton bool
	ChainUsername   string
	ChainTweetID    string
}

// BuildKeyboard creates an inline keyboard with optional buttons.
// Always includes "Delete original" button, optionally includes "Send full chain".
func BuildKeyboard(replyToMsgID int64, opts *KeyboardOpts) *gotgbot.InlineKeyboardMarkup {
	var buttons []gotgbot.InlineKeyboardButton

	if opts != nil && opts.ShowChainButton {
		buttons = append(buttons, gotgbot.InlineKeyboardButton{
			Text:         "Send full chain",
			CallbackData: EncodeChainCallback(opts.ChainUsername, opts.ChainTweetID, replyToMsgID),
		})
	}

	buttons = append(buttons, gotgbot.InlineKeyboardButton{
		Text:         "Delete original",
		CallbackData: EncodeDeleteCallback(replyToMsgID, opts),
	})

	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{buttons},
	}
}

// FindChainButton searches for the chain button in the keyboard and returns its callback data.
// Returns empty string if not found.
func FindChainButton(markup *gotgbot.InlineKeyboardMarkup) string {
	if markup == nil {
		return ""
	}
	for _, row := range markup.InlineKeyboard {
		for _, btn := range row {
			if strings.HasPrefix(btn.CallbackData, ChainCallbackPrefix) {
				return btn.CallbackData
			}
		}
	}
	return ""
}

// BuildChainOnlyKeyboard creates a keyboard with only the "Send full chain" button.
func BuildChainOnlyKeyboard(chainCallbackData string) *gotgbot.InlineKeyboardMarkup {
	return &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Send full chain",
					CallbackData: chainCallbackData,
				},
			},
		},
	}
}
