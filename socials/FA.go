package socials

import (
	"errors"
	"regexp"
)

type FA struct{}

func (f FA) GetName(url string) (string, error) {
	return "", nil
}
func (f FA) GetUsername(url string) (string, error) {
	return "", nil
}
func (f FA) IntoTeleEmbedLink(url string) (string, error) {
	return url, nil
}
func (t FA) IsValidURL(url string) error {
	pattern := `https:\/\/www\.furaffinity\.net\/view\/\d+`
	if !regexp.MustCompile(pattern).MatchString(url) {
		return errors.New("invalid URL")
	}

	return nil
}
