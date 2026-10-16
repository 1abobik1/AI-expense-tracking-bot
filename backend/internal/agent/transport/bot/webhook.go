package bot

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"expensebot/internal/app/telegram"
)

// WebhookHandler всегда отвечает 200: иначе Telegram повторит апдейт и трата запишется дважды.
func WebhookHandler(h *Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if secret == "" || subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var u telegram.Update
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&u); err == nil {
			h.HandleUpdate(r.Context(), u)
		}

		w.WriteHeader(http.StatusOK)
	})
}
