package tweet

import (
	"context"
	"errors"
	"fmt"

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
	ReplyMarkup       *gotgbot.InlineKeyboardMarkup
	RequesterUsername string
}

// Sender sends tweets to Telegram.
type Sender struct {
	Bot       BotAPI
	Formatter Formatter
	Telegraph ArticleCreator // Optional: for creating articles when text is too long
}

// SendResponse sends a single tweet reply to the chat message in ctx.
func (s Sender) SendResponse(ctx *ext.Context, tweet *twitterxapi.Tweet, opts *SendResponseOpts) error {
	if ctx == nil {
		return nil
	}
	return s.SendTweet(context.Background(), ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, tweet, opts)
}

// SendTweet sends a single tweet response to the given chat.
func (s Sender) SendTweet(_ context.Context, chatID, replyToMsgID int64, tweet *twitterxapi.Tweet, opts *SendResponseOpts) error {
	if tweet == nil {
		return nil
	}
	if s.Bot == nil {
		return errors.New("tweet sender: bot is nil")
	}

	var replyParams *gotgbot.ReplyParameters
	if replyToMsgID != 0 {
		replyParams = &gotgbot.ReplyParameters{
			MessageId:                replyToMsgID,
			AllowSendingWithoutReply: true,
		}
	}

	var replyMarkup *gotgbot.InlineKeyboardMarkup
	var requesterUsername string
	if opts != nil {
		replyMarkup = opts.ReplyMarkup
		requesterUsername = opts.RequesterUsername
	}

	_, err := s.sendTweetMessage(chatID, tweet, &sendTweetMessageOpts{
		ReplyParams:       replyParams,
		ReplyMarkup:       replyMarkup,
		RequesterUsername: requesterUsername,
	})
	return err
}

// sendTweetMessageOpts contains options for sendTweetMessage.
type sendTweetMessageOpts struct {
	ReplyParams       *gotgbot.ReplyParameters
	ReplyMarkup       *gotgbot.InlineKeyboardMarkup
	RequesterUsername string
}

// sendTweetMessage sends a tweet as a Telegram message and returns the sent message.
// This helper is used for both single tweet responses and chain threading.
func (s Sender) sendTweetMessage(chatID int64, tweet *twitterxapi.Tweet, opts *sendTweetMessageOpts) (*gotgbot.Message, error) {
	if s.Bot == nil {
		return nil, errors.New("tweet sender: bot is nil")
	}
	if opts == nil {
		opts = &sendTweetMessageOpts{}
	}

	f := s.Formatter.withDefaults()

	// Check if we need Telegraph for long text
	caption := s.prepareCaption(context.Background(), tweet, opts.RequesterUsername, f)

	// Priority 1: Video
	if tweet.Media != nil && len(tweet.Media.Videos) > 0 {
		video := tweet.Media.Videos[0]
		if video.URL != "" {
			videoOpts := &gotgbot.SendVideoOpts{
				Caption:         caption,
				ParseMode:       "HTML",
				Width:           int64(video.Width),
				Height:          int64(video.Height),
				ReplyParameters: opts.ReplyParams,
			}
			if opts.ReplyMarkup != nil {
				videoOpts.ReplyMarkup = opts.ReplyMarkup
			}
			return s.Bot.SendVideo(chatID, gotgbot.InputFileByURL(video.URL), videoOpts)
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
				ReplyParameters: opts.ReplyParams,
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
			photoOpts := &gotgbot.SendPhotoOpts{
				Caption:         caption,
				ParseMode:       "HTML",
				ReplyParameters: opts.ReplyParams,
			}
			if opts.ReplyMarkup != nil {
				photoOpts.ReplyMarkup = opts.ReplyMarkup
			}
			return s.Bot.SendPhoto(chatID, gotgbot.InputFileByURL(photo.URL), photoOpts)
		}
	}

	// Priority 4: Text only
	message := f.HTMLMessageTextWithRequester(tweet, opts.RequesterUsername)
	if message != "" {
		msgOpts := &gotgbot.SendMessageOpts{
			ParseMode:       "HTML",
			ReplyParameters: opts.ReplyParams,
		}
		if opts.ReplyMarkup != nil {
			msgOpts.ReplyMarkup = opts.ReplyMarkup
		}
		return s.Bot.SendMessage(chatID, message, msgOpts)
	}

	return nil, nil
}

// prepareCaption creates the caption text for a tweet.
// If Telegraph is configured and the text exceeds limits, it creates a Telegraph article.
func (s Sender) prepareCaption(ctx context.Context, tweet *twitterxapi.Tweet, requesterUsername string, f Formatter) string {
	baseCaption := f.HTMLContentWithRequester(tweet, requesterUsername)

	// If no Telegraph configured or text fits, use regular caption
	if s.Telegraph == nil || len(baseCaption) <= MaxCaptionLength {
		return f.HTMLCaptionWithRequester(tweet, requesterUsername)
	}

	// Text is too long, try Telegraph
	title := f.Title(tweet)
	fullText := f.Content(tweet)

	articleURL, err := s.Telegraph.CreateArticle(ctx, fullText, title)
	if err != nil {
		// Fallback: truncate text and add link to original tweet
		return s.fallbackCaption(tweet, requesterUsername, f)
	}

	// Success: truncate and add Telegraph link
	truncatedCaption := TruncateHTML(baseCaption, MaxCaptionLength-60) // Reserve space for Telegraph link
	return fmt.Sprintf("%s\n\n📖 %s", truncatedCaption, articleURL)
}

// fallbackCaption creates a truncated caption with link to original tweet.
// Used when Telegraph is unavailable or fails.
func (s Sender) fallbackCaption(tweet *twitterxapi.Tweet, requesterUsername string, f Formatter) string {
	if tweet == nil {
		return ""
	}

	caption := f.HTMLCaptionWithRequester(tweet, requesterUsername)

	// If text fits, return as-is
	if len(caption) <= MaxCaptionLength {
		return caption
	}

	// Truncate and add link to original
	if tweet.URL != "" {
		truncated := TruncateHTML(caption, MaxCaptionLength-50)
		return fmt.Sprintf("%s\n\n📎 %s", truncated, tweet.URL)
	}

	return TruncateHTML(caption, MaxCaptionLength)
}

// SendChainResponseOpts contains options for SendChainResponse.
type SendChainResponseOpts struct {
	RequesterUsername string
}

// SendChainResponse sends a chain of tweets as separate messages, each replying to the previous.
// If replyToMsgID is provided (non-zero), the first message will reply to that message.
// The last message in the chain will have a "Delete original" button and requester username.
func (s Sender) SendChainResponse(chatID int64, items []chain.ChainItem, replyToMsgID int64, opts *SendChainResponseOpts) error {
	if len(items) == 0 {
		return nil
	}
	if s.Bot == nil {
		return errors.New("tweet sender: bot is nil")
	}
	if opts == nil {
		opts = &SendChainResponseOpts{}
	}

	prevMsgID := replyToMsgID

	for i, item := range items {
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

		msgOpts := &sendTweetMessageOpts{
			ReplyParams: replyParams,
		}

		// Add "Delete original" button and requester username only to the last message
		if i == len(items)-1 {
			if replyToMsgID != 0 {
				msgOpts.ReplyMarkup = BuildKeyboard(replyToMsgID, nil)
			}
			msgOpts.RequesterUsername = opts.RequesterUsername
		}

		msg, err := s.sendTweetMessage(chatID, item.Tweet, msgOpts)
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

func SendChainResponse(b *gotgbot.Bot, chatID int64, items []chain.ChainItem, replyToMsgID int64, opts *SendChainResponseOpts) error {
	sender := Sender{Bot: b}
	return sender.SendChainResponse(chatID, items, replyToMsgID, opts)
}
