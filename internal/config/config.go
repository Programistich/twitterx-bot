package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken       string
	Debug          bool
	TwitterXAPIURL string
}

func Load() (Config, error) {
	cfg := Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		Debug:          os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "True" || os.Getenv("DEBUG") == "1",
		TwitterXAPIURL: os.Getenv("TWITTERX_API_URL"),
	}
	if cfg.TwitterXAPIURL == "" {
		cfg.TwitterXAPIURL = "http://127.0.0.1:8080"
	}
	if cfg.BotToken == "" {
		return Config{}, errors.New("BOT_TOKEN is required")
	}
	return cfg, nil
}
