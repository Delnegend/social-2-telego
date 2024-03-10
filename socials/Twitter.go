package socials

import (
	"errors"
	"regexp"
	"strings"
)

type Twitter struct{}

func (t Twitter) GetName(url string) (string, error) {
	return t.GetUsername(url)
}
func (t Twitter) GetUsername(url string) (string, error) {
	slice := strings.Split(url, "/")
	if len(slice) < 4 {
		return "", errors.New("invalid URL")
	}
	return slice[3], nil
}
func (t Twitter) IntoTeleEmbedLink(url string) (string, error) {
	url = strings.Replace(url, "twitter.com", "i.fxtwitter.com", 1)
	url = strings.Replace(url, "x.com", "i.fxtwitter.com", 1)
	if strings.Contains(url, "?") {
		url = strings.Split(url, "?")[0]
	}
	return url, nil
}

func (t Twitter) IsValidURL(url string) error {
	pattern := `https:\/\/((twitter)|x).com\/([\w_]{1,15})\/status\/\d+`
	if !regexp.MustCompile(pattern).MatchString(url) {
		return errors.New("invalid URL")
	}
	return nil
}
