package telegram

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"social-2-telego/social"
	"social-2-telego/utils"
)

// Continuously watching new messages from MessageQueue and respond to them
func Responder(appState *utils.AppState) {
	var wg sync.WaitGroup
	wg.Add(1)

	// create a number of for loops, each inside a goroutine
	for i := 0; i < appState.GetNumWorkers(); i++ {
		go func() {
			for msg := range appState.MsgQueue {
				slog.Debug("received message", "from", msg.From.Username, "text", msg.Text)

				if !appState.IsAuthorized(msg.From.Username) {
					slog.Warn("unauthorized user", "username", msg.From.Username)
				}

				// match the input to a social struct for scraping
				matchedSocial := social.NewSocialInstance(msg.Text)
				if matchedSocial == nil {
					slog.Warn("no social media matched")
					continue
				}
				matchedSocial.SetAppState(appState)

				// split and cleanup the input
				slice := func() []string {
					slice := strings.Split(msg.Text, ",")
					temp := make([]string, 0)
					for _, item := range slice {
						item = strings.TrimSpace(item)
						if item != "" {
							temp = append(temp, item)
						}
					}
					return temp
				}()

				// analyze the input
				var postURL, authorInfo, hashtags string
				var err error
				switch len(slice) {
				case 0:
					slog.Warn("no post URL found")
					continue
				case 1:
					postURL = slice[0]
					if err := matchedSocial.SetURL(postURL); err != nil {
						slog.Warn("failed to set URL", "err", err)
						continue
					}
					authorInfo, err = matchedSocial.GetUsername()
					if err != nil {
						slog.Warn("failed to get author", "err", err)
						continue
					}
				case 2:
					postURL = slice[0]
					if err := matchedSocial.SetURL(postURL); err != nil {
						slog.Warn("failed to set URL", "err", err)
						continue
					}
					switch {
					case strings.HasPrefix(slice[1], "#"):
						hashtags = slice[1]
					default:
						authorInfo = slice[1]
					}
				case 3:
					postURL = slice[0]
					if err := matchedSocial.SetURL(postURL); err != nil {
						slog.Warn("failed to set URL", "err", err)
						continue
					}
					switch {
					case strings.HasPrefix(slice[1], "#"):
						authorInfo = slice[1]
						hashtags = slice[2]
					default:
						authorInfo = slice[1]
						hashtags = slice[2]
					}
				}
				if err := matchedSocial.SetURL(postURL); err != nil {
					slog.Warn("failed to set URL", "err", err)
					continue
				}

				// scrape the content and media
				mdContent, err := matchedSocial.GetMarkdownContent()
				if err != nil {
					slog.Error("failed to get HTML content", "err", err)
					continue
				}
				media, err := matchedSocial.GetMedia()
				if err != nil {
					slog.Error("failed to get media", "err", err)
					continue
				}

				// add necessary data to the message struct
				teleMsg := TelegramMessage{}
				teleMsg.
					SetContent(mdContent).
					SetArtistNameAndUsername(authorInfo).
					SetHashtags(hashtags).
					SetMedia(media).
					SetPostURL(postURL)

				// getting the target channel
				targetChannel := appState.GetTargetChannel()
				if targetChannel == "" {
					targetChannel = strconv.Itoa(msg.From.ID)
				}

				// from the message struct serialize everything to a complete
				// data package to be sent to Telegram
				data, endPoint, err := teleMsg.ToData(targetChannel)
				if err != nil {
					slog.Error("failed to compose message", "err", err)
					continue
				}

				// init the request
				url := "https://api.telegram.org/bot" + appState.GetBotToken() + "/" + string(endPoint)
				resp, err := http.PostForm(url, data)
				if err != nil {
					slog.Error("failed to send message", "err", err)
					continue
				}

				// read & log the response
				var respBody struct {
					OK          bool   `json:"ok"`
					ErrorCode   int    `json:"error_code"`
					Description string `json:"description"`
				}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					resp.Body.Close()
					slog.Error("failed to read response body: ", "err", err)
					continue
				}
				resp.Body.Close()
				if err := json.Unmarshal(body, &respBody); err != nil {
					slog.Error("failed to unmarshal response: ", "err", err)
					continue
				}
				if !respBody.OK {
					slog.Error("message not sent", "error_code", respBody.ErrorCode, "description", respBody.Description)
				}
			}
		}()
	}

	wg.Wait()
	log.Fatal("responder stopped for some reason, this should not happen")
}
