package pkg

import (
	"errors"
	"fmt"
	"os"
	"social-2-telego/socials"
	"strings"
)

const WRONG_MSG_FORMAT = "wrong message format, /help for more info"

func escapeSpecialChar(s *string) {
	for _, char := range []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"} {
		*s = strings.Replace(*s, char, "\\"+char, -1)
	}
}

// Split the message by comma, rm spaces and empty strings
func splitAndCleanup(message string) ([]string, error) {
	slice := strings.Split(message, ",")
	if len(slice) > 3 {
		return nil, errors.New(WRONG_MSG_FORMAT)
	}
	var newSlice []string
	for _, item := range slice {
		item = strings.TrimSpace(item)
		if item != "" {
			newSlice = append(newSlice, item)
		}
	}
	return newSlice, nil
}

// "  #foo  #bar #baz  baa" -> "#foo #bar #baz"
func processHashtags(hashtags string) string {
	slice := strings.Fields(hashtags)
	var newSlice []string
	for _, item := range slice {
		if strings.HasPrefix(item, "#") {
			escapeSpecialChar(&item)
			newSlice = append(newSlice, item)
		}
	}
	return "\\[" + strings.Join(newSlice, " ") + "\\]"
}

// (name, username)
func processArtistNameAndUsername(nameAndUsername string) (string, string) {
	slice := strings.Fields(nameAndUsername)
	var name, username string

	for _, item := range slice {
		if strings.HasPrefix(item, "@") {
			username = item
			break
		}
	}

	newSlice := []string{}
	for _, item := range slice {
		if item != username {
			newSlice = append(newSlice, item)
		}
	}
	name = strings.Join(newSlice, " ")
	if name == "" {
		name = username
	}

	return name, strings.Replace(username, "@", "", 1)
}

func ComposeMessage(message string, social socials.Social) (string, error) {
	slice, err := splitAndCleanup(message)
	if err != nil {
		return "", err
	}

	var postURL, hashtags, name, username string
	for _, item := range slice {
		if postURL == "" && strings.HasPrefix(item, "https://") {
			postURL = item
		}
		if hashtags == "" && strings.HasPrefix(item, "#") {
			hashtags = processHashtags(item)
		}
		if username == "" && strings.HasPrefix(item, "@") {
			name, username = processArtistNameAndUsername(item)
		}
	}

	if err := social.IsValidURL(postURL); err != nil {
		return "", err
	}
	if name == "" {
		if result, err := social.GetName(postURL); err != nil {
			return "", err
		} else {
			name = result
		}
	}
	if username == "" {
		if result, err := social.GetUsername(postURL); err != nil {
			return "", err
		} else {
			username = result
		}
	}
	escapeSpecialChar(&name)

	artistLink := strings.Replace(os.Getenv("ARTIST_DB"), "{username}", username, 1)
	postURL, err = social.IntoTeleEmbedLink(postURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"[🖌️](%s) [%s](%s) %s",
		postURL,
		name,
		artistLink,
		hashtags,
	), nil
}
