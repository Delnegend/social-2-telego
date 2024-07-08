package message_listener

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"social-2-telego/utils"
	"strings"
	"time"
)

// Get one single update
func (ml *MessageListener) getOneUpdate() {
	path_ := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d", ml.appState.GetBotToken(), ml.offset)
	resp, err := http.Get(path_)
	if err != nil {
		slog.Error("failed to request to get updates: ", "err", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body: ", "err", err)
		return
	}

	var respBody struct {
		Ok     bool `json:"ok"`
		Result []struct {
			UpdateID int                   `json:"update_id"`
			Message  utils.IncomingMessage `json:"message"`
		} `json:"result"`
	}
	if err = json.Unmarshal(body, &respBody); err != nil {
		slog.Error("failed to unmarshal response body: ", "err", err)
		return
	}

	if len(respBody.Result) > 0 {
		ml.offset = respBody.Result[len(respBody.Result)-1].UpdateID + 1
	}

	for _, result := range respBody.Result {
		messages := strings.Split(result.Message.Text, "\n")
		for _, message := range messages {
			if message == "" {
				continue
			}
			ml.appState.MsgQueue <- utils.IncomingMessage{
				Chat: result.Message.Chat,
				Text: message,
				From: result.Message.From,
			}
		}
	}
}

// Continuously get updates
func (ml *MessageListener) GetUpdates() {
	for {
		ml.getOneUpdate()
		time.Sleep(ml.appState.GetGetUpdatesInterval())
	}
}
