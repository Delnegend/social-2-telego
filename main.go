package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"social-2-telego/socials"
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
		slog.Warn("CHANNEL is not set, messages will be sent to the chat where the bot is added")
	}
	if chatId := os.Getenv("WHITELISTED"); chatId == "" {
		slog.Warn("WHITELISTED is not set, everyone can use the bot to send messages to your channel! You can add multiple usernames, @ not included, each separated by a space.")
	}
	port := os.Getenv("PORT")
	if port == "" {
		slog.Warn("PORT is not set, defaulting to 8080")
		os.Setenv("PORT", "8080")
	}

	mode := "polling"
	if enableWebhook := os.Getenv("ENABLE_WEBHOOK"); enableWebhook == "true" {
		if domain := os.Getenv("DOMAIN"); domain == "" {
			slog.Error("ENABLE_WEBHOOK is true, but DOMAIN is not set, defaulting to polling")
		} else {
			mode = "webhook"
		}
	}

	bot := telegram.Bot{
		BotToken:          token,
		TargetChannel:     &targetChannel,
		MessageQueue:      make(chan *telegram.Message, 10),
		PrefixSocialMatch: socials.PrefixSocialMatch(),
	}
	go bot.Responder()

	if mode == "webhook" {
		go bot.RotateWebhook()
		http.HandleFunc("POST /webhook/{token}", bot.HandleWebhookRequest)
		http.HandleFunc("POST /webhook", bot.HandleWebhookRequest)
	} else {
		bot.DeleteWebhook()
		go bot.GetUpdates()
	}

	http.ListenAndServe(":"+port, nil)
}
