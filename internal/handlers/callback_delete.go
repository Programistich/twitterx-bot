package handlers

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/telegram/tweet"
)

func (h *Handlers) deleteCallback(b *gotgbot.Bot, ctx *ext.Context) error {
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

	// Try to delete the original user message (may fail if bot is not admin)
	_, err := b.DeleteMessage(chatID, deleteData.MsgID, nil)
	if err != nil {
		h.log.Debug("failed to delete original message %d: %v", deleteData.MsgID, err)
		_, answerErr := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Cannot delete message",
		})
		return answerErr
	}

	// Get the bot's message ID that contains the callback button
	botMsgID := cb.Message.GetMessageId()

	// Edit keyboard based on whether there's a chain button
	if deleteData.HasChain {
		// Has chain - rebuild chain button with the stored info
		chainCallbackData := tweet.EncodeChainCallback(
			deleteData.ChainUsername,
			deleteData.ChainTweetID,
			deleteData.MsgID,
		)
		_, _, editErr := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
			ChatId:      chatID,
			MessageId:   botMsgID,
			ReplyMarkup: *tweet.BuildChainOnlyKeyboard(chainCallbackData),
		})
		if editErr != nil {
			h.log.Debug("failed to edit reply markup: %v", editErr)
		}
	} else {
		// No chain button - remove keyboard entirely
		_, _, editErr := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
			ChatId:      chatID,
			MessageId:   botMsgID,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
		})
		if editErr != nil {
			h.log.Debug("failed to remove reply markup: %v", editErr)
		}
	}

	_, err = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "Deleted",
	})
	return err
}
