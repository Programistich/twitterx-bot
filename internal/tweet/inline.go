package tweet

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"twitterx-bot/internal/twitterxapi"
)

func BuildInlineResult(tweet *twitterxapi.Tweet, fallbackID string) (gotgbot.InlineQueryResult, bool) {
	if tweet == nil {
		return nil, false
	}

	resultID := strings.TrimSpace(tweet.ID)
	if resultID == "" {
		resultID = strings.TrimSpace(fallbackID)
	}
	if resultID == "" {
		resultID = "tweet"
	}

	title := BuildTitle(tweet)
	previewURL, previewKind := MediaPreview(tweet.Media)
	description := TruncateText(strings.TrimSpace(tweet.Text), MaxDescriptionLength)
	if description == "" {
		description = MediaHint(previewKind)
	}

	if tweet.Media != nil && len(tweet.Media.Videos) > 0 {
		video := tweet.Media.Videos[0]
		videoURL := strings.TrimSpace(video.URL)
		thumbURL := strings.TrimSpace(video.ThumbnailURL)
		if thumbURL == "" {
			thumbURL = previewURL
		}
		if videoURL != "" && thumbURL != "" {
			return gotgbot.InlineQueryResultVideo{
				Id:           resultID + ":video",
				Title:        title,
				VideoUrl:     videoURL,
				MimeType:     MimeTypeForVideo(video.Format),
				ThumbnailUrl: thumbURL,
				Caption:      Caption(tweet),
				Description:  description,
				VideoWidth:   int64(video.Width),
				VideoHeight:  int64(video.Height),
			}, true
		}
	}

	if tweet.Media != nil && len(tweet.Media.Photos) > 0 {
		photoURL, thumbURL, width, height := SelectPhoto(tweet.Media)
		if thumbURL == "" {
			thumbURL = previewURL
		}
		if photoURL != "" && thumbURL != "" {
			return gotgbot.InlineQueryResultPhoto{
				Id:           resultID + ":photo",
				PhotoUrl:     photoURL,
				ThumbnailUrl: thumbURL,
				PhotoWidth:   int64(width),
				PhotoHeight:  int64(height),
				Title:        title,
				Description:  description,
				Caption:      Caption(tweet),
			}, true
		}
	}

	message := MessageText(tweet)
	if message == "" {
		return nil, false
	}

	thumbURL := previewURL
	if thumbURL == "" {
		thumbURL = strings.TrimSpace(tweet.Author.AvatarURL)
	}

	return gotgbot.InlineQueryResultArticle{
		Id:    resultID + ":text",
		Title: title,
		InputMessageContent: gotgbot.InputTextMessageContent{
			MessageText: message,
		},
		Url:          strings.TrimSpace(tweet.URL),
		Description:  description,
		ThumbnailUrl: thumbURL,
	}, true
}
