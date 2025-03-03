package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	newsocial "social-2-telego/backend/new_social"
	"social-2-telego/backend/utils"
)

// Continuously watching new messages from MessageQueue and respond to them
func Responder(appState *utils.AppState) {
	var wg sync.WaitGroup
	wg.Add(1)

	for i := 0; i < appState.GetNumWorkers(); i++ {
		go func() {
			for msg := range appState.MsgQueue {
				slog.Debug("received message", "from", msg.From.Username, "text", msg.Text)

				if !appState.IsAuthorized(msg.From.Username) {
					slog.Warn("unauthorized user", "username", msg.From.Username)
				}

				postURL := strings.TrimSpace(msg.Text)
				if postURL == "" {
					continue
				}

				// scrape the content and media
				var scrapeResult *newsocial.ScrapeResult
				var err error
				switch {
				case strings.HasPrefix(postURL, "https://x.com/"):
					scrapeResult, err = newsocial.ScrapeX(appState, postURL)
				case strings.HasPrefix(postURL, "https://www.furaffinity.net/view/"):
					scrapeResult, err = newsocial.ScrapeFA(appState, postURL)
				default:
					slog.Debug("unsupported post URL", "url", postURL)
					continue
				}
				if err != nil {
					slog.Error("failed to scrape", "err", err)
					continue
				}

				data, sendType, err := composeMessage(
					appState.GetTargetChannel(
						strconv.Itoa(msg.From.ID),
					),
					postURL,
					scrapeResult,
				)
				if err != nil {
					slog.Error("failed to compose message", "err", err)
					continue
				}

				// init the request
				url := fmt.Sprintf(
					"https://api.telegram.org/bot%s/%s",
					appState.GetBotToken(),
					sendType,
				)
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
					slog.Error("message not sent",
						"error_code",
						respBody.ErrorCode,
						"description",
						respBody.Description,
					)
				}
			}
		}()
	}

	wg.Wait()
	log.Fatal("responder stopped for some reason, this should not happen")
}
