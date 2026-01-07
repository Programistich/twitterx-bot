package tweet

import (
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/chain"
	"twitterx-bot/internal/twitterxapi"
)

// BotAPI abstracts the Telegram bot API for testing.
type BotAPI interface {
	SendVideo(chatID int64, video gotgbot.InputFileOrString, opts *gotgbot.SendVideoOpts) (*gotgbot.Message, error)
	SendPhoto(chatID int64, photo gotgbot.InputFileOrString, opts *gotgbot.SendPhotoOpts) (*gotgbot.Message, error)
	SendMediaGroup(chatID int64, media []gotgbot.InputMedia, opts *gotgbot.SendMediaGroupOpts) ([]gotgbot.Message, error)
	SendMessage(chatID int64, text string, opts *gotgbot.SendMessageOpts) (*gotgbot.Message, error)
}

// SendResponseOpts contains optional parameters for SendResponse.
type SendResponseOpts struct {
	ReplyMarkup *gotgbot.InlineKeyboardMarkup
}

// Sender sends tweets to Telegram.
type Sender struct {
	Bot       BotAPI
	Formatter Formatter
}

// SendResponse sends a single tweet reply to the chat message in ctx.
func (s Sender) SendResponse(ctx *ext.Context, tweet *twitterxapi.Tweet, opts *SendResponseOpts) error {
	if tweet == nil {
		return nil
	}
	if s.Bot == nil {
		return errors.New("tweet sender: bot is nil")
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

	_, err := s.sendTweetMessage(chatID, tweet, replyParams, replyMarkup)
	return err
}

// sendTweetMessage sends a tweet as a Telegram message and returns the sent message.
// This helper is used for both single tweet responses and chain threading.
func (s Sender) sendTweetMessage(chatID int64, tweet *twitterxapi.Tweet, replyParams *gotgbot.ReplyParameters, replyMarkup *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
	if s.Bot == nil {
		return nil, errors.New("tweet sender: bot is nil")
	}

	f := s.Formatter.withDefaults()
	caption := f.HTMLCaption(tweet)

	// Priority 1: Video
	if tweet.Media != nil && len(tweet.Media.Videos) > 0 {
		video := tweet.Media.Videos[0]
		if video.URL != "" {
			opts := &gotgbot.SendVideoOpts{
				Caption:         caption,
				ParseMode:       "HTML",
				Width:           int64(video.Width),
				Height:          int64(video.Height),
				ReplyParameters: replyParams,
			}
			if replyMarkup != nil {
				opts.ReplyMarkup = replyMarkup
			}
			return s.Bot.SendVideo(chatID, gotgbot.InputFileByURL(video.URL), opts)
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
				inputPhoto.ParseMode = "HTML"
			}
			mediaGroup = append(mediaGroup, inputPhoto)
		}

		if len(mediaGroup) > 0 {
			// SendMediaGroup returns []Message, use first for threading
			msgs, err := s.Bot.SendMediaGroup(chatID, mediaGroup, &gotgbot.SendMediaGroupOpts{
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
				ParseMode:       "HTML",
				ReplyParameters: replyParams,
			}
			if replyMarkup != nil {
				opts.ReplyMarkup = replyMarkup
			}
			return s.Bot.SendPhoto(chatID, gotgbot.InputFileByURL(photo.URL), opts)
		}
	}

	// Priority 4: Text only
	message := f.HTMLMessageText(tweet)
	if message != "" {
		opts := &gotgbot.SendMessageOpts{
			ParseMode:       "HTML",
			ReplyParameters: replyParams,
		}
		if replyMarkup != nil {
			opts.ReplyMarkup = replyMarkup
		}
		return s.Bot.SendMessage(chatID, message, opts)
	}

	return nil, nil
}

// SendChainResponse sends a chain of tweets as separate messages, each replying to the previous.
// If replyToMsgID is provided (non-zero), the first message will reply to that message.
func (s Sender) SendChainResponse(chatID int64, items []chain.ChainItem, replyToMsgID int64) error {
	if len(items) == 0 {
		return nil
	}
	if s.Bot == nil {
		return errors.New("tweet sender: bot is nil")
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

		msg, err := s.sendTweetMessage(chatID, item.Tweet, replyParams, nil)
		if err != nil {
			return err
		}

		if msg != nil {
			prevMsgID = msg.MessageId
		}
	}

	return nil
}

func SendResponse(b *gotgbot.Bot, ctx *ext.Context, tweet *twitterxapi.Tweet, opts *SendResponseOpts) error {
	sender := Sender{Bot: b}
	return sender.SendResponse(ctx, tweet, opts)
}

func SendChainResponse(b *gotgbot.Bot, chatID int64, items []chain.ChainItem, replyToMsgID int64) error {
	sender := Sender{Bot: b}
	return sender.SendChainResponse(chatID, items, replyToMsgID)
}
