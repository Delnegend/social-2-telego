package telegram

import (
	"time"
)

type Config struct {
	BotToken            string
	TargetChannel       *string
	WhitelistedAccounts []string

	// Don't care about these one
	MessageQueue          chan *Message
	UpdateID              int
	Offset                int
	WebhookToken          string
	WebhookRotateInterval time.Duration
}

type Message struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID           int    `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		Username     string `json:"username"`
		LanguageCode string `json:"language_code"`
	} `json:"from"`
	Chat struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	} `json:"chat"`
	Date     int    `json:"date"`
	Text     string `json:"text"`
	Entities []struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		Type   string `json:"type"`
	} `json:"entities"`
}
