package start

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

// Handler replies to the /start command with a simple greeting.
func Handler(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Hi! Send me any text and I will echo it back.", &gotgbot.SendMessageOpts{})
	return err
}
