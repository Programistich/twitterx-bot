package tweetsvc

import (
	"context"

	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error)
}

// TweetSender sends tweet responses to Telegram.
type TweetSender interface {
	SendTweet(ctx context.Context, chatID, replyToMsgID int64, tweet *twitterxapi.Tweet, opts *tweet.SendResponseOpts) error
}

// Service holds tweet-related use cases.
type Service struct {
	Fetcher TweetFetcher
	Sender  TweetSender
}

// New creates a new tweet service.
func New(fetcher TweetFetcher, sender TweetSender) *Service {
	return &Service{Fetcher: fetcher, Sender: sender}
}
