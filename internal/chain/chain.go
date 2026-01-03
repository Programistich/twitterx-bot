package chain

import (
	"context"

	"twitterx-bot/internal/twitterxapi"
)

// TweetFetcher defines the interface for fetching tweets.
// This allows mocking in tests.
type TweetFetcher interface {
	GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error)
}

// ChainType indicates the relationship type in the chain.
type ChainType string

const (
	ChainTypeReply ChainType = "reply"
	ChainTypeQuote ChainType = "quote"
	ChainTypeRoot  ChainType = "root"
)

// ChainItem represents a single tweet in the chain with its relationship type.
type ChainItem struct {
	Tweet *twitterxapi.Tweet
	Type  ChainType
}

// BuildChain fetches the full chain of tweets for the given tweet.
// It returns tweets in chronological order (oldest first).
// The chain includes:
// - Parent tweets (if this tweet is a reply)
// - Quoted tweets (embedded in the response)
// - The original tweet itself (last in the chain)
func BuildChain(ctx context.Context, fetcher TweetFetcher, tweet *twitterxapi.Tweet) ([]ChainItem, error) {
	if tweet == nil {
		return nil, nil
	}

	var chain []ChainItem

	// First, fetch the reply chain (parent tweets)
	replyChain, err := buildReplyChain(ctx, fetcher, tweet)
	if err != nil {
		return nil, err
	}
	chain = append(chain, replyChain...)

	// Add the original tweet as root
	chain = append(chain, ChainItem{
		Tweet: tweet,
		Type:  ChainTypeRoot,
	})

	return chain, nil
}

// buildReplyChain recursively fetches parent tweets.
// Returns tweets in chronological order (oldest first).
func buildReplyChain(ctx context.Context, fetcher TweetFetcher, tweet *twitterxapi.Tweet) ([]ChainItem, error) {
	if tweet.ReplyingToStatus == nil || tweet.ReplyingTo == nil {
		return nil, nil
	}

	// Fetch the parent tweet
	parent, err := fetcher.GetTweet(ctx, *tweet.ReplyingTo, *tweet.ReplyingToStatus)
	if err != nil {
		// If we can't fetch the parent, we just return what we have
		// This can happen if the parent tweet is deleted or private
		return nil, nil
	}

	var chain []ChainItem

	// Recursively get the parent's chain first
	parentChain, err := buildReplyChain(ctx, fetcher, parent)
	if err != nil {
		return nil, err
	}
	chain = append(chain, parentChain...)

	// Add quoted tweet if present (before the parent)
	if parent.Quote != nil {
		chain = append(chain, ChainItem{
			Tweet: parent.Quote,
			Type:  ChainTypeQuote,
		})
	}

	// Add the parent as a reply reference
	chain = append(chain, ChainItem{
		Tweet: parent,
		Type:  ChainTypeReply,
	})

	return chain, nil
}

// GetQuotedTweets extracts all quoted tweets from a tweet chain.
// This includes the quoted tweet of the root tweet and any quoted tweets in replies.
func GetQuotedTweets(tweet *twitterxapi.Tweet) []*twitterxapi.Tweet {
	if tweet == nil {
		return nil
	}

	var quotes []*twitterxapi.Tweet
	if tweet.Quote != nil {
		quotes = append(quotes, tweet.Quote)
		// Recursively get quotes from the quoted tweet
		quotes = append(quotes, GetQuotedTweets(tweet.Quote)...)
	}

	return quotes
}
