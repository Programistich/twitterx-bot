package start

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/database"
	"twitterx-bot/internal/localization"
	"twitterx-bot/internal/logger"
)

// ChatSettingsProvider provides chat settings operations.
type ChatSettingsProvider interface {
	GetLanguage(ctx context.Context, chatID int64) (string, error)
	UpdateLanguage(ctx context.Context, chatID int64, language string) (*database.ChatSettings, error)
}

// Handler holds dependencies for the /start and /help commands.
type Handler struct {
	log          *logger.Logger
	chatSettings ChatSettingsProvider
}

// New creates a new start handler.
func New(log *logger.Logger, chatSettings ChatSettingsProvider) *Handler {
	return &Handler{log: log, chatSettings: chatSettings}
}

// Handle replies to the /start command with a localized greeting.
func (h *Handler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	log := h.log.With("component", "start")
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}
	if ctx.EffectiveUser != nil {
		log = log.With("user_id", ctx.EffectiveUser.Id, "username", ctx.EffectiveUser.Username)
	}
	log.Info("start command received")

	// Get language for this chat
	lang := database.DefaultLanguage
	if h.chatSettings != nil && ctx.EffectiveChat != nil {
		if l, err := h.chatSettings.GetLanguage(context.Background(), ctx.EffectiveChat.Id); err == nil {
			lang = l
		}
	}

	helpText := localization.Get(lang, localization.KeyHelpText)
	_, err := ctx.EffectiveMessage.Reply(b, helpText, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})
	if err != nil {
		log.Error("send start reply failed", "err", err)
	}
	return err
}
