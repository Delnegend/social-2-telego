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
	useWebhook             bool
	port                   string
	webhookDomain          string
	webhookSecret          string
	retrySetWebhookAttempt int

	getUpdatesInterval time.Duration

	botToken       string
	artistDBDomain string
	allowedUsers   map[string]interface{}

	targetChannel string
	numWorker     int
	faCookieA     string
	faCookieB     string

	MsgQueue chan IncomingMessage
}

// Create a new AppState instance
func NewAppState() *AppState {
	return &AppState{
		useWebhook: func() bool {
			useWebhook := os.Getenv("USE_WEBHOOK")
			return strings.ToLower(useWebhook) == "true"
		}(),
		port: func() string {
			port := os.Getenv("PORT")
			if port == "" {
				slog.Warn("PORT is not set, defaulting to 8080")
				return "8080"
			}
			portInt, err := strconv.Atoi(port)
			if err != nil {
				slog.Warn("PORT must be an integer, defaulting to 8080")
				return "8080"
			}
			if portInt < 0 || portInt > 65535 {
				slog.Warn("PORT must be between 0 and 65535, defaulting to 8080")
				return "8080"
			}
			return fmt.Sprintf("%d", portInt)
		}(),
		webhookDomain: func() string {
			webhookDomain := os.Getenv("WEBHOOK_DOMAIN")
			if _, err := url.ParseRequestURI(webhookDomain); err != nil {
				slog.Warn("WEBHOOK_DOMAIN is not a valid URL, webhook will not be enabled")
				return ""
			}
			return webhookDomain
		}(),
		webhookSecret: func() string {
			webhookSecret := os.Getenv("WEBHOOK_SECRET")
			if match, _ := regexp.MatchString(`[^a-zA-Z0-9_-]`, webhookSecret); match {
				slog.Warn("WEBHOOK_SECRET must only contain alphanumeric characters, underscores, and hyphens, webhook will not be enabled")
				return ""
			}
			return webhookSecret
		}(),
		retrySetWebhookAttempt: func() int {
			retrySetWebhookAttempt := os.Getenv("RETRY_SET_WEBHOOK_ATTEMPT")
			retrySetWebhookAttemptInt, err := strconv.Atoi(retrySetWebhookAttempt)
			if err != nil {
				slog.Warn("RETRY_SET_WEBHOOK_ATTEMPT must be an integer, defaulting to 3")
				return 3
			}
			return retrySetWebhookAttemptInt
		}(),

		getUpdatesInterval: func() time.Duration {
			getUpdatesInterval := os.Getenv("GET_UPDATES_INTERVAL")
			getUpdatesIntervalDur, err := time.ParseDuration(getUpdatesInterval)
			if err != nil {
				slog.Warn("GET_UPDATES_INTERVAL is not a valid duration, defaulting to 1s")
				return time.Second
			}
			return getUpdatesIntervalDur
		}(),

		botToken: func() string {
			botToken := os.Getenv("BOT_TOKEN")
			if botToken == "" {
				log.Fatal("BOT_TOKEN must be set")
			}
			return botToken
		}(),
		artistDBDomain: func() string {
			artistDBDomain := os.Getenv("ARTIST_DB_DOMAIN")
			if _, err := url.ParseRequestURI(artistDBDomain); err != nil {
				slog.Error("ARTIST_DB_DOMAIN is not a valid URL")
				os.Exit(1)
			}

			if !strings.Contains(artistDBDomain, "{username}") {
				slog.Error("ARTIST_DB_DOMAIN must contain {username}")
				os.Exit(1)
			}

			return artistDBDomain
		}(),

		allowedUsers: func() map[string]interface{} {
			allowedAccounts := os.Getenv("ALLOWED_USERS")
			if allowedAccounts == "" {
				slog.Warn("ALLOWED_USERS is not set, everyone can use the bot to send messages to your channel! You can add multiple usernames, @ is optional, each separated by a space")
				return make(map[string]interface{})
			}
			slice := strings.Split(allowedAccounts, ",")
			allowedAccountsMap := make(map[string]interface{})
			for _, account := range slice {
				account = strings.TrimPrefix(account, "@")
				allowedAccountsMap[account] = struct{}{}
			}
			return allowedAccountsMap
		}(),

		targetChannel: func() string {
			targetChannel := os.Getenv("TARGET_CHANNEL")
			if targetChannel == "" {
				slog.Info("TARGET_CHANNEL is not set, messages will be echoed back to the user")
				return ""
			}
			return targetChannel
		}(),
		numWorker: func() int {
			numWorkers := os.Getenv("NUM_WORKERS")
			if numWorkers == "" {
				slog.Warn("NUM_WORKERS is not set, defaulting to 5")
				return 5
			}
			numWorkersInt, err := strconv.Atoi(numWorkers)
			if err != nil {
				slog.Warn("NUM_WORKERS must be an integer, defaulting to 5")
				return 5
			}
			return numWorkersInt
		}(),
		faCookieA: func() string {
			faCookieA := os.Getenv("FA_COOKIE_A")
			if faCookieA == "" {
				slog.Warn("FA_COOKIE_A is not set, scraping FuraAffinity will not be possible")
				return ""
			}
			return faCookieA
		}(),
		faCookieB: func() string {
			faCookieB := os.Getenv("FA_COOKIE_B")
			if faCookieB == "" {
				slog.Warn("FA_COOKIE_B is not set, scraping FuraAffinity will not be possible")
				return ""
			}
			return faCookieB
		}(),

		MsgQueue: make(chan IncomingMessage),
	}
}

// Get whether the app is using the webhook
func (c *AppState) GetUseWebhook() bool {
	return c.useWebhook
}

// Get the port for the app to listen on if using the webhook
func (c *AppState) GetPort() string {
	return c.port
}

// Get the domain for the webhook
func (c *AppState) GetWebhookDomain() string {
	return c.webhookDomain
}

// Get the secret for the webhook
func (c *AppState) GetWebhookSecret() string {
	return c.webhookSecret
}

// Get the number of times to retry setting the webhook
func (c *AppState) GetRetrySetWebhookAttempt() int {
	return c.retrySetWebhookAttempt
}

// Get the interval for getting updates
func (c *AppState) GetGetUpdatesInterval() time.Duration {
	return c.getUpdatesInterval
}

// Get the bot token
func (c *AppState) GetBotToken() string {
	return c.botToken
}

// Get the artistDB domain
func (c *AppState) GetArtistDBDomain() string {
	return c.artistDBDomain
}

// Check if an account is authorized
func (c *AppState) IsAuthorized(account string) bool {
	if len(c.allowedUsers) == 0 {
		return true
	}
	account = strings.TrimPrefix(account, "@")
	_, ok := c.allowedUsers[account]
	return ok
}

// Get the target channel
func (c *AppState) GetTargetChannel() string {
	return c.targetChannel
}

// Get the number of workers
func (c *AppState) GetNumWorkers() int {
	return c.numWorker
}

// Get the FuraAffinity cookie A
func (c *AppState) GetFaCookieA() string {
	return c.faCookieA
}

// Get the FuraAffinity cookie B
func (c *AppState) GetFaCookieB() string {
	return c.faCookieB
}
