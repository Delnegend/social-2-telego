package social

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type FA struct {
	rawContent string
	url        string
	embedUrl   string
}

func (f *FA) Scrape() error {
	req, err := http.NewRequest("GET", f.embedUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "TelegramBot (like FuraffinityBot)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	f.rawContent = string(body)
	return nil
}

func (f *FA) GetContent() (string, error) {
	if f.rawContent == "" {
		if err := f.Scrape(); err != nil {
			return "", err
		}
	}
	return "", nil
}

func (f *FA) GetMedia() ([]ScrapedMedia, error) {
	if f.rawContent == "" {
		if err := f.Scrape(); err != nil {
			return nil, err
		}
	}
	return []ScrapedMedia{}, nil
}

func (f *FA) GetName() (string, error) {
	return "", nil
}

func (f *FA) GetUsername() (string, error) {
	return "", nil
}

// ===== URL stuffs ======

func (f *FA) SetURL(url string) error {
	if err := f.isValidURL(url); err != nil {
		return err
	}

	f.url = url
	f.embedUrl = strings.Replace(url, "furaffinity.net", "fxfuraffinity.net", 1)

	return nil
}

func (f *FA) GetURL() (string, error) {
	if f.url == "" {
		return "", fmt.Errorf("url is not set")
	}
	return f.url, nil
}

func (f *FA) GetEmbedUrl() (string, error) {
	if f.embedUrl == "" {
		return "", fmt.Errorf("url is not set")
	}
	return f.embedUrl, nil
}

func (f *FA) isValidURL(url string) error {
	pattern := `https:\/\/www\.furaffinity\.net\/view\/\d+`
	if !regexp.MustCompile(pattern).MatchString(url) {
		return fmt.Errorf("invalid url for furaffinity")
	}

	return nil
}