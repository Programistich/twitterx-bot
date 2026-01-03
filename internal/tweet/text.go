package tweet

import (
	"fmt"
	"strings"

	"twitterx-bot/internal/twitterxapi"
)

const (
	MaxCaptionLength     = 1024
	MaxMessageLength     = 4096
	MaxDescriptionLength = 140
)

func BuildTitle(tweet *twitterxapi.Tweet) string {
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

func Caption(tweet *twitterxapi.Tweet) string {
	return TruncateText(Content(tweet), MaxCaptionLength)
}

func MessageText(tweet *twitterxapi.Tweet) string {
	return TruncateText(Content(tweet), MaxMessageLength)
}

func Content(tweet *twitterxapi.Tweet) string {
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
