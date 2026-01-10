package sendtweet

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
	ErrSendTweet  = errors.New("send tweet")
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error)
}

// TweetSender sends tweet responses to Telegram.
type TweetSender interface {
	SendTweet(ctx context.Context, chatID, replyToMsgID int64, tweet *twitterxapi.Tweet, opts *tweet.SendResponseOpts) error
}

// ChainSender sends chain of tweets to Telegram.
type ChainSender interface {
	SendChainResponse(chatID int64, items []chain.ChainItem, replyToMsgID int64, opts *tweet.SendChainResponseOpts) error
}

// UseCase handles sending tweets to Telegram.
type UseCase struct {
	Fetcher     TweetFetcher
	Sender      TweetSender
	ChainSender ChainSender
}

// New creates a new sendtweet UseCase.
func New(fetcher TweetFetcher, sender TweetSender) *UseCase {
	return &UseCase{Fetcher: fetcher, Sender: sender}
}

// NewWithChain creates a new sendtweet UseCase with chain support.
func NewWithChain(fetcher TweetFetcher, sender TweetSender, chainSender ChainSender) *UseCase {
	return &UseCase{Fetcher: fetcher, Sender: sender, ChainSender: chainSender}
}

// SendTweet fetches a tweet and sends it to the chat, replying to replyToMsgID.
func (uc *UseCase) SendTweet(ctx context.Context, chatID, replyToMsgID int64, username, tweetID, requester string) error {
	if uc == nil {
		return fmt.Errorf("sendtweet usecase: %w", ErrSendTweet)
	}
	if uc.Fetcher == nil {
		return fmt.Errorf("sendtweet usecase: %w", ErrFetchTweet)
	}
	if uc.Sender == nil {
		return fmt.Errorf("sendtweet usecase: %w", ErrSendTweet)
	}

	tw, err := uc.Fetcher.GetTweet(ctx, username, tweetID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchTweet, err)
	}

	// If tweet has a quote and we have a chain sender, send as chain immediately
	if tw != nil && tw.Quote != nil && uc.ChainSender != nil {
		items := []chain.ChainItem{
			{Tweet: tw.Quote, Type: chain.ChainTypeQuote},
			{Tweet: tw, Type: chain.ChainTypeRoot},
		}
		opts := &tweet.SendChainResponseOpts{
			RequesterUsername: requester,
		}
		if err := uc.ChainSender.SendChainResponse(chatID, items, replyToMsgID, opts); err != nil {
			return fmt.Errorf("%w: %v", ErrSendTweet, err)
		}
		return nil
	}

	// Show chain button only for replies (not quotes)
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

	if err := uc.Sender.SendTweet(ctx, chatID, replyToMsgID, tw, opts); err != nil {
		return fmt.Errorf("%w: %v", ErrSendTweet, err)
	}

	return nil
}
