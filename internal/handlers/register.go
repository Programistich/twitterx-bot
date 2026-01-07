package handlers

import (
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterurl"
	"twitterx-bot/internal/twitterxapi"
	"twitterx-bot/internal/usecase/tweetsvc"
)

type Handlers struct {
	log *logger.Logger
	api *twitterxapi.Client
	svc *tweetsvc.Service
}

const (
	inlineQueryTimeout = 10 * time.Second
	chainTimeout       = 30 * time.Second
)

func Register(d *ext.Dispatcher, log *logger.Logger, api *twitterxapi.Client) {
	if api == nil {
		api = twitterxapi.NewClient("")
	}
	h := &Handlers{
		log: log,
		api: api,
		svc: tweetsvc.New(api, nil),
	}

	d.AddHandler(handlers.NewCommand("start", start))
	d.AddHandler(handlers.NewInlineQuery(func(iq *gotgbot.InlineQuery) bool {
		return true
	}, h.inlineQuery))
	d.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		if msg.Text == "" {
			return false
		}
		_, _, ok := twitterurl.ParseTweetURL(msg.Text)
		return ok
	}, h.messageHandler))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.ChainCallbackPrefix)
	}, h.chainCallback))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.DeleteCallbackPrefix)
	}, h.deleteCallback))
}
