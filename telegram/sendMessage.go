package telegram

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type MessageNotSendResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type MessageSentResponse struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Chat struct {
			ID int `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

func (bot *Bot) sendMessage(chatID string, text string) {
	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+bot.BotToken+"/sendMessage",
		url.Values{
			"chat_id":              {chatID},
			"text":                 {text},
			"parse_mode":           {"MarkdownV2"},
			"disable_notification": {"true"},
			"link_preview_options": {
				`{"prefer_large_media":true, "show_above_text":true}`,
			},
		},
	)
	if err != nil {
		slog.Error("Failed to send message", "err", err)
	}
	defer resp.Body.Close()

	var respBody struct {
		OK          bool    `json:"ok"`
		ErrorCode   int     `json:"error_code"`
		Description string  `json:"description"`
		Result      Message `json:"result"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &respBody); err != nil {
		slog.Error("failed to unmarshal response: ", err)
		return
	}

	if !respBody.OK {
		slog.Error("Message not sent", "error_code", respBody.ErrorCode, "description", respBody.Description)
		return
	}
}
