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
	telegraph    tweet.ArticleCreator
}

// New creates callback handlers with the configured logger, tweet fetcher, and chain timeout.
func New(log *logger.Logger, fetcher TweetFetcher, chainTimeout time.Duration, telegraph tweet.ArticleCreator) *Handlers {
	return &Handlers{log: log, fetcher: fetcher, chainTimeout: chainTimeout, telegraph: telegraph}
}

// Chain processes callback queries that request a tweet chain.
func (h *Handlers) Chain(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	log := h.log.With("component", "callback", "callback", "chain")
	if cb != nil {
		log = log.With("callback_id", cb.Id)
	}
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}
	if cb != nil {
		log = log.With("user_id", cb.From.Id, "username", cb.From.Username)
	}

	username, tweetID, replyToMsgID, ok := tweet.DecodeChainCallback(cb.Data)
	if !ok {
		log.Error("decode chain callback failed", "data", cb.Data)
		_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Invalid callback data",
		})
		return err
	}

	log = log.With("tweet_username", username, "tweet_id", tweetID, "reply_to_msg_id", replyToMsgID)
	log.Info("chain callback received")

	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Fetching full chain...",
	})
	if err != nil {
		log.Debug("answer callback failed", "err", err)
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), h.chainTimeout)
	defer cancel()

	chatID := ctx.EffectiveChat.Id
	uc := sendchain.New(h.fetcher, tweet.Sender{Bot: b, Telegraph: h.telegraph, Log: log})
	if sendErr := uc.SendChain(reqCtx, chatID, replyToMsgID, username, tweetID, shared.UserDisplayName(&cb.From)); sendErr != nil {
		log.Error("send chain failed", "err", sendErr)
		return nil
	}

	if _, delErr := cb.Message.Delete(b, nil); delErr != nil {
		log.Debug("delete original message failed", "err", delErr)
	}

	log.Info("chain sent")
	return nil
}

// Delete handles callbacks that remove a previously sent tweet message.
func (h *Handlers) Delete(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	log := h.log.With("component", "callback", "callback", "delete")
	if cb != nil {
		log = log.With("callback_id", cb.Id, "user_id", cb.From.Id, "username", cb.From.Username)
	}
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}

	deleteData, ok := tweet.DecodeDeleteCallback(cb.Data)
	if !ok {
		log.Error("decode delete callback failed", "data", cb.Data)
		_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Invalid callback data",
		})
		return err
	}

	log = log.With("msg_id", deleteData.MsgID, "has_chain", deleteData.HasChain)
	log.Info("delete callback received")

	chatID := ctx.EffectiveChat.Id

	if _, err := b.DeleteMessage(chatID, deleteData.MsgID, nil); err != nil {
		log.Debug("delete original message failed", "err", err)
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
			log.Debug("edit reply markup failed", "err", editErr)
		}
	} else {
		if _, _, editErr := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
			ChatId:      chatID,
			MessageId:   botMsgID,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
		}); editErr != nil {
			log.Debug("remove reply markup failed", "err", editErr)
		}
	}

	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Deleted",
	})
	if err == nil {
		log.Info("message deleted")
	}
	return err
}
