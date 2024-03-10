package telegram

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Change the webhook token every WEBHOOK_TOKEN_ROTATE_INTERVAL
func (bot *Bot) RotateWebhook() {
	bot.WebhookRotateInterval = 24 * time.Hour
	if interval := os.Getenv("WEBHOOK_TOKEN_ROTATE_INTERVAL"); interval != "" {
		if parsedInterval, err := time.ParseDuration(interval); err != nil {
			slog.Error("Failed to parse WEBHOOK_TOKEN_ROTATE_INTERVAL, defaulting to 24h", "msg", err)
		} else {
			bot.WebhookRotateInterval = parsedInterval
		}
	}

	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for {
		newResult := make([]byte, 30)
		var seededRand *rand.Rand = rand.New(
			rand.NewSource(time.Now().UnixNano()))
		for i := range newResult {
			newResult[i] = charset[seededRand.Intn(len(charset))]
		}
		if bot.WebhookRotateInterval == 0 {
			bot.setWebhook("")
			break
		}
		bot.WebhookToken = string(newResult)
		bot.setWebhook(bot.WebhookToken)

		slog.Info("Webhook token rotated")
		time.Sleep(bot.WebhookRotateInterval)
	}
}

// Set webhook for bot: <domain>/webhook/<token>
func (bot *Bot) setWebhook(token string) {
	retries := 3
	if retriesEnv := os.Getenv("RETRY_ATTEMPTS"); retriesEnv != "" {
		if retriesInt, err := strconv.Atoi(retriesEnv); err != nil {
			slog.Error("Failed to parse RETRY_ATTEMPTS, defaulting to 3", "msg", err)
		} else {
			retries = retriesInt
		}
	}

	for range retries {
		webhookUrl := os.Getenv("DOMAIN") + "/webhook/" + token
		if token == "" {
			webhookUrl = os.Getenv("DOMAIN") + "/webhook"
		}

		slog.Info("Setting webhook", "url", webhookUrl)

		resp, err := http.PostForm(
			"https://api.telegram.org/bot"+bot.BotToken+"/setWebhook",
			url.Values{"url": {webhookUrl}},
		)
		if err != nil {
			slog.Error("Failed to request to set webhook: ", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		respBody := struct {
			OK     bool   `json:"ok"`
			Result bool   `json:"result"`
			Desc   string `json:"description"`
		}{}
		if err := json.Unmarshal(body, &respBody); err != nil {
			slog.Error("Failed to unmarshal response: ", err)
		}
		if respBody.OK && respBody.Result {
			slog.Info("Webhook set successfully",
				"result", respBody.Result,
				"desc", respBody.Desc,
			)
			return
		}

		slog.Warn("Failed to set webhook, cooling down for 3 seconds", "desc", respBody.Desc)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("Failed to set webhook after %d retries\n", retries)
}

func (bot *Bot) DeleteWebhook() {
	// create the request
	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+bot.BotToken+"/deleteWebhook",
		url.Values{},
	)
	if err != nil {
		slog.Error("Failed to request to delete webhook: ", err)
	}
	defer resp.Body.Close()

	// read the request
	respBody := struct {
		OK        bool   `json:"ok"`
		Desc      string `json:"description"`
		ErrorCode int    `json:"error_code"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.Error("Failed to decode response: ", err)
	}

	// logging
	if respBody.OK {
		slog.Info("Webhook deleted", "description", respBody.Desc)
		return
	}
	slog.Warn("Can't delete webhook",
		"error_code", respBody.ErrorCode,
		"description", respBody.Desc,
	)
}

func (bot *Bot) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	// auth middleware
	if bot.WebhookRotateInterval != 0 {
		webhookToken := r.PathValue("token")
		if webhookToken != bot.WebhookToken {
			slog.Warn("Invalid webhook token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// parse, check non-empty request body, send to channel
	incoming := struct {
		UpdateID int     `json:"update_id"`
		Message  Message `json:"message"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		slog.Error("Failed to decode webhook request body: ", err)
	}
	if incoming.Message.Text == "" {
		return
	}
	bot.MessageQueue <- &incoming.Message
}
