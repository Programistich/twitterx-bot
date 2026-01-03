package tweet

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/twitterxapi"
)

func SendResponse(b *gotgbot.Bot, ctx *ext.Context, tweet *twitterxapi.Tweet) error {
	if tweet == nil {
		return nil
	}

	chatID := ctx.EffectiveChat.Id
	caption := Caption(tweet)
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
		if len(photos) > MaxMediaGroupSize {
			photos = photos[:MaxMediaGroupSize]
		}

		mediaGroup := make([]gotgbot.InputMedia, 0, len(photos))
		for i, photo := range photos {
			if photo.URL == "" {
				continue
			}
			inputPhoto := gotgbot.InputMediaPhoto{
				Media: gotgbot.InputFileByURL(photo.URL),
			}
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
	message := MessageText(tweet)
	if message != "" {
		_, err := b.SendMessage(chatID, message, &gotgbot.SendMessageOpts{
			ReplyParameters: replyParams,
		})
		return err
	}

	return nil
}
