package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/config"
	"twitterx-bot/internal/handlers"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/twitterxapi"
)

func NewBot() (*gotgbot.Bot, *ext.Updater, *logger.Logger, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, err
	}

	l := logger.New(cfg.Debug)

	bot, err := gotgbot.NewBot(cfg.BotToken, &gotgbot.BotOpts{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("init bot: %w", err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			l.Error("handler error: %v", err)
			return ext.DispatcherActionNoop
		},
	})
	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{})

	apiClient := twitterxapi.NewClient(cfg.TwitterXAPIURL)
	handlers.Register(dispatcher, l, apiClient)

	return bot, updater, l, nil
}

func Start(bot *gotgbot.Bot, updater *ext.Updater, l *logger.Logger) error {
	if bot == nil {
		return fmt.Errorf("start bot: bot is nil")
	}
	if updater == nil {
		return fmt.Errorf("start bot: updater is nil")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := updater.StartPolling(bot, &ext.PollingOpts{DropPendingUpdates: true}); err != nil {
		return fmt.Errorf("start polling: %w", err)
	}
	if l != nil {
		l.Info("bot started as @%s", bot.User.Username)
	}

	<-ctx.Done()
	updater.Stop()
	return nil
}

func Run() error {
	bot, updater, l, err := NewBot()
	if err != nil {
		return err
	}

	return Start(bot, updater, l)
}
