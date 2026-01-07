package tweet

import (
	"fmt"
	"html"
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

// authorProfileURL returns the Twitter/X profile URL for the author.
func authorProfileURL(screenName string) string {
	if screenName == "" {
		return ""
	}
	name := strings.TrimPrefix(screenName, "@")
	return "https://x.com/" + name
}

// HTMLContent returns HTML-formatted tweet content with linked header.
// Format: <a href="tweet_url">Tweet</a> from <a href="profile_url">Author Name</a>\n\ntext
func (f Formatter) HTMLContent(tweet *twitterxapi.Tweet) string {
	if tweet == nil {
		return ""
	}

	var sb strings.Builder

	// Build header: Tweet from Author
	tweetURL := strings.TrimSpace(tweet.URL)
	authorName := strings.TrimSpace(tweet.Author.Name)
	screenName := strings.TrimSpace(tweet.Author.ScreenName)

	// Use Name for display, fall back to ScreenName if Name is empty
	displayName := authorName
	if displayName == "" {
		displayName = screenName
	}

	if tweetURL != "" {
		sb.WriteString(fmt.Sprintf(`<a href="%s">Tweet</a>`, html.EscapeString(tweetURL)))
	} else {
		sb.WriteString("Tweet")
	}

	if displayName != "" {
		profileURL := authorProfileURL(screenName)
		if profileURL != "" {
			sb.WriteString(fmt.Sprintf(` from <a href="%s">%s</a>`, html.EscapeString(profileURL), html.EscapeString(displayName)))
		} else {
			sb.WriteString(fmt.Sprintf(" from %s", html.EscapeString(displayName)))
		}
	}

	// Add tweet text
	text := strings.TrimSpace(tweet.Text)
	if text != "" {
		sb.WriteString("\n\n")
		sb.WriteString(html.EscapeString(text))
	}

	return sb.String()
}

// HTMLCaption returns HTML-formatted tweet for media captions (max 1024 chars).
func (f Formatter) HTMLCaption(tweet *twitterxapi.Tweet) string {
	f = f.withDefaults()
	return TruncateHTML(f.HTMLContent(tweet), f.MaxCaptionLength)
}

// HTMLMessageText returns HTML-formatted tweet for text messages (max 4096 chars).
func (f Formatter) HTMLMessageText(tweet *twitterxapi.Tweet) string {
	f = f.withDefaults()
	return TruncateHTML(f.HTMLContent(tweet), f.MaxMessageLength)
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

func HTMLCaption(tweet *twitterxapi.Tweet) string {
	return Formatter{}.HTMLCaption(tweet)
}

func HTMLMessageText(tweet *twitterxapi.Tweet) string {
	return Formatter{}.HTMLMessageText(tweet)
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

// TruncateHTML truncates HTML content to max characters while preserving valid HTML structure.
// It counts visible characters (excluding HTML tags) and ensures all opened tags are closed.
func TruncateHTML(input string, max int) string {
	if max <= 0 {
		return ""
	}

	runes := []rune(input)
	if len(runes) <= max {
		return input
	}

	var result strings.Builder
	var openTags []string
	visibleCount := 0
	inTag := false
	tagStart := 0
	truncated := false

	for i, r := range runes {
		if r == '<' {
			inTag = true
			tagStart = i
			continue
		}

		if r == '>' && inTag {
			inTag = false
			tagContent := string(runes[tagStart+1 : i])

			// Check if it's a closing tag
			if strings.HasPrefix(tagContent, "/") {
				tagName := strings.TrimPrefix(tagContent, "/")
				tagName = strings.Fields(tagName)[0] // Get just the tag name
				// Remove from open tags stack
				for j := len(openTags) - 1; j >= 0; j-- {
					if openTags[j] == tagName {
						openTags = append(openTags[:j], openTags[j+1:]...)
						break
					}
				}
				result.WriteString(string(runes[tagStart : i+1]))
			} else if !strings.HasSuffix(tagContent, "/") {
				// Opening tag (not self-closing)
				tagName := strings.Fields(tagContent)[0]
				openTags = append(openTags, tagName)
				result.WriteString(string(runes[tagStart : i+1]))
			} else {
				// Self-closing tag
				result.WriteString(string(runes[tagStart : i+1]))
			}
			continue
		}

		if inTag {
			continue
		}

		// Visible character
		if visibleCount >= max-3 && !truncated {
			truncated = true
			result.WriteString("...")
			break
		}

		result.WriteRune(r)
		visibleCount++
	}

	// Close any remaining open tags in reverse order
	for i := len(openTags) - 1; i >= 0; i-- {
		result.WriteString("</" + openTags[i] + ">")
	}

	return result.String()
}
