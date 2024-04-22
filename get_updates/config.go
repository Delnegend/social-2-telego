package get_update

import (
	"log/slog"
	"net/http"
	"os"
	"social-2-telego/telegram"
)

type Config struct {
	BotToken     string
	MessageQueue *chan *telegram.Message

	// For webhook
	WebhookToken string
	Domain       string

	// For polling
	Offset int
}

func (config *Config) Initialize() {
	config.Domain = os.Getenv("DOMAIN")
	config.Offset = 0

	if os.Getenv("ENABLE_WEBHOOK") == "true" {
		go config.RotateWebhook()
		http.HandleFunc("POST /webhook/{token}", config.HandleWebhookRequest)
		http.HandleFunc("POST /webhook", config.HandleWebhookRequest)
	} else {
		config.DeleteWebhook()
		go config.GetUpdates()
	}

	port := os.Getenv("PORT")
	if port == "" {
		slog.Warn("PORT is not set, defaulting to 8080")
		os.Setenv("PORT", "8080")
	}
	http.ListenAndServe(":"+port, nil)
}
