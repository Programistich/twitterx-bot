package tweetsvc

import (
	"context"
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/telegram/tweet"
)

// BuildInlineResult fetches a tweet and builds an inline query result.
func (s *Service) BuildInlineResult(ctx context.Context, username, tweetID string) (gotgbot.InlineQueryResult, bool, error) {
	if s == nil {
		return nil, false, fmt.Errorf("tweet service: %w", ErrBuildInline)
	}
	if s.Fetcher == nil {
		return nil, false, fmt.Errorf("tweet service: %w", ErrFetchTweet)
	}

	tw, err := s.Fetcher.GetTweet(ctx, username, tweetID)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrFetchTweet, err)
	}

	result, ok := tweet.BuildInlineResult(tw, tweetID)
	if !ok {
		return nil, false, nil
	}

	return result, true, nil
}
