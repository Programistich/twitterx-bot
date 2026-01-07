package handlers

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/telegram/tweet"
)

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

	chatID := ctx.EffectiveChat.Id
	svc := h.svc.WithSender(tweet.Sender{Bot: b})
	if err := svc.SendChain(reqCtx, chatID, replyToMsgID, username, tweetID, userDisplayName(&cb.From)); err != nil {
		h.log.Error("failed to send chain for tweet %s: %v", tweetID, err)
		return nil
	}

	_, delErr := cb.Message.Delete(b, nil)
	if delErr != nil {
		h.log.Debug("failed to delete original message: %v", delErr)
	}

	return nil
}
