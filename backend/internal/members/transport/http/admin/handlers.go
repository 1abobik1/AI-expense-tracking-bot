package admin

import (
	"log/slog"
	"net/http"

	"expensebot/internal/members"

	apphttp "expensebot/internal/app/http"
)

type Handlers struct {
	membersUC members.MembersUseCase
	log       *slog.Logger
}

func NewHandlers(membersUC members.MembersUseCase, log *slog.Logger) *Handlers {
	return &Handlers{membersUC: membersUC, log: log}
}

func (h *Handlers) Register(mux *apphttp.Router, admin func(http.Handler) http.Handler) {
	mux.Handle("GET /api/members", admin(http.HandlerFunc(h.getMembers)))
	mux.Handle("PATCH /api/members/{id}", admin(http.HandlerFunc(h.updateMember)))
}
