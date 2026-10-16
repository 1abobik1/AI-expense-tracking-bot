package server

import (
	"database/sql"
	"log/slog"
	"net/http"

	"expensebot/config"
	"expensebot/internal/agent/clients/llm"
	"expensebot/internal/app/clock"
	"expensebot/internal/app/telegram"
	"expensebot/internal/members"

	agentBot "expensebot/internal/agent/transport/bot"
	agentUsecase "expensebot/internal/agent/usecase"
	appHttp "expensebot/internal/app/http"
	authHttp "expensebot/internal/auth/transport/http"
	authUsecase "expensebot/internal/auth/usecase"
	expenseRepo "expensebot/internal/expenses/repository/sqlite/expense"
	adminExpensesHttp "expensebot/internal/expenses/transport/http/admin"
	expensesUsecase "expensebot/internal/expenses/usecase"
	counterRepo "expensebot/internal/limits/repository/sqlite/counter"
	limitsUsecase "expensebot/internal/limits/usecase"
	memberRepo "expensebot/internal/members/repository/sqlite/member"
	adminMembersHttp "expensebot/internal/members/transport/http/admin"
	membersUsecase "expensebot/internal/members/usecase"
)

type Deps struct {
	DB         *sql.DB
	HTTPClient appHttp.Doer // *http.Client на сервере, fetch в Worker
	Log        *slog.Logger
}

type App struct {
	Telegram *telegram.Client  // nil при TELEGRAM_MODE=off
	Bot      *agentBot.Handler // nil при TELEGRAM_MODE=off
	HTTP     http.Handler
}

func New(cfg *config.Config, d Deps) (*App, error) {
	model, err := llm.New(mapLLMConfig(cfg.LLM), d.HTTPClient)
	if err != nil {
		return nil, err
	}

	a := &App{}
	clk := clock.New(cfg.Timezone.Location)

	var notifier members.Notifier // nil-интерфейс, а не nil-указатель, если бот выключен
	if cfg.Telegram.Mode != config.TelegramOff {
		a.Telegram = telegram.New(cfg.Telegram.Token, cfg.Telegram.APIBase, d.HTTPClient)
		notifier = a.Telegram
	}

	membersRepo := memberRepo.NewRepo(d.DB)
	membersUC := membersUsecase.NewMembersUC(membersRepo, notifier)
	expensesUC := expensesUsecase.NewExpensesUC(expenseRepo.NewRepo(d.DB), membersRepo, clk)
	limiter := limitsUsecase.NewLimiter(counterRepo.NewRepo(d.DB), clk)

	mux := appHttp.NewRouter()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })

	auth := authHttp.NewHandlers(authUsecase.NewAuthUC(cfg.Admin.Password), limiter, cfg.Admin.LoginAttemptsPerHour, d.Log)
	auth.Register(mux)
	adminExpensesHttp.NewHandlers(expensesUC, d.Log).Register(mux, auth.RequireAdmin)
	adminMembersHttp.NewHandlers(membersUC, d.Log).Register(mux, auth.RequireAdmin)

	if a.Telegram != nil {
		agentUC := agentUsecase.NewAgentUC(expensesUC, model, limiter, mapLimits(cfg.Limits), d.Log)

		a.Bot = agentBot.NewHandler(agentUC, membersUC, a.Telegram, d.Log)
		if cfg.Telegram.Mode == config.TelegramWebhook {
			mux.Handle("POST /telegram/webhook", agentBot.WebhookHandler(a.Bot, cfg.Telegram.WebhookSecret))
		}
	}

	a.HTTP = appHttp.Recovery(mux, d.Log)

	return a, nil
}

// NewPoller — long polling Telegram для сервера.
func NewPoller(a *App, log *slog.Logger) *agentBot.Poller {
	return agentBot.NewPoller(a.Telegram, a.Bot, log)
}
