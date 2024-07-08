package social

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"social-2-telego/utils"
	"strings"
)

var (
	faPostUrlRegex = regexp.MustCompile(`https:\/\/www\.furaffinity\.net\/view\/\d+`)
	faContentRegex = regexp.MustCompile(`(<div class="submission-description.+?>)((.|\n)*?)(</div>)`)
	faDownloadUrl  = regexp.MustCompile(`<div class="download"><a href="(.+?)">.+?</div>`)
	faUsernameRgx  = regexp.MustCompile(`submission-id-sub-container(.|\n)+?<strong>(.+?)</strong>`)
)

type FA struct {
	appState   *utils.AppState
	url        string
	rawContent string
}

// Set the AppState
func (f *FA) SetAppState(appState *utils.AppState) {
	f.appState = appState
}

// Set the URL of the post
func (f *FA) SetURL(url_ string) error {
	if !faPostUrlRegex.MatchString(url_) {
		return fmt.Errorf("FA.SetURL: invalid url for furaffinity")
	}
	f.url = url_
	return nil
}

// Scrape and save in `rawContent`
func (f *FA) scrape() error {
	// required condition
	if f.appState == nil {
		return fmt.Errorf("FA.scrape: appState is not set")
	}
	cookieA := f.appState.GetFaCookieA()
	cookieB := f.appState.GetFaCookieB()
	if cookieA == "" || cookieB == "" {
		return fmt.Errorf("FA.scrape: FA_COOKIE_A and FA_COOKIE_B are not set")
	}
	if f.url == "" {
		return fmt.Errorf("FA.scrape: url is not set")
	}

	// create new request
	req, err := http.NewRequest("GET", f.url, nil)
	if err != nil {
		return fmt.Errorf("FA.scrape: %w", err)
	}

	// set headers and cookies
	req.Header.Set("User-Agent", "TelegramBot (like FuraffinityBot)")
	req.AddCookie(&http.Cookie{Name: "a", Value: cookieA, Path: "/"})
	req.AddCookie(&http.Cookie{Name: "b", Value: cookieB, Path: "/"})

	// do the request & read the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("FA.scrape: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("FA.scrape: %w", err)
	}

	// store the response
	f.rawContent = string(body)
	return nil
}

// Get the post's content from `rawContent`
func (f *FA) GetMarkdownContent() (func(string) string, error) {
	if f.rawContent == "" {
		if err := f.scrape(); err != nil {
			return nil, fmt.Errorf("FA.GetMarkdownContent: %w", err)
		}
	}

	// extract the content
	slice := faContentRegex.FindStringSubmatch(f.rawContent)
	if len(slice) < 4 {
		return nil, fmt.Errorf("FA.GetMarkdownContent: expected 4 submatches from rawContent")
	}
	content := strings.TrimSpace(slice[2])

	// replace links and newlines with a markdown-like syntax to avoid
	// special characters being escaped into normal text
	content = strings.Replace(content, `<br />`, ``, -1)
	content = strings.Replace(content, "\n", "NEWLINE", -1)
	content = htmlUrlRgx.ReplaceAllString(content, `HLSTART $2 HLSPLIT $1 HLEND`)

	// escape special characters
	content = utils.EscapeSpecialChars(content, `ESCAPE_CHAR`)

	// replace those links with actual markdown syntax
	content = htmlUrlPlaceholderRgx.ReplaceAllString(content, "[$1]($2)")
	content = strings.ReplaceAll(content, `NEWLINE`, "\n>")

	return func(escapeChar string) string {
		return strings.Replace(content, `ESCAPE_CHAR`, escapeChar, -1)
	}, nil
}

// Get the post's owner's username
func (f *FA) GetUsername() (string, error) {
	if f.rawContent == "" {
		if err := f.scrape(); err != nil {
			return "", fmt.Errorf("FA.GetUsername: %w", err)
		}
	}
	slice := faUsernameRgx.FindStringSubmatch(f.rawContent)
	if len(slice) < 3 {
		return "", fmt.Errorf("FA.GetUsername: expected 2 submatches from rawContent")
	}
	if slice[2] == "" {
		return "", fmt.Errorf("FA.GetUsername: username is empty")
	}
	return slice[2], nil
}

// Get the media urls of the post from `rawContent`
func (f *FA) GetMedia() ([]ScrapedMedia, error) {
	if f.rawContent == "" {
		if err := f.scrape(); err != nil {
			return nil, fmt.Errorf("FA.GetMedia: %w", err)
		}
	}

	slice := faDownloadUrl.FindStringSubmatch(f.rawContent)
	if len(slice) < 2 {
		return nil, fmt.Errorf("FA.GetMedia: expected at least 2 submatches from rawContent")
	}
	mediaUrl := slice[1]
	if mediaUrl == "" {
		return nil, fmt.Errorf("FA.GetMedia: mediaUrl is empty")
	}
	if strings.HasPrefix(mediaUrl, "//") {
		mediaUrl = "https:" + mediaUrl
	}

	return []ScrapedMedia{
		{
			MediaType: MediaTypePhoto,
			MediaUrl:  mediaUrl,
		},
	}, nil
}
