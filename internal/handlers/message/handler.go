package message

import (
	"context"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/database"
	"twitterx-bot/internal/handlers/shared"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/translation"
	"twitterx-bot/internal/twitterurl"
	"twitterx-bot/internal/usecase/tweetsvc/sendtweet"
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	sendtweet.TweetFetcher
}

// ChatSettingsProvider provides chat settings operations.
type ChatSettingsProvider interface {
	GetLanguage(ctx context.Context, chatID int64) (string, error)
	UpdateLanguage(ctx context.Context, chatID int64, language string) (*database.ChatSettings, error)
}

// Handler encapsulates the dependencies required for processing message-based tweets.
type Handler struct {
	log          *logger.Logger
	fetcher      TweetFetcher
	timeout      time.Duration
	telegraph    tweet.ArticleCreator
	chatSettings ChatSettingsProvider
	translator   translation.Translator
}

// New creates a new message handler with the supplied logger, tweet fetcher, and timeout.
func New(log *logger.Logger, fetcher TweetFetcher, timeout time.Duration, telegraph tweet.ArticleCreator, chatSettings ChatSettingsProvider, translator translation.Translator) *Handler {
	return &Handler{log: log, fetcher: fetcher, timeout: timeout, telegraph: telegraph, chatSettings: chatSettings, translator: translator}
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

	// Get language for this chat
	lang := database.DefaultLanguage
	if h.chatSettings != nil {
		if l, langErr := h.chatSettings.GetLanguage(reqCtx, ctx.EffectiveChat.Id); langErr == nil {
			lang = l
		}
	}

	sender := tweet.Sender{Bot: b, Telegraph: h.telegraph, Translator: h.translator, Log: log, Lang: lang}
	uc := sendtweet.NewWithChain(h.fetcher, sender, sender)
	if sendErr := uc.SendTweet(reqCtx, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, username, tweetID, shared.UserDisplayName(ctx.EffectiveUser), lang); sendErr != nil {
		log.Error("send tweet failed", "tweet_username", username, "tweet_id", tweetID, "err", sendErr)
		return nil
	}

	log.Info("tweet sent", "tweet_username", username, "tweet_id", tweetID)
	return nil
}
