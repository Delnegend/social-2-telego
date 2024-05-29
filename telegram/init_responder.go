package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"social-2-telego/social"
	"social-2-telego/utils"
)

func sendMessage(appState *utils.AppState, chatID string, text string, media []social.ScrapedMedia) {
	// Initialize the endpoint and data
	var endPoint string
	data := url.Values{
		"chat_id":              {chatID},
		"parse_mode":           {"MarkdownV2"},
		"disable_notification": {"true"},
	}

	// Decide the endpoint based on the number of media
	switch len(media) {
	case 0:
		endPoint = "sendMessage"
		data.Add("text", strings.Replace(text, `ESCAPE_CHAR`, `\`, -1))
	case 1:
		if media[0].MediaType == social.MediaTypePhoto {
			endPoint = "sendPhoto"
		} else {
			endPoint = "sendVideo"
		}
		data.Add(string(media[0].MediaType), media[0].MediaUrl)
		data.Add("caption", strings.Replace(text, `ESCAPE_CHAR`, `\`, -1))
	default:
		endPoint = "sendMediaGroup"
		result := make([]string, 0)

		// There's no "text", must add "caption" for the first media instead
		result = append(result,
			fmt.Sprintf(`{"type":"%s","media":"%s","caption":"%s","parse_mode":"MarkdownV2"}`,
				media[0].MediaType,
				media[0].MediaUrl,
				strings.Replace(text, `ESCAPE_CHAR`, `\\`, -1)))

		// Add the rest of the media to the result
		for _, media := range media[1:] {
			result = append(result,
				fmt.Sprintf(`{"type":"%s","media":"%s"}`,
					media.MediaType,
					media.MediaUrl))
		}

		data.Add("media", "["+strings.Join(result, ",")+"]")
	}

	// Initialize the request
	resp, err := http.PostForm(
		"https://api.telegram.org/bot"+appState.GetBotToken()+"/"+endPoint,
		data,
	)
	if err != nil {
		slog.Error("failed to send message", "err", err)
		return
	}
	defer resp.Body.Close()

	// Read the response
	var respBody struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
		// Result      string `json:"result"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &respBody); err != nil {
		slog.Error("failed to unmarshal response: ", err)
		return
	}

	// Logging
	if !respBody.OK {
		slog.Error("message not sent", "error_code", respBody.ErrorCode, "description", respBody.Description)
		return
	}
}

// Continuously watching new messages from MessageQueue and respond to them
func InitResponder(appState *utils.AppState) {
	var wg sync.WaitGroup
	wg.Add(1)

	for i := 0; i < appState.GetNumWorkers(); i++ {
		go func() {
			for message := range appState.MsgQueue {
				slog.Debug("received message", "from", message.From.Username, "text", message.Text)

				if !appState.IsAccountAllowed(message.From.Username) {
					slog.Warn("Unauthorized user", "username", message.From.Username)
				}

				socialCode := social.NewSocialInstance(message.Text)
				if socialCode == nil {
					slog.Warn("no social media matched")
					continue
				}

				// compose message
				outgoingText, err := ComposeMessage(message.Text, socialCode)
				if err != nil {
					slog.Error("failed to compose message", "err", err)
					sendMessage(appState, strconv.Itoa(message.Chat.ID), "failed to compose message: "+err.Error(), nil)
					continue
				}

				// send message
				media, err := socialCode.GetMedia()
				if err != nil {
					slog.Error("failed to get media", "err", err)
					continue
				}

				if appState.GetTargetChannel() != "" {
					sendMessage(appState, appState.GetTargetChannel(), outgoingText, media)
					continue
				}
				sendMessage(appState, strconv.Itoa(message.From.ID), outgoingText, media)
			}
		}()
	}

	wg.Wait()
	log.Fatal("responder stopped for some reason, this should not happen")
}
