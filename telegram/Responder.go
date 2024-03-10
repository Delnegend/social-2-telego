package telegram

import (
	"log"
	"log/slog"
	"os"
	"social-2-telego/pkg"
	"social-2-telego/socials"
	"strconv"
	"strings"
)

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// Continuously watching new messages from MessageQueue and respond to them
func (bot *Bot) Responder() {
mainLoop:
	for message := range bot.MessageQueue {
		slog.Debug("Received message", "from", message.From.Username, "text", message.Text)

		if whitelisted := os.Getenv("WHITELISTED"); whitelisted != "" {
			people := strings.Split(whitelisted, ",")
			sender := &message.From.Username
			if !contains(people, *sender) {
				slog.Warn("Unauthorized user", "username", *sender)
				continue mainLoop
			}
		}

		// find matching social media
		var matchedSocial socials.Social
	matchSocialLoop:
		for prefix, social := range bot.PrefixSocialMatch {
			if strings.HasPrefix(message.Text, prefix) {
				matchedSocial = social
				break matchSocialLoop
			}
		}
		if matchedSocial == nil {
			slog.Warn("No social media matched")
			continue mainLoop
		}

		// compose message
		outgoing, err := pkg.ComposeMessage(message.Text, matchedSocial)
		if err != nil {
			slog.Error("Failed to compose message: ", "err", err)
			bot.sendMessage(strconv.Itoa(message.Chat.ID), "Failed to compose message: "+err.Error())
			continue mainLoop
		}

		// send message
		if *bot.TargetChannel != "" {
			bot.sendMessage(*bot.TargetChannel, outgoing)
		} else {
			bot.sendMessage(strconv.Itoa(message.From.ID), outgoing)
		}
	}
	log.Fatal("Responder stopped for some reason, this should not happen")
}
