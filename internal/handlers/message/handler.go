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
	log       *logger.Logger
	fetcher   TweetFetcher
	timeout   time.Duration
	telegraph tweet.ArticleCreator
}

// New creates a new message handler with the supplied logger, tweet fetcher, and timeout.
func New(log *logger.Logger, fetcher TweetFetcher, timeout time.Duration, telegraph tweet.ArticleCreator) *Handler {
	return &Handler{log: log, fetcher: fetcher, timeout: timeout, telegraph: telegraph}
}

// Handle processes incoming Telegram messages that contain Twitter URLs.
func (h *Handler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	log := h.log.With("component", "message")
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}
	if ctx.EffectiveUser != nil {
		log = log.With("user_id", ctx.EffectiveUser.Id, "username", ctx.EffectiveUser.Username)
	}
	if ctx.EffectiveMessage != nil {
		log = log.With("message_id", ctx.EffectiveMessage.MessageId)
	}
	log.Debug("message received", "text", text)

	username, tweetID, ok := twitterurl.ParseTweetURL(text)
	if !ok {
		log.Debug("message ignored: no tweet url")
		return nil
	}

	log.Info("tweet url parsed", "tweet_username", username, "tweet_id", tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	_, err := b.SendChatAction(ctx.EffectiveChat.Id, gotgbot.ChatActionTyping, &gotgbot.SendChatActionOpts{})
	if err != nil {
		log.Debug("send chat action failed", "err", err)
	}

	uc := sendtweet.New(h.fetcher, tweet.Sender{Bot: b, Telegraph: h.telegraph, Log: log})
	if sendErr := uc.SendTweet(reqCtx, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, username, tweetID, shared.UserDisplayName(ctx.EffectiveUser)); sendErr != nil {
		log.Error("send tweet failed", "tweet_username", username, "tweet_id", tweetID, "err", sendErr)
		return nil
	}

	log.Info("tweet sent", "tweet_username", username, "tweet_id", tweetID)
	return nil
}
