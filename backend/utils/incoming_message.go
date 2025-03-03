package utils

type IncomingMessage struct {
	MessageID int `json:"message_id"`
	From      struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	} `json:"from"`
	Chat struct {
		ID int `json:"id"`
	} `json:"chat"`
	Text string `json:"text"`
}
