package telegram

import (
	"fmt"
	"net/url"
	"social-2-telego/social"
	"social-2-telego/utils"
	"strings"
)

type (
	SendType string
	Caption  string
)

const (
	SendTypeMessage    SendType = "sendMessage"
	SendTypePhoto      SendType = "sendPhoto"
	SendTypeVideo      SendType = "sendVideo"
	SendTypeMediaGroup SendType = "sendMediaGroup"
)

type TelegramMessage struct {
	content     func(string) string
	postURL     string
	username    string
	displayName string
	hashtags    []string
	media       []social.ScrapedMedia
}

// Set the content from the raw HTML to the message
func (tmc *TelegramMessage) SetContent(content func(string) string) *TelegramMessage {
	tmc.content = content
	return tmc
}

// Set the hashtags to the message, separated by spaces
func (tmc *TelegramMessage) SetHashtags(hashtags string) *TelegramMessage {
	if hashtags == "" {
		return tmc
	}
	for _, item := range strings.Fields(hashtags) {
		hashtag := strings.TrimSpace(strings.TrimPrefix(item, "#"))
		if hashtag != "" {
			tmc.hashtags = append(tmc.hashtags, hashtag)
		}
	}
	return tmc
}

// Set the artist's username and optional display name to the message. Example:
//
//	"@lorem ipsum" -> username = lorem, displayName = ipsum
//	"lorem ipsum" -> username = lorem, displayName = ipsum
//	"lorem @ipsum" -> username = ipsum, displayName = lorem
//	"lorem" || "@lorem" -> username = lorem, displayName = lorem
func (tmc *TelegramMessage) SetArtistNameAndUsername(s string) *TelegramMessage {
	slice := strings.Fields(s)
	switch len(slice) {
	case 0:
		return tmc
	case 1:
		temp := strings.TrimPrefix(slice[0], "@")
		tmc.username = temp
		tmc.displayName = temp
		return tmc
	case 2:
		switch {
		case strings.HasPrefix(slice[0], "@"):
			tmc.username = slice[0][1:]
			tmc.displayName = slice[1]
		case strings.HasPrefix(slice[1], "@"):
			tmc.username = slice[1][1:]
			tmc.displayName = slice[0]
		default:
			tmc.username = slice[0]
			tmc.displayName = slice[1]
		}
		return tmc
	default:
		return tmc
	}
}

// Set the media slice to the message
func (tmc *TelegramMessage) SetMedia(media []social.ScrapedMedia) *TelegramMessage {
	tmc.media = media
	return tmc
}

// Set the post URL to the message
func (tmc *TelegramMessage) SetPostURL(url string) *TelegramMessage {
	tmc.postURL = url
	return tmc
}

// Serialize the content (aka caption, or the message) to a string
func (tmc *TelegramMessage) serializeContent(doubleEscape bool) (string, error) {
	switch {
	case tmc.postURL == "":
		return "", fmt.Errorf("TelegramMsgComposer.Serialize: postURL is empty")
	case tmc.username == "" || tmc.displayName == "":
		return "", fmt.Errorf("TelegramMsgComposer.Serialize: username is empty")
	}

	escapeChar := `\`
	if doubleEscape {
		escapeChar = `\\`
	}

	content := tmc.content(escapeChar)
	if content != "" {
		content = fmt.Sprintf(">%s\n", strings.Join(strings.Split(content, "\n"), "\n>"))
	}

	hashtags := func() string {
		if len(tmc.hashtags) == 0 {
			return ""
		}
		hashtags := strings.TrimPrefix(strings.Join(tmc.hashtags, ` #`), ` `)
		return fmt.Sprintf(" %s[%s%s]", escapeChar, hashtags, escapeChar)
	}()

	return fmt.Sprintf("%s[Post](%s) %s| [%s](https://artistdb.delnegend.com/%s)%s",
		content,
		utils.EscapeSpecialChars(tmc.postURL, escapeChar),
		escapeChar,
		utils.EscapeSpecialChars(tmc.displayName, escapeChar),
		utils.EscapeSpecialChars(tmc.username, escapeChar),
		hashtags,
	), nil
}

// Return a fully processed data to be sent to Telegram
func (tmc *TelegramMessage) ToData(chatID string) (url.Values, SendType, error) {
	data := url.Values{
		"chat_id":              {chatID},
		"parse_mode":           {"MarkdownV2"},
		"disable_notification": {"true"},
	}

	switch len(tmc.media) {
	case 0:
		content, err := tmc.serializeContent(false)
		if err != nil {
			return data, "", err
		}
		data.Add("text", content)
		return data, SendTypeMessage, nil
	case 1:
		var endPoint SendType
		content, err := tmc.serializeContent(false)
		if err != nil {
			return data, "", err
		}

		switch tmc.media[0].MediaType {
		case social.MediaTypePhoto:
			endPoint = SendTypePhoto
		case social.MediaTypeVideo:
			endPoint = SendTypeVideo
		default:
			return data, "", fmt.Errorf("invalid media type")
		}
		data.Add(string(tmc.media[0].MediaType), tmc.media[0].MediaUrl)
		data.Add("caption", content)

		return data, endPoint, nil
	default:
		content, err := tmc.serializeContent(true)
		if err != nil {
			return data, "", err
		}

		result := make([]string, 0)
		// There's no "text", must add "caption" for the first media instead
		result = append(result,
			fmt.Sprintf(`{"type":"%s","media":"%s","caption":"%s","parse_mode":"MarkdownV2"}`,
				tmc.media[0].MediaType,
				tmc.media[0].MediaUrl,
				content))

		// Add the rest of the media to the result
		for _, media := range tmc.media[1:] {
			result = append(result,
				fmt.Sprintf(`{"type":"%s","media":"%s"}`,
					media.MediaType,
					media.MediaUrl))
		}

		data.Add("media", "["+strings.Join(result, ",")+"]")
		return data, SendTypeMediaGroup, nil
	}
}
