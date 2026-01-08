package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"twitterx-bot/internal/handlers/callback"
	"twitterx-bot/internal/handlers/inline"
	"twitterx-bot/internal/handlers/message"
	"twitterx-bot/internal/handlers/start"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterurl"
	"twitterx-bot/internal/twitterxapi"
	inlineuc "twitterx-bot/internal/usecase/tweetsvc/inline"
)

const (
	inlineQueryTimeout = 10 * time.Second
	messageTimeout     = 10 * time.Second
	chainTimeout       = 30 * time.Second
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error)
}

func Register(d *ext.Dispatcher, log *logger.Logger, api *twitterxapi.Client, telegraph tweet.ArticleCreator) {
	if api == nil {
		api = twitterxapi.NewClient("")
	}
	RegisterWithFetcher(d, log, api, telegraph)
}

// RegisterWithFetcher registers handlers using a custom TweetFetcher implementation.
// This is useful for testing with mock implementations.
func RegisterWithFetcher(d *ext.Dispatcher, log *logger.Logger, fetcher TweetFetcher, telegraph tweet.ArticleCreator) {
	// Start and help commands
	d.AddHandler(handlers.NewCommand("start", start.Handler))
	d.AddHandler(handlers.NewCommand("help", start.Handler))

	// Inline query handler
	inlineUC := inlineuc.New(fetcher)
	inlineHandler := inline.New(log, inlineUC, inlineQueryTimeout)
	d.AddHandler(handlers.NewInlineQuery(func(iq *gotgbot.InlineQuery) bool {
		return true
	}, inlineHandler.Handle))

	// Message handler for Twitter URLs
	messageHandler := message.New(log, fetcher, messageTimeout, telegraph)
	d.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		if msg.Text == "" {
			return false
		}
		_, _, ok := twitterurl.ParseTweetURL(msg.Text)
		return ok
	}, messageHandler.Handle))

	// Callback handlers
	callbackHandlers := callback.New(log, fetcher, chainTimeout, telegraph)
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.ChainCallbackPrefix)
	}, callbackHandlers.Chain))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.DeleteCallbackPrefix)
	}, callbackHandlers.Delete))
}
