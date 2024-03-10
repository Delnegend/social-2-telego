package telegram

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

type PollingResponse struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int     `json:"update_id"`
		Message  Message `json:"message"`
	} `json:"result"`
}

// Get one single update
func (bot *Bot) getUpdate() {
	resp, err := http.Get("https://api.telegram.org/bot" + bot.BotToken + "/getUpdates?offset=" + strconv.Itoa(bot.Offset))
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
		bot.Offset = respBody.Result[len(respBody.Result)-1].UpdateID + 1
	}

	for _, result := range respBody.Result {
		bot.MessageQueue <- &result.Message
	}
}

// Continuously get updates
func (bot *Bot) GetUpdates() {
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
		bot.getUpdate()
		time.Sleep(delayDur)
	}
}
