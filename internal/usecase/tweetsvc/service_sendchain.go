package tweetsvc

import (
	"context"
	"fmt"

	"twitterx-bot/internal/chain"
	"twitterx-bot/internal/telegram/tweet"
)

// SendChain fetches a tweet and sends the full reply chain to the chat.
func (s *Service) SendChain(ctx context.Context, chatID, replyToMsgID int64, username, tweetID, requester string) error {
	if s == nil {
		return fmt.Errorf("tweet service: %w", ErrSendChain)
	}
	if s.Fetcher == nil {
		return fmt.Errorf("tweet service: %w", ErrFetchTweet)
	}
	if s.Sender == nil {
		return fmt.Errorf("tweet service: %w", ErrSendChain)
	}

	tw, err := s.Fetcher.GetTweet(ctx, username, tweetID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchTweet, err)
	}

	items, err := chain.BuildChain(ctx, s.Fetcher, tw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBuildChain, err)
	}

	opts := &tweet.SendChainResponseOpts{
		RequesterUsername: requester,
	}
	if err := s.Sender.SendChainResponse(chatID, items, replyToMsgID, opts); err != nil {
		return fmt.Errorf("%w: %v", ErrSendChain, err)
	}

	return nil
}
