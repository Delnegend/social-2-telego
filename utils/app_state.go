package utils

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type AppState struct {
	botToken       string
	targetChannel  string
	numWorkers     int
	allowedAccount []string
	webhookDomain  string
	port           string // for polling, ignore if use webhook

	MsgQueue chan IncomingMessage
}

func InitAppState() *AppState {
	as := &AppState{}
	as.allowedAccount = make([]string, 0)
	as.MsgQueue = make(chan IncomingMessage)

	// parsing envs

	numWorkers := os.Getenv("NUM_WORKERS")
	if numWorkers == "" {
		as.numWorkers = 5
	} else {
		num, err := strconv.Atoi(numWorkers)
		if err != nil {
			slog.Warn("NUM_WORKERS must be an integer, using default value 5")
		}
		as.numWorkers = num
	}

	as.botToken = os.Getenv("BOT_TOKEN")
	if as.botToken == "" {
		log.Fatal("BOT_TOKEN must be set")
	}

	as.targetChannel = os.Getenv("CHANNEL")
	if as.targetChannel == "" {
		slog.Info("CHANNEL is not set, messages will be echoed back to the user.")
	}

	allowedAccount := os.Getenv("WHITELISTED")
	if allowedAccount == "" {
		slog.Info("WHITELISTED is not set, everyone can use the bot to send messages to your channel! You can add multiple usernames, @ is optional, each separated by a space.")
	} else {
		slice := strings.Split(allowedAccount, " ")
		for _, account := range slice {
			account = strings.TrimPrefix(account, "@")
			as.allowedAccount = append(as.allowedAccount, account)
		}
	}

	as.webhookDomain = os.Getenv("DOMAIN")
	if as.webhookDomain == "" {
		slog.Warn("DOMAIN is not set, webhook will not be enabled")
	} else {
		port := os.Getenv("PORT")
		portInt, err := strconv.Atoi(port)
		if err != nil {
			slog.Warn("PORT is not set, defaulting to 8080")
			portInt = 8080
		}
		if portInt < 0 || portInt > 65535 {
			slog.Warn("PORT must be between 0 and 65535, defaulting to 8080")
			portInt = 8080
		}
		as.port = fmt.Sprintf("%d", portInt)
		slog.Info("listening on port", "port", as.port)
	}

	return as
}

func (as *AppState) BotToken() string {
	return as.botToken
}
func (as *AppState) TargetChannel() string {
	return as.targetChannel
}
func (as *AppState) NumWorkers() int {
	return as.numWorkers
}
func (as *AppState) IsAccountAllowed(account string) bool {
	if len(as.allowedAccount) == 0 {
		return true
	}
	for _, acc := range as.allowedAccount {
		if acc == account {
			return true
		}
	}
	return false
}
func (as *AppState) WebhookDomain() string {
	return as.webhookDomain
}
func (as *AppState) Port() string {
	return as.port
}
