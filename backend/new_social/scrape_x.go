package newsocial

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"social-2-telego/backend/utils"
	"strings"
)

var (
	xPostUrlRegex = regexp.MustCompile(`https:\/\/((twitter)|x).com\/([\w_]{1,15})\/status\/\d+`)
	xContentRegex = regexp.MustCompile(`(<!-- Embed Status text -->)(.*?)(<!--)`)
	xVideoRgx     = regexp.MustCompile(`<video src="([^"]+)"`)
	xPostNotFound = regexp.MustCompile(`<meta property="og:description" content="Sorry, that post doesn't exist :\("\/>`)
)

func ScrapeX(appState *utils.AppState, url string) (*ScrapeResult, error) {
	// validate inputs
	if appState == nil {
		return nil, fmt.Errorf("ScrapeX: appState is not set")
	}
	if !xPostUrlRegex.MatchString(url) {
		return nil, fmt.Errorf("x.SetURL: invalid url for 𝕏")
	}

	rawContent, err := func() (*string, error) {
		path := strings.Replace(strings.Split(url, "?")[0], "x.com", "i.fxtwitter.com", 1)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "TelegramBot (like TwitterBot)")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		rawContent := html.UnescapeString(string(body))

		if xPostNotFound.MatchString(rawContent) {
			return nil, fmt.Errorf("post not found")
		}

		return &rawContent, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeX (rawContent): %w", err)
	}

	media, err := func() ([]ScrapedMedia, error) {
		// Get the section containing media
		start := "<!-- Embed media -->"
		startIndex := strings.Index(*rawContent, start) + len(start)
		endIndex := strings.Index((*rawContent)[startIndex:], "<!--")
		if startIndex < 0 || endIndex < 0 {
			return nil, fmt.Errorf("media not found")
		}
		content := (*rawContent)[startIndex : startIndex+endIndex]

		result := make([]ScrapedMedia, 0)

		// Get all the images
		pattern, err := regexp.Compile(`<img src="([^"]+)" />`)
		if err != nil {
			return nil, err
		}
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, v := range matches {
			result = append(result, ScrapedMedia{
				MediaType: MediaTypePhoto,
				MediaUrl:  v[1],
			})
		}

		// Get all the videos
		matches = xVideoRgx.FindAllStringSubmatch(content, -1)
		for _, v := range matches {
			result = append(result, ScrapedMedia{
				MediaType: MediaTypeVideo,
				MediaUrl:  v[1],
			})
		}
		return result, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeX (media): %w", err)
	}

	// NOTE: telegram use double backslash to escape special
	// characters for messages with 2+ photos/videos
	escapeChar := `\`
	if len(media) > 1 {
		escapeChar = `\\`
	}

	markdownContent, err := func() (*string, error) {
		// extract the content
		slice := xContentRegex.FindStringSubmatch(*rawContent)
		if len(slice) < 3 {
			return nil, fmt.Errorf("expected 3 submatches from rawContent")
		}

		// replace links and paragraphs with a markdown-like syntax to avoid
		// special characters being escaped into normal text
		content := htmlUrlRgx.ReplaceAllString(slice[2], "HLSTART $2 HLSPLIT $1 HLEND")
		paragraphs := make([]string, 0)
		for _, v := range htmlParaRgx.FindAllStringSubmatch(content, -1) {
			paragraphs = append(paragraphs, v[1])
		}
		content = strings.Join(paragraphs, `NEWLINE`)

		// escape special characters
		content = utils.EscapeSpecialChars(content, escapeChar)

		// replace those links and paragraphs with actual markdown syntax
		content = htmlUrlPlaceholderRgx.ReplaceAllString(content, "[$1]($2)")
		content = strings.ReplaceAll(content, `NEWLINE`, "\n\n")

		// add a space between consecutive links
		content = strings.ReplaceAll(content, `)[`, `) [`)

		return &content, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeX (content): %w", err)
	}

	username, err := func() (*string, error) {
		slice := strings.Split(url, "/")
		if len(slice) < 4 {
			return nil, fmt.Errorf("invalid URL to get username")
		}

		username := strings.TrimSpace(slice[3])
		if username == "" {
			return nil, fmt.Errorf("username is empty")
		}

		return &username, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("ScrapeX (username): %w", err)
	}

	profileURL := fmt.Sprintf("https://x.com/%s", *username)

	return &ScrapeResult{
		MarkdownContent: markdownContent,
		Username:        username,
		ProfileURL:      &profileURL,
		EscapeChar:      &escapeChar,
		Media:           media,
	}, nil
}
