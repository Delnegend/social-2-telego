package main

import (
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	get_update "social-2-telego/get_updates"
	"social-2-telego/telegram"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN must be set")
	}

	// Colorful logging
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC1123Z,
		}),
	))

	targetChannel := os.Getenv("CHANNEL")
	if targetChannel == "" {
		slog.Info("CHANNEL is not set, messages will be echoed back to the user.")
	}

	whitelistedAccountsEnv := os.Getenv("WHITELISTED")
	whitelistedAccounts := strings.Split(whitelistedAccountsEnv, " ")
	if len(whitelistedAccounts) == 0 {
		slog.Info("WHITELISTED is not set, everyone can use the bot to send messages to your channel! You can add multiple usernames, @ not included, each separated by a space.")
	}

	bot := telegram.Config{
		BotToken:            token,
		TargetChannel:       &targetChannel,
		MessageQueue:        make(chan *telegram.Message, 10),
		WhitelistedAccounts: whitelistedAccounts,
	}
	go bot.Responder()

	update_getter := get_update.Config{
		BotToken:     token,
		MessageQueue: &bot.MessageQueue,
	}
	update_getter.Initialize()
}
