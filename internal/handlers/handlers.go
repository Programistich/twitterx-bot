package handlers

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"twitterx-bot/internal/chain"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/tweet"
	"twitterx-bot/internal/twitterxapi"
)

var twitterURLRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?(?:twitter\.com|x\.com)/([^/]+)/status/(\d+)`)

type Handlers struct {
	log *logger.Logger
	api *twitterxapi.Client
}

const (
	inlineQueryTimeout = 10 * time.Second
	chainTimeout       = 30 * time.Second
)

func Register(d *ext.Dispatcher, log *logger.Logger, api *twitterxapi.Client) {
	if api == nil {
		api = twitterxapi.NewClient("")
	}
	h := &Handlers{log: log, api: api}

	d.AddHandler(handlers.NewCommand("start", start))
	d.AddHandler(handlers.NewInlineQuery(func(iq *gotgbot.InlineQuery) bool {
		return true
	}, h.inlineQuery))
	d.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		if msg.Text == "" {
			return false
		}
		return twitterURLRegex.MatchString(msg.Text)
	}, h.messageHandler))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.ChainCallbackPrefix)
	}, h.chainCallback))
	d.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, tweet.DeleteCallbackPrefix)
	}, h.deleteCallback))
}

func start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Hi! Send me any text and I will echo it back.", &gotgbot.SendMessageOpts{})
	return err
}

func (h *Handlers) inlineQuery(b *gotgbot.Bot, ctx *ext.Context) error {
	query := strings.TrimSpace(ctx.InlineQuery.Query)
	h.log.Debug("received inline query: %s", query)

	matches := twitterURLRegex.FindStringSubmatch(query)
	if matches == nil {
		h.log.Debug("no twitter URL found in query")
		_, err := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		return err
	}

	username := matches[1]
	tweetID := matches[2]

	h.log.Info("twitter URL: %s", query)
	h.log.Info("twitter URL parsed - username: %s, tweet_id: %s", username, tweetID)
	h.log.Debug("full match details - query: %s, username: %s, tweet_id: %s", query, username, tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), inlineQueryTimeout)
	defer cancel()

	tw, err := h.api.GetTweet(reqCtx, username, tweetID)
	if err != nil {
		h.log.Error("failed to fetch tweet %s for %s: %v", tweetID, username, err)
		_, answerErr := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		if answerErr != nil {
			return answerErr
		}
		return nil
	}

	result, ok := tweet.BuildInlineResult(tw, tweetID)
	if !ok {
		h.log.Error("no suitable inline result for tweet %s", tweetID)
		_, answerErr := ctx.InlineQuery.Answer(b, nil, &gotgbot.AnswerInlineQueryOpts{
			CacheTime:  0,
			IsPersonal: true,
		})
		if answerErr != nil {
			return answerErr
		}
		return nil
	}

	results := []gotgbot.InlineQueryResult{result}

	_, err = ctx.InlineQuery.Answer(b, results, &gotgbot.AnswerInlineQueryOpts{
		CacheTime:  0,
		IsPersonal: true,
	})
	return err
}

func (h *Handlers) messageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	text := strings.TrimSpace(ctx.EffectiveMessage.Text)
	h.log.Debug("received message with twitter URL: %s", text)

	matches := twitterURLRegex.FindStringSubmatch(text)
	if matches == nil {
		return nil
	}

	username := matches[1]
	tweetID := matches[2]

	h.log.Info("twitter URL from message - username: %s, tweet_id: %s", username, tweetID)

	reqCtx, cancel := context.WithTimeout(context.Background(), inlineQueryTimeout)
	defer cancel()

	_, err := b.SendChatAction(ctx.EffectiveChat.Id, gotgbot.ChatActionTyping, &gotgbot.SendChatActionOpts{})
	if err != nil {
		h.log.Debug("failed to send typing action: %v", err)
	}

	tw, err := h.api.GetTweet(reqCtx, username, tweetID)
	if err != nil {
		h.log.Error("failed to fetch tweet %s for %s: %v", tweetID, username, err)
		return nil
	}

	// Build keyboard with optional chain button and always delete button
	var keyboardOpts *tweet.KeyboardOpts
	if tw.ReplyingToStatus != nil {
		keyboardOpts = &tweet.KeyboardOpts{
			ShowChainButton: true,
			ChainUsername:   username,
			ChainTweetID:    tweetID,
		}
	}

	opts := &tweet.SendResponseOpts{
		ReplyMarkup: tweet.BuildKeyboard(ctx.EffectiveMessage.MessageId, keyboardOpts),
	}

	return tweet.SendResponse(b, ctx, tw, opts)
}

func (h *Handlers) chainCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery

	username, tweetID, replyToMsgID, ok := tweet.DecodeChainCallback(cb.Data)
	if !ok {
		h.log.Error("failed to decode chain callback: %s", cb.Data)
		_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Invalid callback data",
		})
		return err
	}

	h.log.Info("chain callback - username: %s, tweet_id: %s, reply_to_msg_id: %d", username, tweetID, replyToMsgID)

	// Answer callback immediately with loading message
	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Fetching full chain...",
	})
	if err != nil {
		h.log.Debug("failed to answer callback: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), chainTimeout)
	defer cancel()

	// Fetch the tweet
	tw, err := h.api.GetTweet(reqCtx, username, tweetID)
	if err != nil {
		h.log.Error("failed to fetch tweet %s for %s: %v", tweetID, username, err)
		return nil
	}

	// Build the chain
	items, err := chain.BuildChain(reqCtx, h.api, tw)
	if err != nil {
		h.log.Error("failed to build chain for tweet %s: %v", tweetID, err)
		return nil
	}

	_, delErr := cb.Message.Delete(b, nil)
	if delErr != nil {
		h.log.Debug("failed to delete original message: %v", delErr)
	}

	// Send the chain, replying to the user's message
	chatID := ctx.EffectiveChat.Id
	return tweet.SendChainResponse(b, chatID, items, replyToMsgID)
}

func (h *Handlers) deleteCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery

	msgID, ok := tweet.DecodeDeleteCallback(cb.Data)
	if !ok {
		h.log.Error("failed to decode delete callback: %s", cb.Data)
		_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Invalid callback data",
		})
		return err
	}

	h.log.Info("delete callback - msg_id: %d", msgID)

	chatID := ctx.EffectiveChat.Id

	// Try to delete the original user message (may fail if bot is not admin)
	_, err := b.DeleteMessage(chatID, msgID, nil)
	if err != nil {
		h.log.Debug("failed to delete original message %d: %v", msgID, err)
		_, answerErr := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Cannot delete message",
		})
		return answerErr
	}

	// Check if there's a chain button - if so, edit to keep only that button
	// Otherwise, remove the keyboard entirely
	msg, ok := cb.Message.(*gotgbot.Message)
	if ok && msg.ReplyMarkup != nil {
		chainCallbackData := tweet.FindChainButton(msg.ReplyMarkup)
		if chainCallbackData != "" {
			// Has chain button - edit keyboard to keep only chain button
			_, _, editErr := msg.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
				ReplyMarkup: *tweet.BuildChainOnlyKeyboard(chainCallbackData),
			})
			if editErr != nil {
				h.log.Debug("failed to edit reply markup: %v", editErr)
			}
		} else {
			// No chain button - remove keyboard entirely
			_, _, editErr := msg.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
			})
			if editErr != nil {
				h.log.Debug("failed to remove reply markup: %v", editErr)
			}
		}
	}

	_, err = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Deleted",
	})
	return err
}
