package tweetsvc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/twitterxapi"
)

type fakeFetcher struct {
	tweet *twitterxapi.Tweet
	err   error

	calls int
	got   struct {
		username string
		tweetID  string
	}
}

func (f *fakeFetcher) GetTweet(_ context.Context, username, tweetID string) (*twitterxapi.Tweet, error) {
	f.calls++
	f.got.username = username
	f.got.tweetID = tweetID
	return f.tweet, f.err
}

type fakeBot struct {
	videoCalls      int
	photoCalls      int
	mediaGroupCalls int
	messageCalls    int

	err error

	lastVideoOpts      *gotgbot.SendVideoOpts
	lastPhotoOpts      *gotgbot.SendPhotoOpts
	lastMedia          []gotgbot.InputMedia
	lastMediaGroupOpts *gotgbot.SendMediaGroupOpts
	lastMessageText    string
	lastMessageOpts    *gotgbot.SendMessageOpts
}

func (b *fakeBot) SendVideo(_ int64, _ gotgbot.InputFileOrString, opts *gotgbot.SendVideoOpts) (*gotgbot.Message, error) {
	b.videoCalls++
	b.lastVideoOpts = opts
	if b.err != nil {
		return nil, b.err
	}
	return &gotgbot.Message{MessageId: 1}, nil
}

func (b *fakeBot) SendPhoto(_ int64, _ gotgbot.InputFileOrString, opts *gotgbot.SendPhotoOpts) (*gotgbot.Message, error) {
	b.photoCalls++
	b.lastPhotoOpts = opts
	if b.err != nil {
		return nil, b.err
	}
	return &gotgbot.Message{MessageId: 1}, nil
}

func (b *fakeBot) SendMediaGroup(_ int64, media []gotgbot.InputMedia, opts *gotgbot.SendMediaGroupOpts) ([]gotgbot.Message, error) {
	b.mediaGroupCalls++
	b.lastMedia = media
	b.lastMediaGroupOpts = opts
	if b.err != nil {
		return nil, b.err
	}
	return []gotgbot.Message{{MessageId: 1}}, nil
}

func (b *fakeBot) SendMessage(_ int64, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error) {
	b.messageCalls++
	b.lastMessageText = text
	b.lastMessageOpts = opts
	if b.err != nil {
		return nil, b.err
	}
	return &gotgbot.Message{MessageId: 1}, nil
}

func TestServiceSendTweetSelectsVideo(t *testing.T) {
	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:               "123",
			Text:             "hello",
			URL:              "https://x.com/user/status/123",
			ReplyingToStatus: strPtr("1"),
			Media: &twitterxapi.Media{
				Videos: []twitterxapi.Video{{URL: "https://video/1.mp4", Width: 640, Height: 480}},
				Photos: []twitterxapi.Photo{{URL: "https://img/1.jpg"}},
			},
		},
	}
	bot := &fakeBot{}
	svc := New(fetcher, tweet.Sender{Bot: bot})

	err := svc.SendTweet(context.Background(), 1001, 42, "user", "123", "@req")
	if err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if fetcher.calls != 1 {
		t.Fatalf("fetcher calls = %d, want 1", fetcher.calls)
	}
	if fetcher.got.username != "user" || fetcher.got.tweetID != "123" {
		t.Fatalf("fetcher args = (%q, %q), want (%q, %q)", fetcher.got.username, fetcher.got.tweetID, "user", "123")
	}
	if bot.videoCalls != 1 || bot.photoCalls != 0 || bot.mediaGroupCalls != 0 || bot.messageCalls != 0 {
		t.Fatalf("calls: video=%d photo=%d media=%d msg=%d, want video only", bot.videoCalls, bot.photoCalls, bot.mediaGroupCalls, bot.messageCalls)
	}
	if bot.lastVideoOpts == nil || bot.lastVideoOpts.ReplyParameters == nil || bot.lastVideoOpts.ReplyParameters.MessageId != 42 {
		t.Fatalf("reply parameters not set")
	}
	if bot.lastVideoOpts.Caption == "" || !strings.Contains(bot.lastVideoOpts.Caption, "@req") {
		t.Fatalf("caption missing requester info")
	}
	if bot.lastVideoOpts.ReplyMarkup == nil {
		t.Fatalf("reply markup missing")
	}
	videoMarkup, ok := bot.lastVideoOpts.ReplyMarkup.(*gotgbot.InlineKeyboardMarkup)
	if !ok || videoMarkup == nil {
		t.Fatalf("reply markup type = %T, want InlineKeyboardMarkup", bot.lastVideoOpts.ReplyMarkup)
	}
	if tweet.FindChainButton(videoMarkup) == "" {
		t.Fatalf("chain button missing")
	}
}

func TestServiceSendTweetSelectsPhoto(t *testing.T) {
	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "321",
			Text: "photo",
			URL:  "https://x.com/user/status/321",
			Media: &twitterxapi.Media{
				Photos: []twitterxapi.Photo{{URL: "https://img/1.jpg"}},
			},
		},
	}
	bot := &fakeBot{}
	svc := New(fetcher, tweet.Sender{Bot: bot})

	err := svc.SendTweet(context.Background(), 10, 7, "user", "321", "")
	if err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if bot.photoCalls != 1 || bot.videoCalls != 0 || bot.mediaGroupCalls != 0 || bot.messageCalls != 0 {
		t.Fatalf("calls: video=%d photo=%d media=%d msg=%d, want photo only", bot.videoCalls, bot.photoCalls, bot.mediaGroupCalls, bot.messageCalls)
	}
	if bot.lastPhotoOpts == nil || bot.lastPhotoOpts.ReplyParameters == nil || bot.lastPhotoOpts.ReplyParameters.MessageId != 7 {
		t.Fatalf("reply parameters not set")
	}
	if bot.lastPhotoOpts.ReplyMarkup == nil {
		t.Fatalf("reply markup missing")
	}
	photoMarkup, ok := bot.lastPhotoOpts.ReplyMarkup.(*gotgbot.InlineKeyboardMarkup)
	if !ok || photoMarkup == nil {
		t.Fatalf("reply markup type = %T, want InlineKeyboardMarkup", bot.lastPhotoOpts.ReplyMarkup)
	}
	if tweet.FindChainButton(photoMarkup) != "" {
		t.Fatalf("chain button unexpectedly present")
	}
}

func TestServiceSendTweetSelectsMediaGroup(t *testing.T) {
	photos := make([]twitterxapi.Photo, 0, 12)
	for i := 0; i < 12; i++ {
		photos = append(photos, twitterxapi.Photo{URL: "https://img/" + string(rune('a'+i)) + ".jpg"})
	}

	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "555",
			Text: "multi",
			URL:  "https://x.com/user/status/555",
			Media: &twitterxapi.Media{
				Photos: photos,
			},
		},
	}
	bot := &fakeBot{}
	svc := New(fetcher, tweet.Sender{Bot: bot})

	err := svc.SendTweet(context.Background(), 10, 7, "user", "555", "@req")
	if err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if bot.mediaGroupCalls != 1 || bot.videoCalls != 0 || bot.photoCalls != 0 || bot.messageCalls != 0 {
		t.Fatalf("calls: video=%d photo=%d media=%d msg=%d, want media only", bot.videoCalls, bot.photoCalls, bot.mediaGroupCalls, bot.messageCalls)
	}
	if len(bot.lastMedia) != tweet.MaxMediaGroupSize {
		t.Fatalf("media group size = %d, want %d", len(bot.lastMedia), tweet.MaxMediaGroupSize)
	}
	if len(bot.lastMedia) > 0 {
		if first, ok := bot.lastMedia[0].(gotgbot.InputMediaPhoto); ok {
			if first.Caption == "" {
				t.Fatalf("first media caption empty")
			}
		} else {
			t.Fatalf("first media type = %T, want InputMediaPhoto", bot.lastMedia[0])
		}
	}
}

func TestServiceSendTweetSelectsText(t *testing.T) {
	fetcher := &fakeFetcher{
		tweet: &twitterxapi.Tweet{
			ID:   "777",
			Text: "just text",
			URL:  "https://x.com/user/status/777",
		},
	}
	bot := &fakeBot{}
	svc := New(fetcher, tweet.Sender{Bot: bot})

	err := svc.SendTweet(context.Background(), 10, 7, "user", "777", "@req")
	if err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if bot.messageCalls != 1 || bot.videoCalls != 0 || bot.photoCalls != 0 || bot.mediaGroupCalls != 0 {
		t.Fatalf("calls: video=%d photo=%d media=%d msg=%d, want msg only", bot.videoCalls, bot.photoCalls, bot.mediaGroupCalls, bot.messageCalls)
	}
	if bot.lastMessageText == "" {
		t.Fatalf("message text empty")
	}
}

func TestServiceSendTweetFetcherError(t *testing.T) {
	fetcher := &fakeFetcher{err: errors.New("boom")}
	bot := &fakeBot{}
	svc := New(fetcher, tweet.Sender{Bot: bot})

	err := svc.SendTweet(context.Background(), 10, 7, "user", "321", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrFetchTweet) {
		t.Fatalf("expected ErrFetchTweet, got %v", err)
	}
	if bot.videoCalls+bot.photoCalls+bot.mediaGroupCalls+bot.messageCalls != 0 {
		t.Fatalf("sender should not be called")
	}
}

func TestServiceSendTweetSenderError(t *testing.T) {
	fetcher := &fakeFetcher{tweet: &twitterxapi.Tweet{ID: "123", Text: "text", URL: "https://x.com/u/s/123"}}
	bot := &fakeBot{err: errors.New("send fail")}
	svc := New(fetcher, tweet.Sender{Bot: bot})

	err := svc.SendTweet(context.Background(), 10, 7, "user", "123", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrSendTweet) {
		t.Fatalf("expected ErrSendTweet, got %v", err)
	}
}

func TestServiceSendTweetMissingDeps(t *testing.T) {
	svc := &Service{}
	if err := svc.SendTweet(context.Background(), 1, 1, "u", "t", ""); err == nil {
		t.Fatalf("expected error for missing deps")
	}
}

func strPtr(s string) *string {
	return &s
}
