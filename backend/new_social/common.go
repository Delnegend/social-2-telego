package newsocial

import "regexp"

type MediaType string

var (
	htmlUrlRgx            = regexp.MustCompile(`<a href="([^"]+)"[^>]*>([^<]+)</a>`)
	htmlUrlPlaceholderRgx = regexp.MustCompile(`HLSTART ([^ ]+) HLSPLIT ([^ ]+) HLEND`)
	htmlParaRgx           = regexp.MustCompile(`<p>([^<]+)</p>`)
)

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVideo MediaType = "video"
)

type ScrapedMedia struct {
	MediaType MediaType
	MediaUrl  string
}

type ScrapeResult struct {
	MarkdownContent *string
	Username        *string
	ProfileURL      *string
	EscapeChar      *string
	Media           []ScrapedMedia
}
