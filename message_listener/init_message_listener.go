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

func InitMessageListender(appState *utils.AppState) {
	ml := &MessageListener{
		offset:   0,
		appState: appState,
	}
		go ml.rotateWebhook()
	if ml.appState.GetUseWebhook() {
		http.HandleFunc("POST /webhook/{token}", ml.handleWebhookRequest)
		http.HandleFunc("POST /webhook", ml.handleWebhookRequest)

		slog.Info("listening on port " + ml.appState.GetPort())
		http.ListenAndServe(":"+ml.appState.GetPort(), nil)
	}
	slog.Info("polling updates")
	ml.deleteWebhook()
	go ml.GetUpdates()
}
