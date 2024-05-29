package utils

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AppState struct {
	port string

	useWebhook             bool
	webhookDomain          string
	webhookSecret          string
	retrySetWebhookAttempt int

	getUpdatesInterval time.Duration

	botToken        string
	artistDBDomain  string
	allowedAccounts map[string]interface{}
	targetChannel   string
	numWorker       int

	MsgQueue chan IncomingMessage
}

func NewAppState() *AppState {
	appState := &AppState{
		allowedAccounts: make(map[string]interface{}),
		MsgQueue:        make(chan IncomingMessage),
	}

	appState.port = os.Getenv("PORT")
	if appState.port == "" {
		slog.Warn("PORT is not set, defaulting to 8080")
		appState.port = "8080"
	}
	portInt, err := strconv.Atoi(appState.port)
	if err != nil {
		slog.Warn("PORT must be an integer, defaulting to 8080")
		appState.port = "8080"
	}
	if portInt < 0 || portInt > 65535 {
		slog.Warn("PORT must be between 0 and 65535, defaulting to 8080")
		appState.port = "8080"
	}
	appState.port = fmt.Sprintf("%d", portInt)

	appState.useWebhook = strings.ToLower(os.Getenv("USE_WEBHOOK")) == "true"

	appState.webhookDomain = os.Getenv("WEBHOOK_DOMAIN")
	if _, err := url.ParseRequestURI(appState.webhookDomain); err != nil {
		slog.Warn("WEBHOOK_DOMAIN is not a valid URL, webhook will not be enabled")
		appState.useWebhook = false
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if match, _ := regexp.MatchString(`[^a-zA-Z0-9_-]`, webhookSecret); match {
		slog.Warn("WEBHOOK_SECRET must only contain alphanumeric characters, underscores, and hyphens, webhook will not be enabled")
		appState.useWebhook = false
	} else {
		appState.webhookSecret = webhookSecret
	}

	appState.retrySetWebhookAttempt = 3
	if appState.useWebhook {
		retrySetWebhookAttempt := os.Getenv("RETRY_SET_WEBHOOK_ATTEMPT")
		retrySetWebhookAttemptInt, err := strconv.Atoi(retrySetWebhookAttempt)
		if err != nil {
			slog.Warn("RETRY_SET_WEBHOOK_ATTEMPT must be an integer, defaulting to 3")
		}
		appState.retrySetWebhookAttempt = retrySetWebhookAttemptInt

	}

	appState.getUpdatesInterval = time.Second
	if !appState.useWebhook {
		getUpdatesInterval := os.Getenv("GET_UPDATES_INTERVAL")
		getUpdatesIntervalDur, err := time.ParseDuration(getUpdatesInterval)
		if err != nil {
			slog.Warn("GET_UPDATES_INTERVAL is not a valid duration, defaulting to 1s")
		}
		appState.getUpdatesInterval = getUpdatesIntervalDur
	}

	appState.botToken = os.Getenv("BOT_TOKEN")
	if appState.botToken == "" {
		log.Fatal("BOT_TOKEN must be set")
	}

	appState.artistDBDomain = os.Getenv("ARTIST_DB_DOMAIN")
	_, err = url.ParseRequestURI(appState.artistDBDomain)
	if err != nil {
		slog.Warn("ARTIST_DB_DOMAIN is not a valid URL, using the same social media as the given URL")
	}

	allowedAccounts := os.Getenv("ALLOWED_USERS")
	if allowedAccounts == "" {
		slog.Warn("ALLOWED_USERS is not set, everyone can use the bot to send messages to your channel! You can add multiple usernames, @ is optional, each separated by a space")
	} else {
		slice := strings.Split(allowedAccounts, ",")
		for _, account := range slice {
			account = strings.TrimPrefix(account, "@")
			appState.allowedAccounts[account] = struct{}{}
		}
	}

	appState.targetChannel = os.Getenv("TARGET_CHANNEL")
	if appState.targetChannel == "" {
		slog.Info("TARGET_CHANNEL is not set, messages will be echoed back to the user")
	}

	numWorkers := os.Getenv("NUM_WORKERS")
	if numWorkers == "" {
		slog.Warn("NUM_WORKERS is not set, defaulting to 5")
		appState.numWorker = 5
	} else {
		numWorkersInt, err := strconv.Atoi(numWorkers)
		if err != nil {
			slog.Warn("NUM_WORKERS must be an integer, defaulting to 5")
			appState.numWorker = 5
		} else {
			appState.numWorker = numWorkersInt
		}
	}

	return appState
}

func (c *AppState) GetPort() string {
	return c.port
}
func (c *AppState) GetUseWebhook() bool {
	return c.useWebhook
}
func (c *AppState) GetWebhookDomain() string {
	return c.webhookDomain
}
func (c *AppState) GetWebhookSecret() string {
	return c.webhookSecret
}
func (c *AppState) GetRetrySetWebhookAttempt() int {
	return c.retrySetWebhookAttempt
}
func (c *AppState) GetGetUpdatesInterval() time.Duration {
	return c.getUpdatesInterval
}
func (c *AppState) GetBotToken() string {
	return c.botToken
}
func (c *AppState) GetArtistDBDomain() string {
	return c.artistDBDomain
}
func (c *AppState) IsAccountAllowed(account string) bool {
	if len(c.allowedAccounts) == 0 {
		return true
	}
	account = strings.TrimPrefix(account, "@")
	_, ok := c.allowedAccounts[account]
	return ok
}
func (c *AppState) GetTargetChannel() string {
	return c.targetChannel
}
func (c *AppState) GetNumWorkers() int {
	return c.numWorker
}
