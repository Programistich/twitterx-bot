package lang

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/database"
	"twitterx-bot/internal/localization"
	"twitterx-bot/internal/logger"
)

// CallbackPrefix is the prefix for language selection callbacks.
const CallbackPrefix = "lang:"

// Handler holds dependencies for the /lang command.
type Handler struct {
	log *logger.Logger
}

// New creates a new lang handler.
func New(log *logger.Logger) *Handler {
	return &Handler{log: log}
}

// Handle sends the language selection inline keyboard.
func (h *Handler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	log := h.log.With("component", "lang")
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}
	if ctx.EffectiveUser != nil {
		log = log.With("user_id", ctx.EffectiveUser.Id, "username", ctx.EffectiveUser.Username)
	}
	log.Info("lang command received")

	// Build inline keyboard with 3 language buttons
	keyboard := &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: localization.LanguageDisplayNames[database.LangUkrainian], CallbackData: CallbackPrefix + database.LangUkrainian},
				{Text: localization.LanguageDisplayNames[database.LangEnglish], CallbackData: CallbackPrefix + database.LangEnglish},
				{Text: localization.LanguageDisplayNames[database.LangRussian], CallbackData: CallbackPrefix + database.LangRussian},
			},
		},
	}

	// Use English as default for the initial prompt
	_, err := ctx.EffectiveMessage.Reply(b, localization.Get(database.LangEnglish, localization.KeySelectLanguage), &gotgbot.SendMessageOpts{
		ReplyMarkup: keyboard,
	})
	if err != nil {
		log.Error("send lang selection failed", "err", err)
	}
	return err
}
