package start

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/handlers/shared"
)

// Handler replies to the /start command with a simple greeting.
func Handler(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, shared.HelpText, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})
	return err
}
