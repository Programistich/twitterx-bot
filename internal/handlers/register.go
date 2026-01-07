package handlers

import (
	"regexp"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/tweet"
	"twitterx-bot/internal/twitterxapi"
)

var twitterURLRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?(?:twitter\.com|x\.com)/([^/]+)/status/(\d+)`)

type Handlers struct {
	log *logger.Logger
	api *twitterxapi.Client
}

const (
	inlineQueryTimeout = 10 * time.Second
	chainTimeout       = 30 * time.Second
)

func Register(d *ext.Dispatcher, log *logger.Logger, api *twitterxapi.Client) {
	if api == nil {
		api = twitterxapi.NewClient("")
	}
	h := &Handlers{log: log, api: api}

	d.AddHandler(handlers.NewCommand("start", start))
	d.AddHandler(handlers.NewInlineQuery(func(iq *gotgbot.InlineQuery) bool {
		return true
	}, h.inlineQuery))
	d.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		if msg.Text == "" {
			return false
		}
		return twitterURLRegex.MatchString(msg.Text)
	}, h.messageHandler))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.ChainCallbackPrefix)
	}, h.chainCallback))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.DeleteCallbackPrefix)
	}, h.deleteCallback))
}
