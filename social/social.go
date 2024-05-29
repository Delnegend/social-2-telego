package social

import "strings"

type Social interface {
	// Scraping
	Scrape() error
	GetContent() (string, error)
	GetMedia() ([]ScrapedMedia, error)
	GetName() (string, error)
	GetUsername() (string, error)

	// Basic URL stuffs
	SetURL(string) error
	GetURL() (string, error)
	GetProfileURL(string) (string, error)
}

type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVideo MediaType = "video"
)

type ScrapedMedia struct {
	MediaType MediaType
	MediaUrl  string
}

func NewSocialInstance(url string) Social {
	switch {
	case strings.HasPrefix(url, "https://twitter.com/"):
		return &Twitter{}
	case strings.HasPrefix(url, "https://x.com/"):
		return &Twitter{}
	case strings.HasPrefix(url, "https://www.furaffinity.net/view/"):
		return &FA{}
	default:
		return nil
	}
}
