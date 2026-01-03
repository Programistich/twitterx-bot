package handlers

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/twitterxapi"
)

const (
	maxCaptionLength     = 1024
	maxMessageLength     = 4096
	maxDescriptionLength = 140
	maxMediaGroupSize    = 10
)

func buildInlineResult(tweet *twitterxapi.Tweet, fallbackID string) (gotgbot.InlineQueryResult, bool) {
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

	title := buildTitle(tweet)
	previewURL, previewKind := mediaPreview(tweet.Media)
	description := truncateText(strings.TrimSpace(tweet.Text), maxDescriptionLength)
	if description == "" {
		description = mediaHint(previewKind)
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
				MimeType:     mimeTypeForVideo(video.Format),
				ThumbnailUrl: thumbURL,
				Caption:      tweetCaption(tweet),
				Description:  description,
				VideoWidth:   int64(video.Width),
				VideoHeight:  int64(video.Height),
			}, true
		}
	}

	if tweet.Media != nil && len(tweet.Media.Photos) > 0 {
		photoURL, thumbURL, width, height := selectPhoto(tweet.Media)
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
				Caption:      tweetCaption(tweet),
			}, true
		}
	}

	message := tweetMessageText(tweet)
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

func buildTitle(tweet *twitterxapi.Tweet) string {
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

func tweetCaption(tweet *twitterxapi.Tweet) string {
	return truncateText(tweetContent(tweet), maxCaptionLength)
}

func tweetMessageText(tweet *twitterxapi.Tweet) string {
	return truncateText(tweetContent(tweet), maxMessageLength)
}

func tweetContent(tweet *twitterxapi.Tweet) string {
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

func selectPhoto(media *twitterxapi.Media) (string, string, int, int) {
	if media == nil || len(media.Photos) == 0 {
		return "", "", 0, 0
	}

	if len(media.Photos) > 1 {
		if mosaicURL := pickMosaicURL(media.Mosaic); mosaicURL != "" {
			width, height := mosaicDimensions(media.Mosaic)
			return mosaicURL, mosaicURL, width, height
		}
	}

	photo := media.Photos[0]
	return photo.URL, photo.URL, photo.Width, photo.Height
}

func mediaPreview(media *twitterxapi.Media) (string, string) {
	if media == nil {
		return "", ""
	}

	if len(media.Videos) > 0 {
		if url := strings.TrimSpace(media.Videos[0].ThumbnailURL); url != "" {
			return url, "video"
		}
	}

	if len(media.Photos) > 0 {
		if len(media.Photos) > 1 {
			if mosaicURL := pickMosaicURL(media.Mosaic); mosaicURL != "" {
				return mosaicURL, "mosaic"
			}
		}
		if url := strings.TrimSpace(media.Photos[0].URL); url != "" {
			return url, "photo"
		}
	}

	return "", ""
}

func mediaHint(kind string) string {
	switch kind {
	case "video":
		return "Video"
	case "mosaic":
		return "Mosaic"
	case "photo":
		return "Photo"
	default:
		return ""
	}
}

func pickMosaicURL(mosaic *twitterxapi.Mosaic) string {
	if mosaic == nil || len(mosaic.Formats) == 0 {
		return ""
	}
	if url := strings.TrimSpace(mosaic.Formats["jpeg"]); url != "" {
		return url
	}
	if url := strings.TrimSpace(mosaic.Formats["jpg"]); url != "" {
		return url
	}
	return ""
}

func mosaicDimensions(mosaic *twitterxapi.Mosaic) (int, int) {
	if mosaic == nil {
		return 0, 0
	}
	width := 0
	height := 0
	if mosaic.Width != nil {
		width = *mosaic.Width
	}
	if mosaic.Height != nil {
		height = *mosaic.Height
	}
	return width, height
}

func mimeTypeForVideo(format string) string {
	format = strings.TrimSpace(format)
	if format == "" {
		return "video/mp4"
	}
	if strings.Contains(format, "/") {
		return format
	}
	return "video/" + format
}

func truncateText(input string, max int) string {
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

func sendTweetResponse(b *gotgbot.Bot, ctx *ext.Context, tweet *twitterxapi.Tweet) error {
	if tweet == nil {
		return nil
	}

	chatID := ctx.EffectiveChat.Id
	caption := tweetCaption(tweet)
	replyParams := &gotgbot.ReplyParameters{
		MessageId:                ctx.EffectiveMessage.MessageId,
		AllowSendingWithoutReply: true,
	}

	// Priority 1: Video
	if tweet.Media != nil && len(tweet.Media.Videos) > 0 {
		video := tweet.Media.Videos[0]
		if video.URL != "" {
			_, err := b.SendVideo(chatID, gotgbot.InputFileByURL(video.URL), &gotgbot.SendVideoOpts{
				Caption:         caption,
				Width:           int64(video.Width),
				Height:          int64(video.Height),
				ReplyParameters: replyParams,
			})
			return err
		}
	}

	// Priority 2: Multiple photos as media group
	if tweet.Media != nil && len(tweet.Media.Photos) > 1 {
		photos := tweet.Media.Photos
		if len(photos) > maxMediaGroupSize {
			photos = photos[:maxMediaGroupSize]
		}

		mediaGroup := make([]gotgbot.InputMedia, 0, len(photos))
		for i, photo := range photos {
			if photo.URL == "" {
				continue
			}
			inputPhoto := gotgbot.InputMediaPhoto{
				Media: gotgbot.InputFileByURL(photo.URL),
			}
			// Only first photo gets caption
			if i == 0 {
				inputPhoto.Caption = caption
			}
			mediaGroup = append(mediaGroup, inputPhoto)
		}

		if len(mediaGroup) > 0 {
			_, err := b.SendMediaGroup(chatID, mediaGroup, &gotgbot.SendMediaGroupOpts{
				ReplyParameters: replyParams,
			})
			return err
		}
	}

	// Priority 3: Single photo
	if tweet.Media != nil && len(tweet.Media.Photos) == 1 {
		photo := tweet.Media.Photos[0]
		if photo.URL != "" {
			_, err := b.SendPhoto(chatID, gotgbot.InputFileByURL(photo.URL), &gotgbot.SendPhotoOpts{
				Caption:         caption,
				ReplyParameters: replyParams,
			})
			return err
		}
	}

	// Priority 4: Text only
	message := tweetMessageText(tweet)
	if message != "" {
		_, err := b.SendMessage(chatID, message, &gotgbot.SendMessageOpts{
			ReplyParameters: replyParams,
		})
		return err
	}

	return nil
}
