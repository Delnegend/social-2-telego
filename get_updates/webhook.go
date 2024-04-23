package get_update

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"social-2-telego/telegram"
	"strconv"
	"strings"
	"time"
)

// Generate a new token
func NewToken() string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 30)
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}

// Change the webhook token every WEBHOOK_TOKEN_ROTATE_INTERVAL
func (config *Config) RotateWebhook() {
	rotateInterval := 24 * time.Hour

	intervalEnv := os.Getenv("WEBHOOK_TOKEN_ROTATE_INTERVAL")
	parsedInterval, err := time.ParseDuration(intervalEnv)
	if err != nil {
		slog.Error("Failed to parse WEBHOOK_TOKEN_ROTATE_INTERVAL, defaulting to 24h", "msg", err)
	} else {
		rotateInterval = parsedInterval
	}

	for {
		if rotateInterval == 0 {
			config.WebhookToken = ""
			config.setWebhookWithRetry()
			break
		}
		config.WebhookToken = NewToken()
		config.setWebhookWithRetry()

		slog.Info("Webhook token rotated")
		time.Sleep(rotateInterval)
	}
}

func (config *Config) setWebhookWithRetry() {
	retries := 3
	if retriesEnv := os.Getenv("RETRY_ATTEMPTS"); retriesEnv != "" {
		if retriesInt, err := strconv.Atoi(retriesEnv); err != nil {
			slog.Error("Failed to parse RETRY_ATTEMPTS, defaulting to 3", "msg", err)
		} else {
			retries = retriesInt
		}
	}
	for range retries {
		if err := config.setWebhook(); err != nil {
			slog.Warn("failed to set webhook, cooling down for 3 seconds", "msg", err)
			time.Sleep(5 * time.Second)
		} else {
			return
		}
	}

	log.Fatalf("failed to set webhook after %d retries\n", retries)
}

// Set webhook for bot: <domain>/webhook/<token>
func (config *Config) setWebhook() error {
	webhookUrl := os.Getenv("DOMAIN") + "/webhook"
	if config.WebhookToken != "" {
		webhookUrl = os.Getenv("DOMAIN") + "/webhook/" + config.WebhookToken
	}

	slog.Info("Setting webhook", "url", webhookUrl)

	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+config.BotToken+"/setWebhook",
		url.Values{"url": {webhookUrl}},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	respBody := struct {
		OK     bool   `json:"ok"`
		Result bool   `json:"result"`
		Desc   string `json:"description"`
	}{}
	if err := json.Unmarshal(body, &respBody); err != nil {
		return fmt.Errorf("failed to unmarshal response when setting webhook: %w", err)
	}

	if respBody.OK && respBody.Result {
		slog.Info("webhook set successfully",
			"result", respBody.Result,
			"desc", respBody.Desc,
		)
		return nil
	}
	return fmt.Errorf("failed to set webhook: %s", respBody.Desc)
}

func (config *Config) DeleteWebhook() {
	// create the request
	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+config.BotToken+"/deleteWebhook",
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

func (config *Config) HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	// check for the webhook token if set
	if config.WebhookToken != "" {
		webhookToken := r.PathValue("token")
		if webhookToken != config.WebhookToken {
			slog.Warn("Invalid webhook token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// parse, check non-empty request body, send to channel
	incoming := struct {
		UpdateID int              `json:"update_id"`
		Message  telegram.Message `json:"message"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		slog.Error("Failed to decode webhook request body: ", err)
	}
	if incoming.Message.Text == "" {
		return
	}

	messages := strings.Split(incoming.Message.Text, "\n")
	for _, message := range messages {
		if message == "" {
			continue
		}
		*config.MessageQueue <- &telegram.Message{
			Chat: incoming.Message.Chat,
			Text: message,
			From: incoming.Message.From,
		}
	}
}
