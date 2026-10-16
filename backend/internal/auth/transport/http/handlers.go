package http

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"expensebot/internal/auth"
	"expensebot/internal/limits"

	apphttp "expensebot/internal/app/http"
)

type Handlers struct {
	authUC        auth.AuthUseCase
	limiter       limits.Limiter
	loginsPerHour int
	log           *slog.Logger
}

func NewHandlers(authUC auth.AuthUseCase, limiter limits.Limiter, loginsPerHour int, log *slog.Logger) *Handlers {
	return &Handlers{authUC: authUC, limiter: limiter, loginsPerHour: loginsPerHour, log: log}
}

func (h *Handlers) Register(mux *apphttp.Router) {
	mux.HandleFunc("POST /api/login", h.login)
}

// RequireAdmin пропускает только запросы с действующим токеном панели.
func (h *Handlers) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || !h.authUC.VerifyToken(token) {
			apphttp.WriteError(w, http.StatusUnauthorized, "нужен вход")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	allowed, err := h.limiter.Allow(r.Context(), "login:"+clientIP(r), h.loginsPerHour, limits.PerHour)
	if err != nil {
		apphttp.Fail(w, h.log, err)
		return
	}

	if !allowed {
		apphttp.WriteError(w, http.StatusTooManyRequests, "слишком много попыток, попробуйте через час")
		return
	}

	var req loginRequest
	if !apphttp.Decode(w, r, &req) {
		return
	}

	if !h.authUC.CheckPassword(req.Password) {
		apphttp.WriteError(w, http.StatusUnauthorized, "неверный пароль")
		return
	}

	token, expires := h.authUC.IssueToken()
	apphttp.WriteJSON(w, http.StatusOK, loginResponse{Token: token, ExpiresAt: expires})
}

// clientIP доверяет заголовкам: их ставят Cloudflare и nginx, бэкенд наружу не публикуется.
func clientIP(r *http.Request) string {
	for _, h := range []string{"CF-Connecting-IP", "X-Real-IP"} {
		if ip := r.Header.Get(h); ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
