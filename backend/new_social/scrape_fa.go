package newsocial

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"social-2-telego/backend/utils"
	"strings"
)

var (
	faPostUrlRegex = regexp.MustCompile(`https:\/\/www\.furaffinity\.net\/view\/\d+`)
	faContentRegex = regexp.MustCompile(`(<div class="submission-description.+?>)((.|\n)*?)(</div>)`)
	faDownloadUrl  = regexp.MustCompile(`<div class="download"><a href="(.+?)">.+?</div>`)
	faUsernameRgx  = regexp.MustCompile(`submission-id-sub-container(.|\n)+?<strong>(.+?)</strong>`)
)

func ScrapeFA(appState *utils.AppState, url string) (*ScrapeResult, error) {
	// validate inputs
	if appState == nil {
		return nil, fmt.Errorf("ScrapeFA: appState is not set")
	}
	cookieA := appState.GetFaCookieA()
	cookieB := appState.GetFaCookieB()
	if cookieA == "" || cookieB == "" {
		return nil, fmt.Errorf("FA.scrape: FA_COOKIE_A and FA_COOKIE_B are not set")
	}

	if !faPostUrlRegex.MatchString(url) {
		return nil, fmt.Errorf("ScrapeFA: invalid url for furaffinity")
	}

	req, err := func() (*http.Request, error) {
		// create new request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		// set headers and cookies
		req.Header.Set("User-Agent", "TelegramBot (like FuraffinityBot)")
		req.AddCookie(&http.Cookie{Name: "a", Value: cookieA, Path: "/"})
		req.AddCookie(&http.Cookie{Name: "b", Value: cookieB, Path: "/"})

		return req, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeFA: %w", err)
	}

	rawContent, err := func() (*string, error) {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		rawContent := string(body)
		if rawContent == "" {
			return nil, fmt.Errorf("ScrapeFA: empty response")
		}
		return &rawContent, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeFA: %w", err)
	}

	markdownContent, err := func() (*string, error) {
		slice := faContentRegex.FindStringSubmatch(*rawContent)
		if len(slice) < 4 {
			return nil, fmt.Errorf("expected 4 submatches from rawContent")
		}
		content := strings.TrimSpace(slice[2])
		// replace links and newlines with a markdown-like syntax to avoid
		// special characters being escaped into normal text
		content = strings.Replace(content, `<br />`, ``, -1)
		content = strings.Replace(content, "\n", "NEWLINE", -1)
		content = htmlUrlRgx.ReplaceAllString(content, `HLSTART $2 HLSPLIT $1 HLEND`)
		content = utils.EscapeSpecialChars(content, `\`)
		return &content, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeFA: %w", err)
	}

	username, err := func() (*string, error) {
		slice := faUsernameRgx.FindStringSubmatch(*rawContent)
		if len(slice) < 3 {
			return nil, fmt.Errorf("expected 2 submatches from rawContent")
		}
		if slice[2] == "" {
			return nil, fmt.Errorf("username is empty")
		}

		username := strings.TrimSpace(slice[2])
		if username == "" {
			return nil, fmt.Errorf("username is empty")
		}

		return &username, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeFA: %w", err)
	}

	mediaUrl, err := func() (*string, error) {
		slice := faDownloadUrl.FindStringSubmatch(*rawContent)
		if len(slice) < 2 {
			return nil, fmt.Errorf("expected at least 2 submatches from rawContent")
		}
		mediaUrl := slice[1]
		if mediaUrl == "" {
			return nil, fmt.Errorf("mediaUrl is empty")
		}
		if strings.HasPrefix(mediaUrl, "//") {
			mediaUrl = "https:" + mediaUrl
		}
		return &mediaUrl, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeFA: %w", err)
	}

	profileURL := fmt.Sprintf("https://www.furaffinity.net/view/%s", *username)

	escapeChar := `\`

	return &ScrapeResult{
		MarkdownContent: markdownContent,
		Username:        username,
		ProfileURL:      &profileURL,
		EscapeChar:      &escapeChar,
		Media: []ScrapedMedia{
			{
				MediaType: MediaTypePhoto,
				MediaUrl:  *mediaUrl,
			},
		},
	}, nil
}
