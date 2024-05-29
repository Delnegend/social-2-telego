package social

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type FA struct {
	rawContent string
	url        string
}

func (f *FA) Scrape() error {
	url_, err := f.GetURL()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("GET", url_, nil)
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
	// if f.rawContent == "" {
	// 	if err := f.Scrape(); err != nil {
	// 		return "", err
	// 	}
	// }
	return "", fmt.Errorf("work in progress")
}

func (f *FA) GetMedia() ([]ScrapedMedia, error) {
	// if f.rawContent == "" {
	// 	if err := f.Scrape(); err != nil {
	// 		return nil, err
	// 	}
	// }
	// return []ScrapedMedia{}, nil

	return nil, fmt.Errorf("work in progress")
}

func (f *FA) GetName() (string, error) {
	return "", fmt.Errorf("work in progress")
}

func (f *FA) GetUsername() (string, error) {
	return "", fmt.Errorf("work in progress")
}

// ===== URL stuffs ======

func (f *FA) SetURL(url_ string) error {
	match, err := regexp.MatchString(`https:\/\/www\.furaffinity\.net\/view\/\d+`, url_)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("invalid url for furaffinity")
	}
	f.url = url_
	return nil
}

func (f *FA) GetURL() (string, error) {
	if f.url == "" {
		return "", fmt.Errorf("url is not set")
	}
	return f.url, nil
}

func (f *FA) GetProfileURL(username string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("username is empty")
	}
	return fmt.Sprintf("https://www.furaffinity.net/user/%s", username), nil
}
