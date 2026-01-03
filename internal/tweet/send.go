package tweet

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/chain"
	"twitterx-bot/internal/twitterxapi"
)

// SendResponseOpts contains optional parameters for SendResponse.
type SendResponseOpts struct {
	ReplyMarkup *gotgbot.InlineKeyboardMarkup
}

func SendResponse(b *gotgbot.Bot, ctx *ext.Context, tweet *twitterxapi.Tweet, opts *SendResponseOpts) error {
	if tweet == nil {
		return nil
	}

	chatID := ctx.EffectiveChat.Id
	replyParams := &gotgbot.ReplyParameters{
		MessageId:                ctx.EffectiveMessage.MessageId,
		AllowSendingWithoutReply: true,
	}

	var replyMarkup *gotgbot.InlineKeyboardMarkup
	if opts != nil {
		replyMarkup = opts.ReplyMarkup
	}

	_, err := sendTweetMessage(b, chatID, tweet, replyParams, replyMarkup)
	return err
}

// sendTweetMessage sends a tweet as a Telegram message and returns the sent message.
// This helper is used for both single tweet responses and chain threading.
func sendTweetMessage(b *gotgbot.Bot, chatID int64, tweet *twitterxapi.Tweet, replyParams *gotgbot.ReplyParameters, replyMarkup *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
	caption := Caption(tweet)

	// Priority 1: Video
	if tweet.Media != nil && len(tweet.Media.Videos) > 0 {
		video := tweet.Media.Videos[0]
		if video.URL != "" {
			opts := &gotgbot.SendVideoOpts{
				Caption:         caption,
				Width:           int64(video.Width),
				Height:          int64(video.Height),
				ReplyParameters: replyParams,
			}
			if replyMarkup != nil {
				opts.ReplyMarkup = replyMarkup
			}
			return b.SendVideo(chatID, gotgbot.InputFileByURL(video.URL), opts)
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
			// SendMediaGroup returns []Message, use first for threading
			msgs, err := b.SendMediaGroup(chatID, mediaGroup, &gotgbot.SendMediaGroupOpts{
				ReplyParameters: replyParams,
			})
			if err != nil {
				return nil, err
			}
			if len(msgs) > 0 {
				return &msgs[0], nil
			}
			return nil, nil
		}
	}

	// Priority 3: Single photo
	if tweet.Media != nil && len(tweet.Media.Photos) == 1 {
		photo := tweet.Media.Photos[0]
		if photo.URL != "" {
			opts := &gotgbot.SendPhotoOpts{
				Caption:         caption,
				ReplyParameters: replyParams,
			}
			if replyMarkup != nil {
				opts.ReplyMarkup = replyMarkup
			}
			return b.SendPhoto(chatID, gotgbot.InputFileByURL(photo.URL), opts)
		}
	}

	// Priority 4: Text only
	message := MessageText(tweet)
	if message != "" {
		opts := &gotgbot.SendMessageOpts{
			ReplyParameters: replyParams,
		}
		if replyMarkup != nil {
			opts.ReplyMarkup = replyMarkup
		}
		return b.SendMessage(chatID, message, opts)
	}

	return nil, nil
}

// SendChainResponse sends a chain of tweets as separate messages, each replying to the previous.
// If replyToMsgID is provided (non-zero), the first message will reply to that message.
func SendChainResponse(b *gotgbot.Bot, chatID int64, items []chain.ChainItem, replyToMsgID int64) error {
	if len(items) == 0 {
		return nil
	}

	prevMsgID := replyToMsgID

	for _, item := range items {
		if item.Tweet == nil {
			continue
		}

		var replyParams *gotgbot.ReplyParameters
		if prevMsgID != 0 {
			replyParams = &gotgbot.ReplyParameters{
				MessageId:                prevMsgID,
				AllowSendingWithoutReply: true,
			}
		}

		msg, err := sendTweetMessage(b, chatID, item.Tweet, replyParams, nil)
		if err != nil {
			return err
		}

		if msg != nil {
			prevMsgID = msg.MessageId
		}
	}

	return nil
}
