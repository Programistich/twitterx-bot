package shared

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// UserDisplayName returns the display name for a Telegram user, mirroring the previous helper.
// It prefers FirstName + LastName and falls back to the username if needed.
func UserDisplayName(user *gotgbot.User) string {
	if user == nil {
		return ""
	}
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name != "" {
		return name
	}
	if user.Username != "" {
		return "@" + user.Username
	}
	return ""
}
