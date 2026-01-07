package handlers

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/chain"
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
	chainOpts := &tweet.SendChainResponseOpts{
		RequesterUsername: userDisplayName(&cb.From),
	}
	return tweet.SendChainResponse(b, chatID, items, replyToMsgID, chainOpts)
}
