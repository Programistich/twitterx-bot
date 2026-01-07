package handlers

import (
	"context"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/telegram/tweet"
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

	svc := h.svc.WithSender(tweet.Sender{Bot: b})
	if err := svc.SendTweet(reqCtx, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, username, tweetID, userDisplayName(ctx.EffectiveUser)); err != nil {
		h.log.Error("failed to send tweet %s for %s: %v", tweetID, username, err)
		return nil
	}

	return nil
}
