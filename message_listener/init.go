package message_listener

import (
	"log/slog"
	"net/http"
	"social-2-telego/utils"
)

type MessageListener struct {
	appState     *utils.AppState
	offset       int
	webhookToken string
}

// Either launch a http server for webhook
// or poll updates from Telegram's servers
func InitMessageListener(appState *utils.AppState) {
	ml := &MessageListener{
		offset:   0,
		appState: appState,
	}
	switch ml.appState.GetUseWebhook() {
	case true:
		ml.setWebhook()
		http.HandleFunc("POST /webhook", ml.handleWebhookRequest)
		slog.Info("listening on port " + ml.appState.GetPort())
		if err := http.ListenAndServe(":"+ml.appState.GetPort(), nil); err != nil {
			slog.Error("failed to start webhook server: ", "err", err)
		}
	case false:
		slog.Info("polling updates")
		ml.deleteWebhook()
		go ml.GetUpdates()
	}
}
