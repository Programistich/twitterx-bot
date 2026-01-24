package shared

import (
	"twitterx-bot/internal/database"
	"twitterx-bot/internal/localization"
)

// HelpText returns English help text for backwards compatibility.
var HelpText = localization.Get(database.LangEnglish, localization.KeyHelpText)

// GetHelpText returns localized help text.
func GetHelpText(lang string) string {
	return localization.Get(lang, localization.KeyHelpText)
}
