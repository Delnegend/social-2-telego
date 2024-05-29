package social

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"social-2-telego/utils"
	"strings"
)

type Twitter struct {
	rawContent string
	url        string
}

func (t *Twitter) Scrape() error {
	url_, err := t.GetURL()
	if err != nil {
		return err
	}
	path := strings.Replace(url_, "x.com", "fxtwitter.com", 1)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "TelegramBot (like TwitterBot)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	t.rawContent = html.UnescapeString(string(body))
	return nil
}

func (t *Twitter) GetMedia() ([]ScrapedMedia, error) {
	if t.rawContent == "" {
		if err := t.Scrape(); err != nil {
			return nil, err
		}
	}

	// Get the section containing media
	start := "<!-- Embed media -->"
	startIndex := strings.Index(t.rawContent, start) + len(start)
	endIndex := strings.Index(t.rawContent[startIndex:], "<!--")
	if startIndex < 0 || endIndex < 0 {
		return nil, errors.New("media not found")
	}
	content := t.rawContent[startIndex : startIndex+endIndex]

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
	pattern, err = regexp.Compile(`<video src="([^"]+)"`)
	if err != nil {
		return nil, err
	}
	matches = pattern.FindAllStringSubmatch(content, -1)
	for _, v := range matches {
		result = append(result, ScrapedMedia{
			MediaType: MediaTypeVideo,
			MediaUrl:  v[1],
		})
	}

	return result, nil
}

func (t *Twitter) GetContent() (string, error) {
	if t.rawContent == "" {
		if err := t.Scrape(); err != nil {
			return "", err
		}
	}

	start := "<!-- Embed Status text -->"
	startIndex := strings.Index(t.rawContent, start) + len(start)
	endIndex := strings.Index(t.rawContent[startIndex:], "<!--")
	if startIndex < 0 || endIndex < 0 {
		return "", errors.New("status text not found")
	}
	content := t.rawContent[startIndex : startIndex+endIndex]

	if content == "" {
		return "", nil
	}

	// replace all <a> tags with a custom wrapper to be processed later
	pattern, err := regexp.Compile(`<a href="([^"]+)"[^>]*>([^<]+)</a>`)
	if err != nil {
		return "", err
	}
	content = pattern.ReplaceAllString(content, "HLSTART $2 HLSPLIT $1 HLEND")

	// each <p> will be treated as a text block, each text block
	// are separated by 2 newline-characters
	sections := make([]string, 0)
	if pattern, err = regexp.Compile(`<p>([^<]+)</p>`); err != nil {
		return "", err
	} else {
		for _, v := range pattern.FindAllStringSubmatch(content, -1) {
			sections = append(sections, v[1])
		}
	}
	content = strings.Join(sections, "NEWLINE")

	content = utils.EscapeTelegramChar(content)

	pattern, err = regexp.Compile(`HLSTART ([^ ]+) HLSPLIT ([^ ]+) HLEND`)
	if err != nil {
		return "", err
	}
	content = pattern.ReplaceAllString(content, "[$1]($2)")
	content = strings.ReplaceAll(content, "NEWLINE", "\n>\n>")
	content = strings.ReplaceAll(content, ")[", ") [")

	return fmt.Sprintf(">%s", content), nil
}

func (t *Twitter) GetName() (string, error) {
	return t.GetUsername()
}

func (t *Twitter) GetUsername() (string, error) {
	if t.url == "" {
		return "", errors.New("url is not set")
	}

	slice := strings.Split(t.url, "/")
	if len(slice) < 4 {
		return "", errors.New("invalid URL")
	}
	return slice[3], nil
}

// ===== URL stuffs ===== //

func (t *Twitter) SetURL(url_ string) error {
	match, err := regexp.MatchString(`https:\/\/((twitter)|x).com\/([\w_]{1,15})\/status\/\d+`, url_)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("invalid url for twitter")
	}
	t.url = strings.Replace(strings.Split(url_, "?")[0], "twitter.com", "x.com", 1)
	return nil
}

func (t *Twitter) GetURL() (string, error) {
	if t.url == "" {
		return "", errors.New("url is not set")
	}
	return t.url, nil
}

func (t *Twitter) GetProfileURL(username string) (string, error) {
	if username == "" {
		return "", errors.New("username is empty")
	}
	return fmt.Sprintf("https://x.com/%s", username), nil
}
