package chain

import (
	"context"
	"errors"
	"testing"

	"twitterx-bot/internal/twitterxapi"
)

// ptr is a helper to create string pointers for tests.
func ptr(s string) *string {
	return &s
}

// mockFetcher implements TweetFetcher for testing.
type mockFetcher struct {
	tweets map[string]*twitterxapi.Tweet // key: "username/tweetID"
	err    error
}

func (m *mockFetcher) GetTweet(ctx context.Context, username, tweetID string) (*twitterxapi.Tweet, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := username + "/" + tweetID
	tweet, ok := m.tweets[key]
	if !ok {
		return nil, errors.New("tweet not found")
	}
	return tweet, nil
}

func TestBuildChain_NilTweet(t *testing.T) {
	fetcher := &mockFetcher{}
	chain, err := BuildChain(context.Background(), fetcher, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain != nil {
		t.Errorf("expected nil chain, got %v", chain)
	}
}

func TestBuildChain_SingleTweet(t *testing.T) {
	tweet := &twitterxapi.Tweet{
		ID:   "1",
		Text: "Hello world",
		Author: twitterxapi.Author{
			ScreenName: "alice",
		},
	}

	fetcher := &mockFetcher{}
	chain, err := BuildChain(context.Background(), fetcher, tweet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain) != 1 {
		t.Fatalf("expected 1 item in chain, got %d", len(chain))
	}
	if chain[0].Tweet.ID != "1" {
		t.Errorf("expected tweet ID '1', got %q", chain[0].Tweet.ID)
	}
	if chain[0].Type != ChainTypeRoot {
		t.Errorf("expected type Root, got %v", chain[0].Type)
	}
}

func TestBuildChain_ReplyToParent(t *testing.T) {
	parentTweet := &twitterxapi.Tweet{
		ID:   "1",
		Text: "Parent tweet",
		Author: twitterxapi.Author{
			ScreenName: "alice",
		},
	}

	childTweet := &twitterxapi.Tweet{
		ID:               "2",
		Text:             "Reply tweet",
		ReplyingToStatus: ptr("1"),
		ReplyingTo:       ptr("alice"),
		Author: twitterxapi.Author{
			ScreenName: "bob",
		},
	}

	fetcher := &mockFetcher{
		tweets: map[string]*twitterxapi.Tweet{
			"alice/1": parentTweet,
		},
	}

	chain, err := BuildChain(context.Background(), fetcher, childTweet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain) != 2 {
		t.Fatalf("expected 2 items in chain, got %d", len(chain))
	}

	// First item should be the parent (oldest)
	if chain[0].Tweet.ID != "1" {
		t.Errorf("expected first tweet ID '1', got %q", chain[0].Tweet.ID)
	}
	if chain[0].Type != ChainTypeReply {
		t.Errorf("expected first type Reply, got %v", chain[0].Type)
	}

	// Second item should be the child (root)
	if chain[1].Tweet.ID != "2" {
		t.Errorf("expected second tweet ID '2', got %q", chain[1].Tweet.ID)
	}
	if chain[1].Type != ChainTypeRoot {
		t.Errorf("expected second type Root, got %v", chain[1].Type)
	}
}

func TestBuildChain_MultiLevelReplies(t *testing.T) {
	grandparent := &twitterxapi.Tweet{
		ID:   "1",
		Text: "Grandparent",
		Author: twitterxapi.Author{
			ScreenName: "alice",
		},
	}

	parent := &twitterxapi.Tweet{
		ID:               "2",
		Text:             "Parent",
		ReplyingToStatus: ptr("1"),
		ReplyingTo:       ptr("alice"),
		Author: twitterxapi.Author{
			ScreenName: "bob",
		},
	}

	child := &twitterxapi.Tweet{
		ID:               "3",
		Text:             "Child",
		ReplyingToStatus: ptr("2"),
		ReplyingTo:       ptr("bob"),
		Author: twitterxapi.Author{
			ScreenName: "charlie",
		},
	}

	fetcher := &mockFetcher{
		tweets: map[string]*twitterxapi.Tweet{
			"alice/1": grandparent,
			"bob/2":   parent,
		},
	}

	chain, err := BuildChain(context.Background(), fetcher, child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain) != 3 {
		t.Fatalf("expected 3 items in chain, got %d", len(chain))
	}

	// Verify order: grandparent -> parent -> child
	expectedIDs := []string{"1", "2", "3"}
	expectedTypes := []ChainType{ChainTypeReply, ChainTypeReply, ChainTypeRoot}

	for i, item := range chain {
		if item.Tweet.ID != expectedIDs[i] {
			t.Errorf("chain[%d]: expected ID %q, got %q", i, expectedIDs[i], item.Tweet.ID)
		}
		if item.Type != expectedTypes[i] {
			t.Errorf("chain[%d]: expected type %v, got %v", i, expectedTypes[i], item.Type)
		}
	}
}

func TestBuildChain_ParentWithQuote(t *testing.T) {
	quotedTweet := &twitterxapi.Tweet{
		ID:   "0",
		Text: "Quoted tweet",
		Author: twitterxapi.Author{
			ScreenName: "dave",
		},
	}

	parentTweet := &twitterxapi.Tweet{
		ID:    "1",
		Text:  "Parent with quote",
		Quote: quotedTweet,
		Author: twitterxapi.Author{
			ScreenName: "alice",
		},
	}

	childTweet := &twitterxapi.Tweet{
		ID:               "2",
		Text:             "Reply",
		ReplyingToStatus: ptr("1"),
		ReplyingTo:       ptr("alice"),
		Author: twitterxapi.Author{
			ScreenName: "bob",
		},
	}

	fetcher := &mockFetcher{
		tweets: map[string]*twitterxapi.Tweet{
			"alice/1": parentTweet,
		},
	}

	chain, err := BuildChain(context.Background(), fetcher, childTweet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chain) != 3 {
		t.Fatalf("expected 3 items in chain, got %d", len(chain))
	}

	// Order: quoted -> parent -> child
	expectedIDs := []string{"0", "1", "2"}
	expectedTypes := []ChainType{ChainTypeQuote, ChainTypeReply, ChainTypeRoot}

	for i, item := range chain {
		if item.Tweet.ID != expectedIDs[i] {
			t.Errorf("chain[%d]: expected ID %q, got %q", i, expectedIDs[i], item.Tweet.ID)
		}
		if item.Type != expectedTypes[i] {
			t.Errorf("chain[%d]: expected type %v, got %v", i, expectedTypes[i], item.Type)
		}
	}
}

func TestBuildChain_ParentFetchError(t *testing.T) {
	childTweet := &twitterxapi.Tweet{
		ID:               "2",
		Text:             "Reply to deleted tweet",
		ReplyingToStatus: ptr("1"),
		ReplyingTo:       ptr("alice"),
		Author: twitterxapi.Author{
			ScreenName: "bob",
		},
	}

	// Fetcher returns error for parent
	fetcher := &mockFetcher{
		tweets: map[string]*twitterxapi.Tweet{},
	}

	chain, err := BuildChain(context.Background(), fetcher, childTweet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still return the child tweet
	if len(chain) != 1 {
		t.Fatalf("expected 1 item in chain, got %d", len(chain))
	}
	if chain[0].Tweet.ID != "2" {
		t.Errorf("expected tweet ID '2', got %q", chain[0].Tweet.ID)
	}
}

func TestBuildChain_MissingReplyUsername(t *testing.T) {
	// Tweet has ReplyingToStatus but no ReplyingTo
	tweet := &twitterxapi.Tweet{
		ID:               "2",
		Text:             "Reply without username",
		ReplyingToStatus: ptr("1"),
		Author: twitterxapi.Author{
			ScreenName: "bob",
		},
	}

	fetcher := &mockFetcher{}
	chain, err := BuildChain(context.Background(), fetcher, tweet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return just the tweet since we can't fetch parent
	if len(chain) != 1 {
		t.Fatalf("expected 1 item in chain, got %d", len(chain))
	}
}

func TestGetQuotedTweets_NilTweet(t *testing.T) {
	quotes := GetQuotedTweets(nil)
	if quotes != nil {
		t.Errorf("expected nil, got %v", quotes)
	}
}

func TestGetQuotedTweets_NoQuotes(t *testing.T) {
	tweet := &twitterxapi.Tweet{
		ID:   "1",
		Text: "No quotes here",
	}
	quotes := GetQuotedTweets(tweet)
	if len(quotes) != 0 {
		t.Errorf("expected 0 quotes, got %d", len(quotes))
	}
}

func TestGetQuotedTweets_SingleQuote(t *testing.T) {
	quoted := &twitterxapi.Tweet{
		ID:   "1",
		Text: "Quoted",
	}
	tweet := &twitterxapi.Tweet{
		ID:    "2",
		Text:  "Quote tweet",
		Quote: quoted,
	}

	quotes := GetQuotedTweets(tweet)
	if len(quotes) != 1 {
		t.Fatalf("expected 1 quote, got %d", len(quotes))
	}
	if quotes[0].ID != "1" {
		t.Errorf("expected quote ID '1', got %q", quotes[0].ID)
	}
}

func TestGetQuotedTweets_NestedQuotes(t *testing.T) {
	innerQuote := &twitterxapi.Tweet{
		ID:   "1",
		Text: "Inner quote",
	}
	outerQuote := &twitterxapi.Tweet{
		ID:    "2",
		Text:  "Outer quote",
		Quote: innerQuote,
	}
	tweet := &twitterxapi.Tweet{
		ID:    "3",
		Text:  "Main tweet",
		Quote: outerQuote,
	}

	quotes := GetQuotedTweets(tweet)
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}
	if quotes[0].ID != "2" {
		t.Errorf("expected first quote ID '2', got %q", quotes[0].ID)
	}
	if quotes[1].ID != "1" {
		t.Errorf("expected second quote ID '1', got %q", quotes[1].ID)
	}
}
