package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	BotToken       string
	Debug          bool
	TwitterXAPIURL string
	TelegramAPIURL string

	TelegraphAuthorName string
	TelegraphAuthorURL  string
}

func Load() (Config, error) {
	telegramAPIURL := strings.TrimSpace(os.Getenv("TELEGRAM_API_URL"))
	telegramAPIURL = strings.TrimRight(telegramAPIURL, "/")

	cfg := Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		Debug:          os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "True" || os.Getenv("DEBUG") == "1",
		TwitterXAPIURL: os.Getenv("TWITTERX_API_URL"),
		TelegramAPIURL: telegramAPIURL,

		TelegraphAuthorName: "TwitterX",
		TelegraphAuthorURL:  "https://t.me/twitter_x_bot",
	}
	if cfg.TwitterXAPIURL == "" {
		return Config{}, errors.New("TwitterXAPIURL is required")
	}
	if cfg.BotToken == "" {
		return Config{}, errors.New("BOT_TOKEN is required")
	}
	return cfg, nil
}
