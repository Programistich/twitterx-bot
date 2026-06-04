package tweet

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/translation"
	"twitterx-bot/internal/twitterxapi"
)

// translateFakeBot records SendMessage calls for translation assertions.
type translateFakeBot struct {
	messageCalls int
	lastText     string
	lastOpts     *gotgbot.SendMessageOpts
}

func (b *translateFakeBot) SendVideo(_ int64, _ gotgbot.InputFileOrString, _ *gotgbot.SendVideoOpts) (*gotgbot.Message, error) {
	return &gotgbot.Message{MessageId: 1}, nil
}

func (b *translateFakeBot) SendPhoto(_ int64, _ gotgbot.InputFileOrString, _ *gotgbot.SendPhotoOpts) (*gotgbot.Message, error) {
	return &gotgbot.Message{MessageId: 1}, nil
}

func (b *translateFakeBot) SendMediaGroup(_ int64, _ []gotgbot.InputMedia, _ *gotgbot.SendMediaGroupOpts) ([]gotgbot.Message, error) {
	return []gotgbot.Message{{MessageId: 1}}, nil
}

func (b *translateFakeBot) SendMessage(_ int64, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error) {
	b.messageCalls++
	b.lastText = text
	b.lastOpts = opts
	return &gotgbot.Message{MessageId: int64(b.messageCalls)}, nil
}

// fakeTranslator returns a configured translation result.
type fakeTranslator struct {
	result *translation.Translation
	err    error
	calls  int
}

func (f *fakeTranslator) Translate(_ context.Context, text string, to translation.Language) (*translation.Translation, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if f.result != nil {
		return f.result, nil
	}
	return &translation.Translation{From: translation.Language{ISO: "en"}, To: to, Text: text}, nil
}

func textTweet(text string) *twitterxapi.Tweet {
	return &twitterxapi.Tweet{
		ID:     "123",
		Text:   text,
		URL:    "https://x.com/user/status/123",
		Author: twitterxapi.Author{Name: "User", ScreenName: "user"},
	}
}

func TestSendTweetSendsTranslationReply(t *testing.T) {
	bot := &translateFakeBot{}
	tr := &fakeTranslator{result: &translation.Translation{
		From: translation.Language{ISO: "en"},
		To:   translation.LangUkrainian,
		Text: "Привіт світ",
	}}
	s := Sender{Bot: bot, Translator: tr, Lang: "uk"}

	if err := s.SendTweet(context.Background(), 1001, 42, textTweet("Hello world"), nil); err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}

	if tr.calls != 1 {
		t.Fatalf("translator calls = %d, want 1", tr.calls)
	}
	// One message for the tweet text, one for the translation reply.
	if bot.messageCalls != 2 {
		t.Fatalf("message calls = %d, want 2 (tweet + translation)", bot.messageCalls)
	}
	if !strings.Contains(bot.lastText, "Привіт світ") {
		t.Fatalf("translation body missing translated text: %q", bot.lastText)
	}
	if bot.lastOpts == nil || bot.lastOpts.ReplyParameters == nil || bot.lastOpts.ReplyParameters.MessageId != 1 {
		t.Fatalf("translation reply parameters not set to primary message id")
	}
}

func TestSendTweetSkipsTranslationWhenSameLanguage(t *testing.T) {
	bot := &translateFakeBot{}
	tr := &fakeTranslator{result: &translation.Translation{
		From: translation.Language{ISO: "uk"},
		To:   translation.LangUkrainian,
		Text: "Привіт",
	}}
	s := Sender{Bot: bot, Translator: tr, Lang: "uk"}

	if err := s.SendTweet(context.Background(), 1001, 42, textTweet("Привіт"), nil); err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if bot.messageCalls != 1 {
		t.Fatalf("message calls = %d, want 1 (no translation reply)", bot.messageCalls)
	}
}

func TestSendTweetSkipsTranslationWhenNoText(t *testing.T) {
	bot := &translateFakeBot{}
	tr := &fakeTranslator{}
	s := Sender{Bot: bot, Translator: tr, Lang: "uk"}

	tw := textTweet("")
	tw.Media = &twitterxapi.Media{Photos: []twitterxapi.Photo{{URL: "https://img/1.jpg"}}}

	if err := s.SendTweet(context.Background(), 1001, 42, tw, nil); err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if tr.calls != 0 {
		t.Fatalf("translator calls = %d, want 0 (no text to translate)", tr.calls)
	}
}

func TestSendTweetTranslationErrorDoesNotFail(t *testing.T) {
	bot := &translateFakeBot{}
	tr := &fakeTranslator{err: errors.New("boom")}
	s := Sender{Bot: bot, Translator: tr, Lang: "uk"}

	if err := s.SendTweet(context.Background(), 1001, 42, textTweet("Hello world"), nil); err != nil {
		t.Fatalf("SendTweet() error = %v, want nil (translation failure must not break send)", err)
	}
	// Only the primary tweet message should be sent.
	if bot.messageCalls != 1 {
		t.Fatalf("message calls = %d, want 1 (primary only)", bot.messageCalls)
	}
}

func TestSendTweetNoTranslatorSendsOnlyTweet(t *testing.T) {
	bot := &translateFakeBot{}
	s := Sender{Bot: bot, Lang: "uk"}

	if err := s.SendTweet(context.Background(), 1001, 42, textTweet("Hello world"), nil); err != nil {
		t.Fatalf("SendTweet() error = %v", err)
	}
	if bot.messageCalls != 1 {
		t.Fatalf("message calls = %d, want 1 (no translator configured)", bot.messageCalls)
	}
}
