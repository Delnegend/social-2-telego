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

	"social-2-telego/social"
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

func (bot *Config) sendMessage(chatID string, text string, media []social.ScrapedMedia) {
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
		"https://api.telegram.org/bot"+bot.BotToken+"/"+endPoint,
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

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// Continuously watching new messages from MessageQueue and respond to them
func (bot *Config) Responder() {
mainLoop:
	for message := range bot.MessageQueue {
		slog.Debug("received message", "from", message.From.Username, "text", message.Text)

		// check if the user is authorized
		if len(bot.WhitelistedAccounts) != 0 {
			if !contains(bot.WhitelistedAccounts, message.From.Username) {
				slog.Warn("Unauthorized user", "username", message.From.Username)
				continue mainLoop
			}
		}

		socialCode := social.NewSocialInstance(message.Text)
		if socialCode == nil {
			slog.Warn("no social media matched")
			continue mainLoop
		}

		// compose message
		outgoingText, err := ComposeMessage(message.Text, socialCode)
		if err != nil {
			slog.Error("failed to compose message", "err", err)
			bot.sendMessage(strconv.Itoa(message.Chat.ID), "failed to compose message: "+err.Error(), nil)
			continue mainLoop
		}

		// send message
		media, err := socialCode.GetMedia()
		if err != nil {
			slog.Error("failed to get media", "err", err)
			continue mainLoop
		}

		if *bot.TargetChannel != "" {
			bot.sendMessage(*bot.TargetChannel, outgoingText, media)
			continue mainLoop
		}
		bot.sendMessage(strconv.Itoa(message.From.ID), outgoingText, media)
	}
	log.Fatal("responder stopped for some reason, this should not happen")
}
