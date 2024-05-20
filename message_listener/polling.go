package message_listener

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"social-2-telego/utils"
	"strconv"
	"strings"
	"time"
)

// Get one single update
func (ml *MessageListener) getOneUpdate() {
	resp, err := http.Get("https://api.telegram.org/bot" + ml.appState.BotToken() + "/getUpdates?offset=" + strconv.Itoa(ml.offset))
	if err != nil {
		slog.Error("failed to request to get updates: ", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body: ", err)
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
		slog.Error("failed to unmarshal response body: ", err)
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
	delayDur := 3 * time.Second
	if delayDurEnv := os.Getenv("GET_UPDATE_DELAY"); delayDurEnv != "" {
		parsedDelayDurEnv, err := time.ParseDuration(delayDurEnv)
		if err != nil {
			slog.Error("failed to parse GET_UPDATE_DELAY, defaulting to 3s", "msg", err)
		} else {
			delayDur = parsedDelayDurEnv
		}
	}

	for {
		ml.getOneUpdate()
		time.Sleep(delayDur)
	}
}
