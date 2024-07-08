package social

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"social-2-telego/utils"
	"strings"
)

var (
	xPostUrlRegex = regexp.MustCompile(`https:\/\/((twitter)|x).com\/([\w_]{1,15})\/status\/\d+`)
	xContentRegex = regexp.MustCompile(`(<!-- Embed Status text -->)(.*?)(<!--)`)
	xRootDomain   = regexp.MustCompile(`(x|twitter).com`)
	xVideoRgx     = regexp.MustCompile(`<video src="([^"]+)"`)
)

type X struct {
	appState *utils.AppState

	rawContent string
	url        string
}

func (t *X) SetAppState(appState *utils.AppState) {
	t.appState = appState
}

// Set the URL of the post
func (t *X) SetURL(url_ string) error {
	if !xPostUrlRegex.MatchString(url_) {
		return fmt.Errorf("x.SetURL: invalid url for 𝕏")
	}
	t.url = strings.Replace(strings.Split(url_, "?")[0], "twitter.com", "x.com", 1) // TODO: remove
	return nil
}

// Scrape and save in `rawContent`
func (t *X) scrape() error {
	if t.url == "" {
		return fmt.Errorf("x.scrape: url is not set")
	}

	path := xRootDomain.ReplaceAllString(t.url, "i.fxtwitter.com")
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("x.scrape: %w", err)
	}

	req.Header.Set("User-Agent", "TelegramBot (like TwitterBot)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("x.scrape: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("x.scrape: %w", err)
	}

	t.rawContent = html.UnescapeString(string(body))
	return nil
}

// Get the post's content in MD format from `rawContent`. This returns a
// function that you need to provide the escape character
func (t *X) GetMarkdownContent() (func(string) string, error) {
	if t.rawContent == "" {
		if err := t.scrape(); err != nil {
			return nil, fmt.Errorf("x.GetMarkdownContent: %w", err)
		}
	}

	// extract the content
	slice := xContentRegex.FindStringSubmatch(t.rawContent)
	if len(slice) < 3 {
		return nil, fmt.Errorf("x.GetMarkdownContent: expected 3 submatches from rawContent")
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
	content = utils.EscapeSpecialChars(content, `ESCAPE_CHAR`)

	// replace those links and paragraphs with actual markdown syntax
	content = htmlUrlPlaceholderRgx.ReplaceAllString(content, "[$1]($2)")
	content = strings.ReplaceAll(content, `NEWLINE`, "\n>\n>")

	// add a space between consecutive links
	content = strings.ReplaceAll(content, `)[`, `) [`)

	return func(escapeChar string) string {
		return strings.Replace(content, `ESCAPE_CHAR`, escapeChar, -1)
	}, nil
}

func (t *X) GetUsername() (string, error) {
	if t.url == "" {
		return "", fmt.Errorf("x.GetUsername: url is not set")
	}

	slice := strings.Split(t.url, "/")
	if len(slice) < 4 {
		return "", fmt.Errorf("x.GetUsername: invalid URL")
	}
	return slice[3], nil
}

// Get the media urls of the post from `rawContent`
func (t *X) GetMedia() ([]ScrapedMedia, error) {
	if t.rawContent == "" {
		if err := t.scrape(); err != nil {
			return nil, fmt.Errorf("x.GetMedia: %w", err)
		}
	}

	// Get the section containing media
	start := "<!-- Embed media -->"
	startIndex := strings.Index(t.rawContent, start) + len(start)
	endIndex := strings.Index(t.rawContent[startIndex:], "<!--")
	if startIndex < 0 || endIndex < 0 {
		return nil, fmt.Errorf("x.GetMedia: media not found")
	}
	content := t.rawContent[startIndex : startIndex+endIndex]

	result := make([]ScrapedMedia, 0)

	// Get all the images
	pattern, err := regexp.Compile(`<img src="([^"]+)" />`)
	if err != nil {
		return nil, fmt.Errorf("x.GetMedia: %w", err)
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
}
