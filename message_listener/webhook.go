package message_listener

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"social-2-telego/utils"
	"strings"
)

func (ml *MessageListener) setWebhook() {
	for range ml.appState.GetRetrySetWebhookAttempt() {
		webhookUrl := ml.appState.GetWebhookDomain() + "/webhook"
		slog.Info("setting webhook", "url", webhookUrl)

		// create a request
		resp, err := http.PostForm(
			"https://api.telegram.org/bot"+ml.appState.GetBotToken()+"/setWebhook",
			url.Values{
				"url":          {webhookUrl},
				"secret_token": {ml.appState.GetWebhookSecret()},
			},
		)
		if err != nil {
			slog.Error("failed to set webhook: ", "err", err)
			continue
		}

		// read the response
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		respBody := struct {
			OK     bool   `json:"ok"`
			Result bool   `json:"result"`
			Desc   string `json:"description"`
		}{}
		if err := json.Unmarshal(body, &respBody); err != nil {
			slog.Error("failed to unmarshal response when setting webhook: ", "err", err)
			continue
		}

		// handle the response
		if respBody.OK && respBody.Result {
			slog.Info("webhook set successfully",
				"result", respBody.Result,
				"desc", respBody.Desc,
			)
			return
		}
		slog.Error("failed to set webhook, retrying...")
	}

	log.Fatalf("failed to set webhook after %d retries\n", ml.appState.GetRetrySetWebhookAttempt())
}

func (ml *MessageListener) deleteWebhook() {
	// create the request
	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+ml.appState.GetBotToken()+"/deleteWebhook",
		url.Values{},
	)
	if err != nil {
		slog.Error("failed to request to delete webhook: ", "err", err)
	}
	defer resp.Body.Close()

	// read the request
	respBody := struct {
		OK        bool   `json:"ok"`
		Desc      string `json:"description"`
		ErrorCode int    `json:"error_code"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.Error("failed to decode response: ", "err", err)
	}

	// logging
	if respBody.OK {
		slog.Info("webhook deleted", "description", respBody.Desc)
		return
	}
	slog.Warn("can't delete webhook",
		"error_code", respBody.ErrorCode,
		"description", respBody.Desc,
	)
}

func (ml *MessageListener) handleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	// authenticate the request
	if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != ml.appState.GetWebhookSecret() {
		slog.Info("invalid webhook secret token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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
		slog.Error("failed to decode webhook request body: ", "err", err)
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
