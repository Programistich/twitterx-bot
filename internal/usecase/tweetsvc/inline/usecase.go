package inline

import (
	"context"
	"errors"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

var (
	ErrFetchTweet  = errors.New("fetch tweet")
	ErrBuildInline = errors.New("build inline result")
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error)
}

// UseCase handles building inline query results from tweets.
type UseCase struct {
	Fetcher TweetFetcher
}

// New creates a new inline UseCase.
func New(fetcher TweetFetcher) *UseCase {
	return &UseCase{Fetcher: fetcher}
}

// BuildInlineResult fetches a tweet and builds an inline query result.
func (uc *UseCase) BuildInlineResult(ctx context.Context, username, tweetID string) (gotgbot.InlineQueryResult, bool, error) {
	if uc == nil {
		return nil, false, fmt.Errorf("inline usecase: %w", ErrBuildInline)
	}
	if uc.Fetcher == nil {
		return nil, false, fmt.Errorf("inline usecase: %w", ErrFetchTweet)
	}

	tw, err := uc.Fetcher.GetTweet(ctx, username, tweetID)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFetchTweet, err)
	}

	result, ok := tweet.BuildInlineResult(tw, tweetID)
	if !ok {
		return nil, false, nil
	}

	return result, true, nil
}
