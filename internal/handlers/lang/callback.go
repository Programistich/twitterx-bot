package lang

import (
	"context"
	"strings"

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

// CallbackHandler handles language selection callbacks.
type CallbackHandler struct {
	chatSettings ChatSettingsProvider
	log          *logger.Logger
}

// NewCallbackHandler creates a new callback handler for language selection.
func NewCallbackHandler(chatSettings ChatSettingsProvider, log *logger.Logger) *CallbackHandler {
	return &CallbackHandler{chatSettings: chatSettings, log: log}
}

// Handle processes the language selection callback.
func (h *CallbackHandler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	log := h.log.With("component", "lang_callback")
	if cb != nil {
		log = log.With("callback_id", cb.Id, "user_id", cb.From.Id, "username", cb.From.Username)
	}
	if ctx.EffectiveChat != nil {
		log = log.With("chat_id", ctx.EffectiveChat.Id)
	}

	// Parse callback data: "lang:uk", "lang:en", "lang:ru"
	langCode := strings.TrimPrefix(cb.Data, CallbackPrefix)
	if !database.IsValidLanguage(langCode) {
		log.Error("invalid language code", "code", langCode)
		// Get user's current language for error message
		currentLang := database.DefaultLanguage
		if ctx.EffectiveChat != nil {
			if l, err := h.chatSettings.GetLanguage(context.Background(), ctx.EffectiveChat.Id); err == nil {
				currentLang = l
			}
		}
		_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: localization.Get(currentLang, localization.KeyInvalidLanguage),
		})
		return err
	}

	log = log.With("lang", langCode)
	log.Info("language selection received")

	// Save language to database
	chatID := ctx.EffectiveChat.Id
	_, err := h.chatSettings.UpdateLanguage(context.Background(), chatID, langCode)
	if err != nil {
		log.Error("update language failed", "err", err)
		// Get user's current language for error message
		currentLang := database.DefaultLanguage
		if l, errLang := h.chatSettings.GetLanguage(context.Background(), chatID); errLang == nil {
			currentLang = l
		}
		_, _ = cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: localization.Get(currentLang, localization.KeyErrorSavingLanguage)})
		return err
	}

	// Answer callback (no toast, the message edit shows the result)
	_, _ = cb.Answer(b, nil)

	// Edit the bot's message to show confirmation in the selected language
	confirmText := localization.Get(langCode, localization.KeyLanguageChanged)
	_, _, err = cb.Message.EditText(b, confirmText, &gotgbot.EditMessageTextOpts{})
	if err != nil {
		log.Error("edit message failed", "err", err)
	}

	// Delete the user's /lang command message (it's the message that was replied to)
	if cb.Message != nil {
		if msg, ok := cb.Message.(*gotgbot.Message); ok && msg.ReplyToMessage != nil {
			if _, delErr := b.DeleteMessage(chatID, msg.ReplyToMessage.MessageId, nil); delErr != nil {
				log.Debug("delete user command message failed", "err", delErr)
			}
		}
	}

	log.Info("language changed", "lang", langCode)
	return nil
}
