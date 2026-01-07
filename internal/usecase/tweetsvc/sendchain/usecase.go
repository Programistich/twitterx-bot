package sendchain

import (
	"context"
	"errors"
	"fmt"

	"twitterx-bot/internal/chain"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

var (
	ErrFetchTweet = errors.New("fetch tweet")
	ErrBuildChain = errors.New("build chain")
	ErrSendChain  = errors.New("send chain")
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error)
}

// ChainSender sends chain responses to Telegram.
type ChainSender interface {
	SendChainResponse(chatID int64, items []chain.ChainItem, replyToMsgID int64, opts *tweet.SendChainResponseOpts) error
}

// UseCase handles sending tweet chains to Telegram.
type UseCase struct {
	Fetcher TweetFetcher
	Sender  ChainSender
}

// New creates a new sendchain UseCase.
func New(fetcher TweetFetcher, sender ChainSender) *UseCase {
	return &UseCase{Fetcher: fetcher, Sender: sender}
}

// SendChain fetches a tweet and sends the full reply chain to the chat.
func (uc *UseCase) SendChain(ctx context.Context, chatID, replyToMsgID int64, username, tweetID, requester string) error {
	if uc == nil {
		return fmt.Errorf("sendchain usecase: %w", ErrSendChain)
	}
	if uc.Fetcher == nil {
		return fmt.Errorf("sendchain usecase: %w", ErrFetchTweet)
	}
	if uc.Sender == nil {
		return fmt.Errorf("sendchain usecase: %w", ErrSendChain)
	}

	tw, err := uc.Fetcher.GetTweet(ctx, username, tweetID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchTweet, err)
	}

	items, err := chain.BuildChain(ctx, uc.Fetcher, tw)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBuildChain, err)
	}

	opts := &tweet.SendChainResponseOpts{
		RequesterUsername: requester,
	}
	if err := uc.Sender.SendChainResponse(chatID, items, replyToMsgID, opts); err != nil {
		return fmt.Errorf("%w: %v", ErrSendChain, err)
	}

	return nil
}
