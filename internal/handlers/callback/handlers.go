package callback

import (
	"context"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/handlers/shared"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/telegram/tweet"
	"twitterx-bot/internal/usecase/tweetsvc/sendchain"
)

// TweetFetcher fetches tweets by username and tweet ID.
type TweetFetcher interface {
	sendchain.TweetFetcher
}

// Handlers groups the callback-related dependencies.
type Handlers struct {
	log          *logger.Logger
	fetcher      TweetFetcher
	chainTimeout time.Duration
}

// New creates callback handlers with the configured logger, tweet fetcher, and chain timeout.
func New(log *logger.Logger, fetcher TweetFetcher, chainTimeout time.Duration) *Handlers {
	return &Handlers{log: log, fetcher: fetcher, chainTimeout: chainTimeout}
}

// Chain processes callback queries that request a tweet chain.
func (h *Handlers) Chain(b *gotgbot.Bot, ctx *ext.Context) error {
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

	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Fetching full chain...",
	})
	if err != nil {
		h.log.Debug("failed to answer callback: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), h.chainTimeout)
	defer cancel()

	chatID := ctx.EffectiveChat.Id
	uc := sendchain.New(h.fetcher, tweet.Sender{Bot: b})
	if sendErr := uc.SendChain(reqCtx, chatID, replyToMsgID, username, tweetID, shared.UserDisplayName(&cb.From)); sendErr != nil {
		h.log.Error("failed to send chain for tweet %s: %v", tweetID, sendErr)
		return nil
	}

	if _, delErr := cb.Message.Delete(b, nil); delErr != nil {
		h.log.Debug("failed to delete original message: %v", delErr)
	}

	return nil
}

// Delete handles callbacks that remove a previously sent tweet message.
func (h *Handlers) Delete(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery

	deleteData, ok := tweet.DecodeDeleteCallback(cb.Data)
	if !ok {
		h.log.Error("failed to decode delete callback: %s", cb.Data)
		_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Invalid callback data",
		})
		return err
	}

	h.log.Info("delete callback - msg_id: %d, has_chain: %v", deleteData.MsgID, deleteData.HasChain)

	chatID := ctx.EffectiveChat.Id

	if _, err := b.DeleteMessage(chatID, deleteData.MsgID, nil); err != nil {
		h.log.Debug("failed to delete original message %d: %v", deleteData.MsgID, err)
		_, answerErr := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Cannot delete message",
		})
		return answerErr
	}

	botMsgID := cb.Message.GetMessageId()

	if deleteData.HasChain {
		chainCallbackData := tweet.EncodeChainCallback(
			deleteData.ChainUsername,
			deleteData.ChainTweetID,
			deleteData.MsgID,
		)
		if _, _, editErr := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
			ChatId:      chatID,
			MessageId:   botMsgID,
			ReplyMarkup: *tweet.BuildChainOnlyKeyboard(chainCallbackData),
		}); editErr != nil {
			h.log.Debug("failed to edit reply markup: %v", editErr)
		}
	} else {
		if _, _, editErr := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
			ChatId:      chatID,
			MessageId:   botMsgID,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
		}); editErr != nil {
			h.log.Debug("failed to remove reply markup: %v", editErr)
		}
	}

	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Deleted",
	})
	return err
}
