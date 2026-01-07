package tweetsvc

import (
	"context"
	"fmt"

	"twitterx-bot/internal/telegram/tweet"
)

// SendTweet fetches a tweet and sends it to the chat, replying to replyToMsgID.
func (s *Service) SendTweet(ctx context.Context, chatID, replyToMsgID int64, username, tweetID, requester string) error {
	if s == nil {
		return fmt.Errorf("tweet service: %w", ErrSendTweet)
	}
	if s.Fetcher == nil {
		return fmt.Errorf("tweet service: %w", ErrFetchTweet)
	}
	if s.Sender == nil {
		return fmt.Errorf("tweet service: %w", ErrSendTweet)
	}

	tw, err := s.Fetcher.GetTweet(ctx, username, tweetID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchTweet, err)
	}

	var keyboardOpts *tweet.KeyboardOpts
	if tw != nil && tw.ReplyingToStatus != nil {
		keyboardOpts = &tweet.KeyboardOpts{
			ShowChainButton: true,
			ChainUsername:   username,
			ChainTweetID:    tweetID,
		}
	}

	opts := &tweet.SendResponseOpts{
		ReplyMarkup:       tweet.BuildKeyboard(replyToMsgID, keyboardOpts),
		RequesterUsername: requester,
	}

	if err := s.Sender.SendTweet(ctx, chatID, replyToMsgID, tw, opts); err != nil {
		return fmt.Errorf("%w: %v", ErrSendTweet, err)
	}

	return nil
}
