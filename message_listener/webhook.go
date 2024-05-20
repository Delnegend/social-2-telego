package message_listener

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
	"social-2-telego/utils"
	"strconv"
	"strings"
	"time"
)

// Generate a new token
func newToken() string {
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
func (ml *MessageListener) rotateWebhook() {
	rotateInterval := 24 * time.Hour

	intervalEnv := os.Getenv("WEBHOOK_TOKEN_ROTATE_INTERVAL")
	parsedInterval, err := time.ParseDuration(intervalEnv)
	if err != nil {
		slog.Error("failed to parse WEBHOOK_TOKEN_ROTATE_INTERVAL, defaulting to 24h", "msg", err)
	} else {
		rotateInterval = parsedInterval
	}

	for {
		if rotateInterval == 0 {
			ml.webhookToken = ""
			ml.setWebhookWithRetry()
			break
		}
		ml.webhookToken = newToken()
		ml.setWebhookWithRetry()

		slog.Info("Webhook token rotated")
		time.Sleep(rotateInterval)
	}
}

func (ml *MessageListener) setWebhookWithRetry() {
	retries := 3
	if retriesEnv := os.Getenv("RETRY_ATTEMPTS"); retriesEnv != "" {
		if retriesInt, err := strconv.Atoi(retriesEnv); err != nil {
			slog.Error("failed to parse RETRY_ATTEMPTS, defaulting to 3", "msg", err)
		} else {
			retries = retriesInt
		}
	}
	for range retries {
		if err := ml.setWebhook(); err != nil {
			slog.Warn("failed to set webhook, cooling down for 3 seconds", "msg", err)
			time.Sleep(5 * time.Second)
		} else {
			return
		}
	}

	log.Fatalf("failed to set webhook after %d retries\n", retries)
}

// Set webhook for bot: <domain>/webhook/<token>
func (ml *MessageListener) setWebhook() error {
	webhookUrl := os.Getenv("DOMAIN") + "/webhook"
	if ml.webhookToken != "" {
		webhookUrl = os.Getenv("DOMAIN") + "/webhook/" + ml.webhookToken
	}

	slog.Info("Setting webhook", "url", webhookUrl)

	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+ml.appState.BotToken()+"/setWebhook",
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

func (ml *MessageListener) deleteWebhook() {
	// create the request
	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+ml.appState.BotToken()+"/deleteWebhook",
		url.Values{},
	)
	if err != nil {
		slog.Error("failed to request to delete webhook: ", err)
	}
	defer resp.Body.Close()

	// read the request
	respBody := struct {
		OK        bool   `json:"ok"`
		Desc      string `json:"description"`
		ErrorCode int    `json:"error_code"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.Error("failed to decode response: ", err)
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

func (ml *MessageListener) handleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Received webhook request")

	// check for the webhook token if set
	if ml.webhookToken != "" {
		webhookToken := r.PathValue("token")
		if webhookToken != ml.webhookToken {
			slog.Warn("Invalid webhook token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// parse, check non-empty request body, send to channel
	incoming := struct {
		UpdateID int                   `json:"update_id"`
		Message  utils.IncomingMessage `json:"message"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		slog.Error("failed to decode webhook request body: ", err)
	}
	if incoming.Message.Text == "" {
		return
	}

	messages := strings.Split(incoming.Message.Text, "\n")
	for _, message := range messages {
		if message == "" {
			continue
		}
		ml.appState.MsgQueue <- utils.IncomingMessage{
			Chat: incoming.Message.Chat,
			Text: message,
			From: incoming.Message.From,
		}
	}
}
