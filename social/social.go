package social

import (
	"regexp"
	"social-2-telego/utils"
	"strings"
)

var (
	htmlUrlRgx            = regexp.MustCompile(`<a href="([^"]+)"[^>]*>([^<]+)</a>`)
	htmlUrlPlaceholderRgx = regexp.MustCompile(`HLSTART ([^ ]+) HLSPLIT ([^ ]+) HLEND`)
	htmlParaRgx           = regexp.MustCompile(`<p>([^<]+)</p>`)
)

type Social interface {
	SetAppState(appState *utils.AppState)
	SetURL(url string) error

	GetMarkdownContent() (func(string) string, error)
	GetUsername() (string, error)
	GetMedia() ([]ScrapedMedia, error)
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
		return &X{}
	case strings.HasPrefix(url, "https://x.com/"):
		return &X{}
	case strings.HasPrefix(url, "https://www.furaffinity.net/view/"):
		return &FA{}
	default:
		return nil
	}
}
