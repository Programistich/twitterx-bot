package tweet

import (
	"fmt"
	"strings"

	"twitterx-bot/internal/twitterxapi"
)

const (
	// Telegram limits: https://core.telegram.org/bots/api#sending-messages
	MaxCaptionLength     = 1024
	MaxMessageLength     = 4096
	MaxDescriptionLength = 140
)

// Formatter provides tweet-to-text helpers with configurable limits.
type Formatter struct {
	MaxCaptionLength     int
	MaxMessageLength     int
	MaxDescriptionLength int
}

// DefaultFormatter returns formatter defaults aligned with Telegram limits.
func DefaultFormatter() Formatter {
	return Formatter{
		MaxCaptionLength:     MaxCaptionLength,
		MaxMessageLength:     MaxMessageLength,
		MaxDescriptionLength: MaxDescriptionLength,
	}
}

func (f Formatter) withDefaults() Formatter {
	if f.MaxCaptionLength <= 0 {
		f.MaxCaptionLength = MaxCaptionLength
	}
	if f.MaxMessageLength <= 0 {
		f.MaxMessageLength = MaxMessageLength
	}
	if f.MaxDescriptionLength <= 0 {
		f.MaxDescriptionLength = MaxDescriptionLength
	}
	return f
}

func (f Formatter) Title(tweet *twitterxapi.Tweet) string {
	if tweet == nil {
		return "Tweet"
	}
	if tweet.Author.ScreenName != "" {
		screenName := tweet.Author.ScreenName
		if !strings.HasPrefix(screenName, "@") {
			screenName = "@" + screenName
		}
		return fmt.Sprintf("Tweet by %s", screenName)
	}
	if tweet.Author.Name != "" {
		return fmt.Sprintf("Tweet by %s", tweet.Author.Name)
	}
	return "Tweet"
}

func (f Formatter) Caption(tweet *twitterxapi.Tweet) string {
	f = f.withDefaults()
	return TruncateText(f.Content(tweet), f.MaxCaptionLength)
}

func (f Formatter) MessageText(tweet *twitterxapi.Tweet) string {
	f = f.withDefaults()
	return TruncateText(f.Content(tweet), f.MaxMessageLength)
}

func (f Formatter) Description(tweet *twitterxapi.Tweet) string {
	f = f.withDefaults()
	if tweet == nil {
		return ""
	}
	return TruncateText(strings.TrimSpace(tweet.Text), f.MaxDescriptionLength)
}

func (f Formatter) Content(tweet *twitterxapi.Tweet) string {
	if tweet == nil {
		return ""
	}

	text := strings.TrimSpace(tweet.Text)
	url := strings.TrimSpace(tweet.URL)
	if text == "" {
		return url
	}
	if url == "" {
		return text
	}
	return text + "\n\n" + url
}

func BuildTitle(tweet *twitterxapi.Tweet) string {
	return Formatter{}.Title(tweet)
}

func Caption(tweet *twitterxapi.Tweet) string {
	return Formatter{}.Caption(tweet)
}

func MessageText(tweet *twitterxapi.Tweet) string {
	return Formatter{}.MessageText(tweet)
}

func Content(tweet *twitterxapi.Tweet) string {
	return Formatter{}.Content(tweet)
}

func TruncateText(input string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(input) <= max {
		return input
	}
	runes := []rune(input)
	if len(runes) <= max {
		return input
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}
