package main

import (
	"log/slog"
	"os"
	"time"

	"social-2-telego/message_listener"
	"social-2-telego/telegram"
	"social-2-telego/utils"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func getLogLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      getLogLevel(),
			TimeFormat: time.RFC1123Z,
		}),
	))
	if err := godotenv.Load(); err != nil {
		slog.Info(err.Error())
	}
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
	message_listener.InitMessageListener(appState)
}
