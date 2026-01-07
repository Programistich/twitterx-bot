package sendtweet

import (
	"context"
	"errors"
	"fmt"

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

// UseCase handles sending tweets to Telegram.
type UseCase struct {
	Fetcher TweetFetcher
	Sender  TweetSender
}

// New creates a new sendtweet UseCase.
func New(fetcher TweetFetcher, sender TweetSender) *UseCase {
	return &UseCase{Fetcher: fetcher, Sender: sender}
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
