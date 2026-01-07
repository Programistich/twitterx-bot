package handlers

import (
	"context"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/tweet"
	"twitterx-bot/internal/twitterurl"
)

func (h *Handlers) messageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	h.log.Debug("received message with twitter URL: %s", text)

	username, tweetID, ok := twitterurl.ParseTweetURL(text)
	if !ok {
		return nil
	}

	h.log.Info("twitter URL from message - username: %s, tweet_id: %s", username, tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), inlineQueryTimeout)
	defer cancel()

	_, err := b.SendChatAction(ctx.EffectiveChat.Id, gotgbot.ChatActionTyping, &gotgbot.SendChatActionOpts{})
	if err != nil {
		h.log.Debug("failed to send typing action: %v", err)
	}

	tw, err := h.api.GetTweet(reqCtx, username, tweetID)
	if err != nil {
		h.log.Error("failed to fetch tweet %s for %s: %v", tweetID, username, err)
		return nil
	}

	// Build keyboard with optional chain button and always delete button
	var keyboardOpts *tweet.KeyboardOpts
	if tw.ReplyingToStatus != nil {
		keyboardOpts = &tweet.KeyboardOpts{
			ShowChainButton: true,
			ChainUsername:   username,
			ChainTweetID:    tweetID,
		}
	}

	opts := &tweet.SendResponseOpts{
		ReplyMarkup:       tweet.BuildKeyboard(ctx.EffectiveMessage.MessageId, keyboardOpts),
		RequesterUsername: userDisplayName(ctx.EffectiveUser),
	}

	return tweet.SendResponse(b, ctx, tw, opts)
}
