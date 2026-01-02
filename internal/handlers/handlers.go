package handlers

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/twitterxapi"
)

var twitterURLRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?(?:twitter\.com|x\.com)/([^/]+)/status/(\d+)`)

type Handlers struct {
	log *logger.Logger
	api *twitterxapi.Client
}

const inlineQueryTimeout = 10 * time.Second

func Register(d *ext.Dispatcher, log *logger.Logger, api *twitterxapi.Client) {
	if api == nil {
		api = twitterxapi.NewClient("")
	}
	h := &Handlers{log: log, api: api}

	d.AddHandler(handlers.NewCommand("start", start))
	d.AddHandler(handlers.NewInlineQuery(func(iq *gotgbot.InlineQuery) bool {
		return true
	}, h.inlineQuery))
}

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Hi! Send me any text and I will echo it back.", &gotgbot.SendMessageOpts{})
	return err
}

func (h *Handlers) inlineQuery(b *gotgbot.Bot, ctx *ext.Context) error {
	query := strings.TrimSpace(ctx.InlineQuery.Query)
	h.log.Debug("received inline query: %s", query)

	matches := twitterURLRegex.FindStringSubmatch(query)
	if matches == nil {
		h.log.Debug("no twitter URL found in query")
		_, err := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		return err
	}

	username := matches[1]
	tweetID := matches[2]

	h.log.Info("twitter URL: %s", query)
	h.log.Info("twitter URL parsed - username: %s, tweet_id: %s", username, tweetID)
	h.log.Debug("full match details - query: %s, username: %s, tweet_id: %s", query, username, tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), inlineQueryTimeout)
	defer cancel()

	tweet, err := h.api.GetTweet(reqCtx, username, tweetID)
	if err != nil {
		h.log.Error("failed to fetch tweet %s for %s: %v", tweetID, username, err)
		_, answerErr := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		if answerErr != nil {
			return answerErr
		}
		return nil
	}

	result, ok := buildInlineResult(tweet, tweetID)
	if !ok {
		h.log.Error("no suitable inline result for tweet %s", tweetID)
		_, answerErr := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		if answerErr != nil {
			return answerErr
		}
		return nil
	}

	results := []gotgbot.InlineQueryResult{result}

	_, err = ctx.InlineQuery.Answer(b, results, &gotgbot.AnswerInlineQueryOpts{
		CacheTime:  0,
		IsPersonal: true,
	})
	return err
}
