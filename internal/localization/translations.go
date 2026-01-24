package localization

import "twitterx-bot/internal/database"

// translations holds all localized strings.
var translations = map[string]map[StringKey]string{
	database.LangEnglish: {
		KeyHelpText: `<b>TwitterX Bot</b>

<b>Direct Messages &amp; Groups</b>
Just send any Twitter/X link and I'll fetch the content for you.

<b>Inline Mode</b>
Use me in any chat by typing:
<code>@twitter_x_bot &lt;link&gt;</code>

Example:
<code>@twitter_x_bot https://x.com/user/status/123</code>

<b>Commands</b>
/start — Start the bot
/help — Show this message
/lang — Change language`,
		KeySelectLanguage:      "Select language:",
		KeyLanguageChanged:     "Language changed to: English",
		KeyInvalidCallbackData: "Invalid callback data",
		KeyFetchingFullChain:   "Fetching full chain...",
		KeyCannotDeleteMessage: "Cannot delete message",
		KeyDeleted:             "Deleted",
		KeySendFullChain:       "Send full chain",
		KeyDeleteOriginal:      "Delete original",
		KeyInvalidLanguage:     "Invalid language",
		KeyErrorSavingLanguage: "Error saving language",
		KeyVideo:               "Video",
		KeyMosaic:              "Mosaic",
		KeyPhoto:               "Photo",
		KeyTweet:               "Tweet",
		KeyTweetBy:             "Tweet by %s",
		KeyFrom:                " from %s",
		KeyBy:                  " by %s",
		KeyLinkFallback:        "%s",
	},
	database.LangUkrainian: {
		KeyHelpText: `<b>TwitterX Бот</b>

<b>Приватні повідомлення та групи</b>
Просто надішліть будь-яке посилання на Twitter/X, і я отримаю контент для вас.

<b>Інлайн режим</b>
Використовуйте мене в будь-якому чаті, набравши:
<code>@twitter_x_bot &lt;посилання&gt;</code>

Приклад:
<code>@twitter_x_bot https://x.com/user/status/123</code>

<b>Команди</b>
/start — Запустити бота
/help — Показати це повідомлення
/lang — Змінити мову`,
		KeySelectLanguage:      "Оберіть мову:",
		KeyLanguageChanged:     "Мову змінено на: Українська",
		KeyInvalidCallbackData: "Невірні дані",
		KeyFetchingFullChain:   "Отримую повний ланцюжок...",
		KeyCannotDeleteMessage: "Не вдалося видалити повідомлення",
		KeyDeleted:             "Видалено",
		KeySendFullChain:       "Надіслати весь ланцюжок",
		KeyDeleteOriginal:      "Видалити оригінал",
		KeyInvalidLanguage:     "Невірна мова",
		KeyErrorSavingLanguage: "Помилка збереження мови",
		KeyVideo:               "Відео",
		KeyMosaic:              "Мозаїка",
		KeyPhoto:               "Фото",
		KeyTweet:               "Твіт",
		KeyTweetBy:             "Твіт від %s",
		KeyFrom:                " від %s",
		KeyBy:                  " від %s",
		KeyLinkFallback:        "%s",
	},
	database.LangRussian: {
		KeyHelpText: `<b>TwitterX Бот</b>

<b>Личные сообщения и группы</b>
Просто отправьте любую ссылку на Twitter/X, и я получу контент для вас.

<b>Инлайн режим</b>
Используйте меня в любом чате, набрав:
<code>@twitter_x_bot &lt;ссылка&gt;</code>

Пример:
<code>@twitter_x_bot https://x.com/user/status/123</code>

<b>Команды</b>
/start — Запустить бота
/help — Показать это сообщение
/lang — Изменить язык`,
		KeySelectLanguage:      "Выберите язык:",
		KeyLanguageChanged:     "Язык изменён на: Русский",
		KeyInvalidCallbackData: "Неверные данные",
		KeyFetchingFullChain:   "Получаю полную цепочку...",
		KeyCannotDeleteMessage: "Не удалось удалить сообщение",
		KeyDeleted:             "Удалено",
		KeySendFullChain:       "Отправить всю цепочку",
		KeyDeleteOriginal:      "Удалить оригинал",
		KeyInvalidLanguage:     "Неверный язык",
		KeyErrorSavingLanguage: "Ошибка сохранения языка",
		KeyVideo:               "Видео",
		KeyMosaic:              "Мозаика",
		KeyPhoto:               "Фото",
		KeyTweet:               "Твит",
		KeyTweetBy:             "Твит от %s",
		KeyFrom:                " от %s",
		KeyBy:                  " от %s",
		KeyLinkFallback:        "%s",
	},
}

// Get returns the localized string for the given key and language.
// Falls back to English if the language or key is not found.
func Get(lang string, key StringKey) string {
	if langStrings, ok := translations[lang]; ok {
		if str, ok := langStrings[key]; ok {
			return str
		}
	}
	// Fallback to English
	if str, ok := translations[database.LangEnglish][key]; ok {
		return str
	}
	return string(key)
}

// LanguageDisplayNames returns the display name for languages (used in buttons).
var LanguageDisplayNames = map[string]string{
	database.LangUkrainian: "🇺🇦 Українська",
	database.LangEnglish:   "🇬🇧 English",
	database.LangRussian:   "🇷🇺 Русский",
}
