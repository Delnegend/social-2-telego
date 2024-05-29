package telegram

import (
	"errors"
	"fmt"
	"log/slog"
	"social-2-telego/social"
	"social-2-telego/utils"
	"strings"
)

// Split the message by comma, rm spaces and empty strings
func splitAndCleanup(message string) ([]string, error) {
	slice := strings.Split(message, ",")
	if len(slice) > 3 {
		return nil, errors.New("too many arguments")
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
			newSlice = append(newSlice, utils.EscapeTelegramChar(item))
		}
	}
	return "ESCAPE_CHAR[" + strings.Join(newSlice, " ") + "ESCAPE_CHAR]"
}

// (name, username)
func processArtistNameAndUsername(rawItems []string, social social.Social) (string, string, error) {
	var stringToParse string
	for _, item := range rawItems {
		if strings.HasPrefix(item, "@") {
			stringToParse = item
			break
		}
	}

	slice := strings.Fields(stringToParse)
	var name, username string

	if len(slice) == 0 || stringToParse == "" {
		if result, err := social.GetName(); err != nil {
			return "", "", err
		} else {
			name = result
		}
		if result, err := social.GetUsername(); err != nil {
			return "", "", err
		} else {
			username = result
		}
		return name, username, nil
	}

	if len(slice) == 1 {
		username = slice[0][1:]
		firstLetter := strings.ToUpper(string(username[0]))
		name = firstLetter + username[1:]
		return name, username, nil
	}

	if len(slice) == 2 {
		name = slice[0][1:]
		username = slice[1]
		return name, username, nil
	}

	return "", "", errors.New("too many arguments for artist name and username")
}

func ComposeMessage(appState *utils.AppState, message string, social social.Social) (string, error) {
	// Split the message by comma, rm spaces and empty strings
	slice, err := splitAndCleanup(message)
	if err != nil {
		return "", err
	}

	urlFound := false
	var hashtags, name, username string

	// Finding the post URL and hashtags. Later finding the artist name and
	// username requires the post URL to be set, so we need to do 2 separate
	// loops.
	for _, item := range slice {
		if !urlFound && strings.HasPrefix(item, "https://") {
			if err := social.SetURL(item); err != nil {
				slog.Warn("can't set URL", "error", err)
			} else {
				urlFound = true
			}
		}
		if hashtags == "" && strings.HasPrefix(item, "#") {
			hashtags = processHashtags(item)
		}
	}

	name, username, err = processArtistNameAndUsername(slice, social)
	if err != nil {
		return "", err
	}

	var artistProfileURL string
	if appState.GetArtistDBDomain() == "" {
		artistProfileURL, err = social.GetProfileURL(username)
		if err != nil {
			return "", fmt.Errorf("failed to get profile URL: %w", err)
		}
	} else {
		artistProfileURL = strings.Replace(appState.GetArtistDBDomain(), "{username}", username, 1)
	}

	origUrl, err := social.GetURL()
	if err != nil {
		return "", err
	}

	content, err := social.GetContent()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s\n[Post](%s) ESCAPE_CHAR| [%s](%s) %s",
		content,
		origUrl,
		utils.EscapeTelegramChar(name),
		artistProfileURL,
		hashtags,
	), nil
}
