package start

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/handlers/shared"
	"twitterx-bot/internal/logger"
)

// Handler replies to the /start command with a simple greeting.
func Handler(b *gotgbot.Bot, ctx *ext.Context) error {
	log := logger.Default().With("component", "start")
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}
	if ctx.EffectiveUser != nil {
		log = log.With("user_id", ctx.EffectiveUser.Id, "username", ctx.EffectiveUser.Username)
	}
	log.Info("start command received")
	_, err := ctx.EffectiveMessage.Reply(b, shared.HelpText, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})
	if err != nil {
		log.Error("send start reply failed", "err", err)
	}
	return err
}
