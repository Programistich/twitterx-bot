package handlers

import (
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// userDisplayName returns the display name for a Telegram user.
// Returns FirstName + LastName if available, otherwise @Username.
func userDisplayName(user *gotgbot.User) string {
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
