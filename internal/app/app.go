package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"twitterx-bot/internal/config"
	"twitterx-bot/internal/database"
	"twitterx-bot/internal/handlers"
	"twitterx-bot/internal/logger"
	"twitterx-bot/internal/telegraph"
	"twitterx-bot/internal/translation"
	"twitterx-bot/internal/twitterxapi"
)

func NewBot() (*gotgbot.Bot, *ext.Updater, *logger.Logger, *database.DB, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	l := logger.New(cfg.Debug)
	log := l.With("component", "app")
	log.Info("config loaded", "debug", cfg.Debug, "twitterx_api_url", cfg.TwitterXAPIURL, "telegram_api_url", cfg.TelegramAPIURL)

	// Initialize database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("init database: %w", err)
	}
	log.Info("database connected")

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.Migrate(ctx); err != nil {
		db.Close()
		return nil, nil, nil, nil, fmt.Errorf("run migrations: %w", err)
	}
	log.Info("database migrations completed")

	// Initialize Telegraph service if enabled
	var telegraphService *telegraph.Service
	telegraphClient := telegraph.NewClient(&http.Client{
		Timeout: 30 * time.Second,
	}, "")
	opts := []telegraph.Option{}
	opts = append(opts, telegraph.WithAuthorName(cfg.TelegraphAuthorName))
	opts = append(opts, telegraph.WithAuthorURL(cfg.TelegraphAuthorURL))
	telegraphService = telegraph.NewService(telegraphClient, opts...)
	log.Info("telegraph integration enabled", "author_name", cfg.TelegraphAuthorName, "author_url", cfg.TelegraphAuthorURL)

	botOpts := &gotgbot.BotOpts{}
	if cfg.TelegramAPIURL != "" {
		botOpts.BotClient = &gotgbot.BaseBotClient{
			DefaultRequestOpts: &gotgbot.RequestOpts{
				APIURL: cfg.TelegramAPIURL,
			},
		}
	}

	bot, err := gotgbot.NewBot(cfg.BotToken, botOpts)
	if err != nil {
		db.Close()
		return nil, nil, nil, nil, fmt.Errorf("init bot: %w", err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			dlog := l.With("component", "dispatcher")
			if ctx != nil {
				if ctx.EffectiveChat != nil {
					dlog = dlog.With("chat_id", ctx.EffectiveChat.Id)
				}
				if ctx.EffectiveUser != nil {
					dlog = dlog.With("user_id", ctx.EffectiveUser.Id, "username", ctx.EffectiveUser.Username)
				}
			}
			dlog.Error("handler error", "err", err)
			return ext.DispatcherActionNoop
		},
	})
	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{})

	chatSettingsRepo := database.NewChatSettingsRepository(db)
	apiClient := twitterxapi.NewClient(cfg.TwitterXAPIURL)

	// Initialize translation service for auto-translating tweet text into the chat language.
	translationService := translation.NewService(&http.Client{Timeout: 10 * time.Second}, "")
	log.Info("translation integration enabled")

	handlers.Register(dispatcher, l, apiClient, telegraphService, chatSettingsRepo, translationService)

	return bot, updater, l, db, nil
}

func Start(bot *gotgbot.Bot, updater *ext.Updater, l *logger.Logger, db *database.DB) error {
	if bot == nil {
		return fmt.Errorf("start bot: bot is nil")
	}
	if updater == nil {
		return fmt.Errorf("start bot: updater is nil")
	}
	log := l.With("component", "app")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if _, err := bot.SetMyName(&gotgbot.SetMyNameOpts{Name: "TwitterX"}); err != nil {
		log.Error("set bot name failed", "err", err)
	}
	if _, err := bot.SetMyDescription(&gotgbot.SetMyDescriptionOpts{
		Description: "Telegram Bot for best read twitter/X tweets https://github.com/Programistich/twitterx-bot",
	}); err != nil {
		log.Error("set bot description failed", "err", err)
	}

	if err := updater.StartPolling(bot, &ext.PollingOpts{DropPendingUpdates: true}); err != nil {
		return fmt.Errorf("start polling: %w", err)
	}
	log.Info("bot started", "username", bot.User.Username)

	<-ctx.Done()
	log.Info("shutdown signal received")
	updater.Stop()
	log.Info("updater stopped")
	if db != nil {
		if err := db.Close(); err != nil {
			log.Error("close database failed", "err", err)
		}
		log.Info("database connection closed")
	}
	return nil
}

func Run() error {
	bot, updater, l, db, err := NewBot()
	if err != nil {
		return err
	}

	return Start(bot, updater, l, db)
}
