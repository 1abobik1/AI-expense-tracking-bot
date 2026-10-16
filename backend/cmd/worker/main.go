//go:build js && wasm

// Cloudflare Worker: D1, Workers AI и вебхук Telegram. Статику панели отдаёт Static Assets.
package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"sync"
	_ "time/tzdata"

	"expensebot/config"
	"expensebot/server"

	"github.com/syumai/workers-go"
	"github.com/syumai/workers-go/cloudflare"
	"github.com/syumai/workers-go/cloudflare/d1"
	"github.com/syumai/workers-go/cloudflare/fetch"
)

var log = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {
	workers.ServeNonBlock(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" { // без сборки приложения: проверка не должна зависеть от конфига
			_, _ = w.Write([]byte("ok"))
			return
		}

		application, err := app()
		if err != nil {
			log.Error("worker setup failed", "err", err)
			http.Error(w, "worker is misconfigured", http.StatusInternalServerError)

			return
		}

		application.HTTP.ServeHTTP(w, r)
	}))
	workers.Ready()
	// Экземпляр живёт между запросами изолята (deploy/cloudflare/worker.mjs).
	select {}
}

var (
	appMu  sync.Mutex
	appVal *server.App
)

// app собирается при первом запросе (env есть только в запросе); ошибка не кэшируется.
func app() (*server.App, error) {
	appMu.Lock()
	defer appMu.Unlock()

	if appVal != nil {
		return appVal, nil
	}

	a, err := build()
	if err != nil {
		return nil, err
	}

	appVal = a

	return a, nil
}

func build() (*server.App, error) {
	cfg, err := config.Load(cloudflare.Getenv)
	if err != nil {
		return nil, err
	}

	if cfg.Telegram.Mode == config.TelegramPolling {
		return nil, errors.New("TELEGRAM_MODE=polling is not possible in a Worker: use webhook")
	}

	connector, err := d1.OpenConnector("DB")
	if err != nil {
		return nil, err
	}

	return server.New(cfg, server.Deps{
		DB:         sql.OpenDB(connector),
		HTTPClient: fetchDoer{fetch.NewClient().HTTPClient(fetch.RedirectModeFollow).Transport},
		Log:        log,
	})
}

// fetchDoer минует http.Client, см. apphttp.Doer.
type fetchDoer struct{ rt http.RoundTripper }

func (d fetchDoer) Do(r *http.Request) (*http.Response, error) { return d.rt.RoundTrip(r) }
