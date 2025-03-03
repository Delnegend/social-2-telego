package telegram

import (
	"fmt"
	"net/url"
	newsocial "social-2-telego/backend/new_social"
	"social-2-telego/backend/utils"
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

func composeMessage(chatID string,
	postURL string,
	scrapeResult *newsocial.ScrapeResult,
) (url.Values, SendType, error) {
	content := func() string {
		content := scrapeResult.MarkdownContent

		if *content != "" {
			*content = fmt.Sprintf(">%s\n", strings.Join(strings.Split(*content, "\n"), "\n>"))
		}

		return fmt.Sprintf("%s[Post](%s) %s| [%s](%s)",
			*content,
			utils.EscapeSpecialChars(postURL, *scrapeResult.EscapeChar),
			*scrapeResult.EscapeChar,
			utils.EscapeSpecialChars(*scrapeResult.Username, *scrapeResult.EscapeChar),
			utils.EscapeSpecialChars(*scrapeResult.ProfileURL, *scrapeResult.EscapeChar),
		)
	}()

	data := url.Values{
		"chat_id":              {chatID},
		"parse_mode":           {"MarkdownV2"},
		"disable_notification": {"true"},
	}

	switch len(scrapeResult.Media) {
	case 0:
		data.Add("text", content)
		return data, SendTypeMessage, nil
	case 1:
		var endPoint SendType

		switch scrapeResult.Media[0].MediaType {
		case newsocial.MediaTypePhoto:
			endPoint = SendTypePhoto
		case newsocial.MediaTypeVideo:
			endPoint = SendTypeVideo
		default:
			return data, "", fmt.Errorf("invalid media type")
		}
		data.Add(string(scrapeResult.Media[0].MediaType), scrapeResult.Media[0].MediaUrl)
		data.Add("caption", content)

		return data, endPoint, nil
	default:
		result := make([]string, 0)
		// There's no "text", must add "caption" for the first media instead
		result = append(result,
			fmt.Sprintf(`{"type":"%s","media":"%s","caption":"%s","parse_mode":"MarkdownV2"}`,
				scrapeResult.Media[0].MediaType,
				scrapeResult.Media[0].MediaUrl,
				content))

		// Add the rest of the media to the result
		for _, media := range scrapeResult.Media[1:] {
			result = append(result,
				fmt.Sprintf(`{"type":"%s","media":"%s"}`,
					media.MediaType,
					media.MediaUrl))
		}

		data.Add("media", "["+strings.Join(result, ",")+"]")
		return data, SendTypeMediaGroup, nil
	}
}
