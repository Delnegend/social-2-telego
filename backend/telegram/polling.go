package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"social-2-telego/backend/utils"
	"strings"
	"time"
)

// Continuously get updates
func Polling(appState *utils.AppState) {
	offset := 0
	for {
		path := fmt.Sprintf(
			"https://api.telegram.org/bot%s/getUpdates?offset=%d",
			appState.GetBotToken(),
			offset,
		)
		resp, err := http.Get(path)
		if err != nil {
			slog.Error("failed to request to get updates: ", "err", err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			slog.Error("failed to read response body: ", "err", err)
			continue
		}
		resp.Body.Close()

		var respBody struct {
			Ok     bool `json:"ok"`
			Result []struct {
				UpdateID int                   `json:"update_id"`
				Message  utils.IncomingMessage `json:"message"`
			} `json:"result"`
		}
		if err = json.Unmarshal(body, &respBody); err != nil {
			slog.Error("failed to unmarshal response body: ", "err", err)
			continue
		}

		if len(respBody.Result) > 0 {
			offset = respBody.Result[len(respBody.Result)-1].UpdateID + 1
		}

		for _, result := range respBody.Result {
			messages := strings.Split(result.Message.Text, "\n")
			for _, message := range messages {
				if message == "" {
					continue
				}
				appState.MsgQueue <- utils.IncomingMessage{
					Chat: result.Message.Chat,
					Text: message,
					From: result.Message.From,
				}
			}
		}

		time.Sleep(appState.GetPollingInterval())
	}
}
