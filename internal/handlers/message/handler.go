package message

import (
	"context"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/handlers/shared"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterurl"
	"twitterx-bot/internal/usecase/tweetsvc/sendtweet"
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	sendtweet.TweetFetcher
}

// Handler encapsulates the dependencies required for processing message-based tweets.
type Handler struct {
	log     *logger.Logger
	fetcher TweetFetcher
	timeout time.Duration
}

// New creates a new message handler with the supplied logger, tweet fetcher, and timeout.
func New(log *logger.Logger, fetcher TweetFetcher, timeout time.Duration) *Handler {
	return &Handler{log: log, fetcher: fetcher, timeout: timeout}
}

// Handle processes incoming Telegram messages that contain Twitter URLs.
func (h *Handler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	h.log.Debug("received message with twitter URL: %s", text)

	username, tweetID, ok := twitterurl.ParseTweetURL(text)
	if !ok {
		return nil
	}

	h.log.Info("twitter URL from message - username: %s, tweet_id: %s", username, tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	_, err := b.SendChatAction(ctx.EffectiveChat.Id, gotgbot.ChatActionTyping, &gotgbot.SendChatActionOpts{})
	if err != nil {
		h.log.Debug("failed to send typing action: %v", err)
	}

	uc := sendtweet.New(h.fetcher, tweet.Sender{Bot: b})
	if sendErr := uc.SendTweet(reqCtx, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, username, tweetID, shared.UserDisplayName(ctx.EffectiveUser)); sendErr != nil {
		h.log.Error("failed to send tweet %s for %s: %v", tweetID, username, sendErr)
		return nil
	}

	return nil
}
