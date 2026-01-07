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
	h.log.Debug("received inline query: %s", query)

	username, tweetID, ok := twitterurl.ParseTweetURL(query)
	if !ok {
		h.log.Debug("no twitter URL found in query")
		_, err := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		return err
	}

	h.log.Info("twitter URL: %s", query)
	h.log.Info("twitter URL parsed - username: %s, tweet_id: %s", username, tweetID)
	h.log.Debug("full match details - query: %s, username: %s, tweet_id: %s", query, username, tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	result, ok, err := h.uc.BuildInlineResult(reqCtx, username, tweetID)
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
