package get_update

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"social-2-telego/telegram"
	"strconv"
	"strings"
	"time"
)

type PollingResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int              `json:"update_id"`
		Message  telegram.Message `json:"message"`
	} `json:"result"`
}

// Get one single update
func (config *Config) getUpdate() {
	resp, err := http.Get("https://api.telegram.org/bot" + config.BotToken + "/getUpdates?offset=" + strconv.Itoa(config.Offset))
	if err != nil {
		slog.Error("Failed to request to get updates: ", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body: ", err)
		return
	}

	var respBody PollingResponse
	if err = json.Unmarshal(body, &respBody); err != nil {
		slog.Error("Failed to unmarshal response body: ", err)
		return
	}

	if len(respBody.Result) > 0 {
		config.Offset = respBody.Result[len(respBody.Result)-1].UpdateID + 1
	}

	for _, result := range respBody.Result {
		messages := strings.Split(result.Message.Text, "\n")
		for _, message := range messages {
			if message == "" {
				continue
			}
			*config.MessageQueue <- &telegram.Message{
				Chat: result.Message.Chat,
				Text: message,
				From: result.Message.From,
			}
		}
	}
}

// Continuously get updates
func (config *Config) GetUpdates() {
	delayDur := 3 * time.Second
	if delayDurEnv := os.Getenv("GET_UPDATE_DELAY"); delayDurEnv != "" {
		parsedDelayDurEnv, err := time.ParseDuration(delayDurEnv)
		if err != nil {
			slog.Error("Failed to parse GET_UPDATE_DELAY, defaulting to 3s", "msg", err)
		} else {
			delayDur = parsedDelayDurEnv
		}
	}

	for {
		config.getUpdate()
		time.Sleep(delayDur)
	}
}
