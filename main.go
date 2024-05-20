package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"social-2-telego/message_listener"
	"social-2-telego/telegram"
	"social-2-telego/utils"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func init() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC1123Z,
		}),
	))
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	appState := utils.InitAppState()

	go telegram.InitResponder(appState)
	message_listener.InitMessageListender(appState)
}
