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
}

func Load() (Config, error) {
	telegramAPIURL := strings.TrimSpace(os.Getenv("TELEGRAM_API_URL"))
	telegramAPIURL = strings.TrimRight(telegramAPIURL, "/")

	cfg := Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		Debug:          os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "True" || os.Getenv("DEBUG") == "1",
		TwitterXAPIURL: os.Getenv("TWITTERX_API_URL"),
		TelegramAPIURL: telegramAPIURL,
	}
	if cfg.TwitterXAPIURL == "" {
		cfg.TwitterXAPIURL = "http://127.0.0.1:8080"
	}
	if cfg.BotToken == "" {
		return Config{}, errors.New("BOT_TOKEN is required")
	}
	return cfg, nil
}
