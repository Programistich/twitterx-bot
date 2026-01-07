package twitterurl

import "regexp"

// TweetURLRegex matches a Twitter/X status URL and captures username and tweet ID.
// Note: intentionally preserves current behavior (including substring matches).
var TweetURLRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?(?:twitter\.com|x\.com)/([^/]+)/status/(\d+)`)

// ParseTweetURL extracts username and tweetID from the first matching URL in text.
func ParseTweetURL(text string) (username string, tweetID string, ok bool) {
	matches := TweetURLRegex.FindStringSubmatch(text)
	if matches == nil {
		return "", "", false
	}
	return matches[1], matches[2], true
}
