package main

import (
	"log/slog"
	"os"
	"time"

	"social-2-telego/backend/telegram"
	"social-2-telego/backend/utils"

	"github.com/lmittmann/tint"
)

func init() {
	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.RFC1123Z,
		}),
	))
}

func main() {
	// This one contains all the environment variables
	// and a message channel to send messages to
	appState := utils.NewAppState()

	// This one listens to a channel and responds when there's a message, it's
	// where all the magic happens. When something goes wrong, it's likely to be
	// happening here
	go telegram.Responder(appState)

	// This one listens to updates from Telegram (webhook or long-polling) and
	// sends them to the message channel. This should not be breaking unless
	// Telegram changes their API
	switch appState.GetUseWebhook() {
	case true:
		telegram.Webhooking(appState)
	case false:
		slog.Info("polling updates")
		telegram.Polling(appState)
	}
}
