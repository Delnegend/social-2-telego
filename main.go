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
	appState := utils.NewAppState()

	go telegram.InitResponder(appState)
	message_listener.InitMessageListender(appState)
}
