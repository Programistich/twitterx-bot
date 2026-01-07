package handlers

import (
	"context"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/tweet"
	"twitterx-bot/internal/twitterurl"
)

func (h *Handlers) inlineQuery(b *gotgbot.Bot, ctx *ext.Context) error {
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

	reqCtx, cancel := context.WithTimeout(context.Background(), inlineQueryTimeout)
	defer cancel()

	tw, err := h.api.GetTweet(reqCtx, username, tweetID)
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

	result, ok := tweet.BuildInlineResult(tw, tweetID)
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
