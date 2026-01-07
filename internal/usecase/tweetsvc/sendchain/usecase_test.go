package sendchain

import (
	"context"
	"errors"
	"testing"

	"twitterx-bot/internal/chain"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

type chainFetcher struct {
	tweets map[string]*twitterxapi.Tweet
	errs   map[string]error

	calls int
	got   []fetchCall
}

type fetchCall struct {
	username string
	tweetID  string
}

func (f *chainFetcher) GetTweet(_ context.Context, username, tweetID string) (*twitterxapi.Tweet, error) {
	f.calls++
	f.got = append(f.got, fetchCall{username: username, tweetID: tweetID})
	key := username + ":" + tweetID
	if f.errs != nil {
		if err, ok := f.errs[key]; ok {
			return nil, err
		}
	}
	if f.tweets != nil {
		if tw, ok := f.tweets[key]; ok {
			return tw, nil
		}
	}
	return nil, errors.New("not found")
}

type fakeSender struct {
	calls int
	err   error

	got struct {
		chatID       int64
		replyToMsgID int64
		items        []chain.ChainItem
		opts         *tweet.SendChainResponseOpts
	}
}

func (s *fakeSender) SendChainResponse(chatID int64, items []chain.ChainItem, replyToMsgID int64, opts *tweet.SendChainResponseOpts) error {
	s.calls++
	s.got.chatID = chatID
	s.got.replyToMsgID = replyToMsgID
	s.got.items = items
	s.got.opts = opts
	return s.err
}

func strPtr(s string) *string {
	return &s
}

func TestUseCaseSendChainSuccess(t *testing.T) {
	parent := &twitterxapi.Tweet{
		ID:   "100",
		Text: "parent",
		URL:  "https://x.com/parent/status/100",
		Author: twitterxapi.Author{
			Name:       "Parent",
			ScreenName: "parent",
		},
	}
	root := &twitterxapi.Tweet{
		ID:               "200",
		Text:             "root",
		URL:              "https://x.com/alice/status/200",
		ReplyingTo:       strPtr("parentUser"),
		ReplyingToStatus: strPtr("100"),
		Author: twitterxapi.Author{
			Name:       "Alice",
			ScreenName: "alice",
		},
	}

	fetcher := &chainFetcher{
		tweets: map[string]*twitterxapi.Tweet{
			"alice:200":      root,
			"parentUser:100": parent,
		},
	}
	sender := &fakeSender{}
	uc := New(fetcher, sender)

	err := uc.SendChain(context.Background(), 11, 22, "alice", "200", "@req")
	if err != nil {
		t.Fatalf("SendChain() error = %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d, want 1", sender.calls)
	}
	if sender.got.chatID != 11 || sender.got.replyToMsgID != 22 {
		t.Fatalf("sender args = (%d, %d), want (%d, %d)", sender.got.chatID, sender.got.replyToMsgID, 11, 22)
	}
	if sender.got.opts == nil || sender.got.opts.RequesterUsername != "@req" {
		t.Fatalf("requester username not set")
	}
	if len(sender.got.items) != 2 {
		t.Fatalf("items len = %d, want 2", len(sender.got.items))
	}
	if sender.got.items[0].Tweet == nil || sender.got.items[0].Tweet.ID != "100" {
		t.Fatalf("first item id = %v, want 100", sender.got.items[0].Tweet)
	}
	if sender.got.items[1].Tweet == nil || sender.got.items[1].Tweet.ID != "200" {
		t.Fatalf("last item id = %v, want 200", sender.got.items[1].Tweet)
	}
}

func TestUseCaseSendChainParentMissing(t *testing.T) {
	root := &twitterxapi.Tweet{
		ID:               "200",
		Text:             "root",
		URL:              "https://x.com/alice/status/200",
		ReplyingTo:       strPtr("parentUser"),
		ReplyingToStatus: strPtr("100"),
	}

	fetcher := &chainFetcher{
		tweets: map[string]*twitterxapi.Tweet{
			"alice:200": root,
		},
		errs: map[string]error{
			"parentUser:100": errors.New("gone"),
		},
	}
	sender := &fakeSender{}
	uc := New(fetcher, sender)

	err := uc.SendChain(context.Background(), 11, 22, "alice", "200", "@req")
	if err != nil {
		t.Fatalf("SendChain() error = %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d, want 1", sender.calls)
	}
	if len(sender.got.items) != 1 {
		t.Fatalf("items len = %d, want 1", len(sender.got.items))
	}
	if sender.got.items[0].Tweet == nil || sender.got.items[0].Tweet.ID != "200" {
		t.Fatalf("item id = %v, want 200", sender.got.items[0].Tweet)
	}
}

func TestUseCaseSendChainFetchError(t *testing.T) {
	fetcher := &chainFetcher{
		errs: map[string]error{
			"alice:200": errors.New("boom"),
		},
	}
	sender := &fakeSender{}
	uc := New(fetcher, sender)

	err := uc.SendChain(context.Background(), 11, 22, "alice", "200", "@req")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrFetchTweet) {
		t.Fatalf("expected ErrFetchTweet, got %v", err)
	}
	if sender.calls != 0 {
		t.Fatalf("sender should not be called")
	}
}

func TestUseCaseSendChainSenderError(t *testing.T) {
	fetcher := &chainFetcher{
		tweets: map[string]*twitterxapi.Tweet{
			"alice:200": {ID: "200", Text: "root", URL: "https://x.com/alice/status/200"},
		},
	}
	sender := &fakeSender{err: errors.New("send fail")}
	uc := New(fetcher, sender)

	err := uc.SendChain(context.Background(), 11, 22, "alice", "200", "@req")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrSendChain) {
		t.Fatalf("expected ErrSendChain, got %v", err)
	}
}
