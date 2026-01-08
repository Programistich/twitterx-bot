package inline

import (
	"context"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/twitterurl"
	inlineuc "twitterx-bot/internal/usecase/tweetsvc/inline"
)

// Handler wraps the dependencies needed to respond to inline queries.
type Handler struct {
	log     *logger.Logger
	uc      *inlineuc.UseCase
	timeout time.Duration
}

// New creates a handler for inline queries.
func New(log *logger.Logger, uc *inlineuc.UseCase, timeout time.Duration) *Handler {
	return &Handler{log: log, uc: uc, timeout: timeout}
}

// Handle answers inline queries that contain Twitter URLs.
func (h *Handler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	query := strings.TrimSpace(ctx.InlineQuery.Query)
	log := h.log.With("component", "inline")
	if ctx.InlineQuery != nil {
		log = log.With("inline_query_id", ctx.InlineQuery.Id)
	}
	if ctx.EffectiveUser != nil {
		log = log.With("user_id", ctx.EffectiveUser.Id, "username", ctx.EffectiveUser.Username)
	}
	log.Debug("inline query received", "query", query)

	username, tweetID, ok := twitterurl.ParseTweetURL(query)
	if !ok {
		log.Debug("inline query ignored: no tweet url")
		_, err := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		return err
	}

	log.Info("tweet url parsed", "tweet_username", username, "tweet_id", tweetID)
	log.Debug("inline query details", "query", query, "tweet_username", username, "tweet_id", tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	result, ok, err := h.uc.BuildInlineResult(reqCtx, username, tweetID)
	if err != nil {
		log.Error("build inline result failed", "tweet_username", username, "tweet_id", tweetID, "err", err)
		_, answerErr := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		if answerErr != nil {
			return answerErr
		}
		return nil
	}

	if !ok {
		log.Warn("no suitable inline result", "tweet_id", tweetID)
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
	if err == nil {
		log.Info("inline result sent", "tweet_id", tweetID)
	}
	return err
}
