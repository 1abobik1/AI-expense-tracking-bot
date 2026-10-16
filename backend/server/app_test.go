package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"expensebot/config"
	"expensebot/internal/testutil"
	"expensebot/server"
)

const password = "family-secret"

// fakeTelegram записывает отправленные ботом сообщения.
type fakeTelegram struct {
	mu   sync.Mutex
	sent []string
}

func (f *fakeTelegram) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/sendMessage") {
		var m struct{ Text string }
		json.NewDecoder(r.Body).Decode(&m)
		f.mu.Lock()
		f.sent = append(f.sent, m.Text)
		f.mu.Unlock()
	}

	w.Write([]byte(`{"ok":true,"result":true}`))
}

func (f *fakeTelegram) last() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.sent) == 0 {
		return ""
	}

	return f.sent[len(f.sent)-1]
}

// fakeLLM — OpenAI-совместимый API, который всегда «понимает» трату на кофе.
func fakeLLM() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		intent := `{"action":"add","items":[{"amount":350,"category":"food","comment":"кофе"}]}`
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": intent}}}})
	})
}

type e2e struct {
	t      *testing.T
	server *httptest.Server
	tg     *fakeTelegram
	token  string
}

func setup(t *testing.T) *e2e {
	t.Helper()

	tg := &fakeTelegram{}
	tgSrv := httptest.NewServer(tg)
	llmSrv := httptest.NewServer(fakeLLM())

	t.Cleanup(tgSrv.Close)
	t.Cleanup(llmSrv.Close)

	env := map[string]string{
		"ADMIN_PASSWORD": password, "ADMIN_LOGIN_ATTEMPTS_PER_HOUR": "5",
		"LLM_PROVIDER": "openai", "LLM_MODEL": "test", "LLM_BASE_URL": llmSrv.URL,
		"TELEGRAM_MODE": "webhook", "TELEGRAM_BOT_TOKEN": "123:abc", "TELEGRAM_WEBHOOK_SECRET": "hook",
		"TELEGRAM_API_BASE": tgSrv.URL,
	}

	cfg, err := config.Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}

	app, err := server.New(cfg, server.Deps{
		DB: testutil.SQLite(t), HTTPClient: http.DefaultClient,
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(app.HTTP)
	t.Cleanup(srv.Close)

	return &e2e{t: t, server: srv, tg: tg}
}

func (e *e2e) do(method, path string, body any, headers ...string) (int, map[string]any) {
	e.t.Helper()

	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}

	req, _ := http.NewRequest(method, e.server.URL+path, &buf)
	if e.token != "" {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}

	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any

	raw, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(raw, &out) != nil && len(raw) > 0 {
		out = map[string]any{"raw": string(raw)}
	}

	return resp.StatusCode, out
}

func (e *e2e) telegramMessage(userID int64, text string, isBot bool) {
	e.t.Helper()

	update := map[string]any{"update_id": 1, "message": map[string]any{
		"message_id": 1, "text": text,
		"from": map[string]any{"id": userID, "is_bot": isBot, "first_name": "Алиса"},
		"chat": map[string]any{"id": userID, "type": "private"},
	}}
	if status, _ := e.do("POST", "/telegram/webhook", update, "X-Telegram-Bot-Api-Secret-Token", "hook"); status != 200 {
		e.t.Fatalf("webhook status %d", status)
	}
}

func TestBotWebhookToAdminPanel(t *testing.T) {
	e := setup(t)

	if status, _ := e.do("POST", "/telegram/webhook", map[string]any{}); status != http.StatusForbidden {
		t.Fatalf("webhook without secret: %d", status)
	}

	// Новый пользователь сначала ждёт одобрения.
	e.telegramMessage(1001, "кофе 350", false)

	if got := e.tg.last(); !strings.Contains(got, "Заявка") {
		t.Fatalf("pending reply = %q", got)
	}

	e.telegramMessage(1002, "кофе 350", true) // сообщения ботов игнорируются

	if status, _ := e.do("GET", "/api/expenses", nil); status != http.StatusUnauthorized {
		t.Fatalf("API without token: %d", status)
	}

	if status, _ := e.do("POST", "/api/login", map[string]string{"password": "wrong"}); status != http.StatusUnauthorized {
		t.Fatalf("wrong password: %d", status)
	}

	status, body := e.do("POST", "/api/login", map[string]string{"password": password})
	if status != 200 {
		t.Fatalf("login: %d %v", status, body)
	}

	e.token = body["token"].(string)

	if status, _ := e.do("POST", "/api/expenses", map[string]any{"user_id": 1001, "amount": 1, "category": "food", "spent_at": "2026-01-10"}); status != http.StatusUnprocessableEntity {
		t.Fatalf("expense for a pending member: %d", status)
	}

	if status, _ := e.do("PATCH", "/api/members/1001", map[string]any{"status": "boss"}); status != http.StatusUnprocessableEntity {
		t.Fatalf("unknown status: %d", status)
	}

	status, member := e.do("PATCH", "/api/members/1001", map[string]any{"status": "approved", "name": "Мама"})
	if status != 200 || member["display_name"] != "Мама" {
		t.Fatalf("approve: %d %v", status, member)
	}

	if got := e.tg.last(); !strings.Contains(got, "открыл доступ") {
		t.Fatalf("approval notice = %q", got)
	}

	e.telegramMessage(1001, "кофе 350", false)

	if got := e.tg.last(); !strings.Contains(got, "Записал") || !strings.Contains(got, "350 ₽") {
		t.Fatalf("bot reply = %q", got)
	}

	_, members := e.do("GET", "/api/members", nil) // массив: do кладёт его в "raw"
	if raw, _ := members["raw"].(string); !strings.Contains(raw, `"display_name":"Мама"`) || strings.Contains(raw, "1002") {
		t.Fatalf("members = %v: want the user and not the bot", members)
	}

	status, list := e.do("GET", "/api/expenses?user_id=1001", nil)

	items := list["items"].([]any)
	if status != 200 || len(items) != 1 || list["total"].(map[string]any)["amount"] != 350.0 {
		t.Fatalf("list: %d %v", status, list)
	}

	id := int(items[0].(map[string]any)["id"].(float64))

	status, created := e.do("POST", "/api/expenses", map[string]any{
		"user_id": 1001, "amount": 1200.5, "category": "transport", "spent_at": "2026-01-10", "comment": "такси",
	})
	if status != http.StatusCreated || created["amount"] != 1200.5 {
		t.Fatalf("create: %d %v", status, created)
	}

	if status, _ := e.do("POST", "/api/expenses", map[string]any{"user_id": 1002, "amount": 1, "category": "food", "spent_at": "2026-01-10"}); status != http.StatusUnprocessableEntity {
		t.Fatalf("create for unknown member (the bot was ignored): %d", status)
	}

	if status, _ := e.do("POST", "/api/expenses", map[string]any{"user_id": 1001, "amount": -5, "category": "food", "spent_at": "2026-01-10"}); status != http.StatusUnprocessableEntity {
		t.Fatalf("negative amount: %d", status)
	}

	path := "/api/expenses/" + strconv.Itoa(id)

	status, updated := e.do("PATCH", path, map[string]any{"amount": 450})
	if status != 200 || updated["amount"] != 450.0 || updated["comment"] != "кофе" {
		t.Fatalf("patch: %d %v", status, updated)
	}

	_, filtered := e.do("GET", "/api/expenses?category=transport", nil)
	if n := len(filtered["items"].([]any)); n != 1 {
		t.Fatalf("category filter: %d items", n)
	}

	if status, _ := e.do("DELETE", path, nil); status != http.StatusNoContent {
		t.Fatalf("delete: %d", status)
	}

	if status, _ := e.do("GET", path, nil); status != http.StatusNotFound {
		t.Fatalf("get deleted: %d", status)
	}
}

func TestLoginAttemptsAreLimited(t *testing.T) {
	e := setup(t)
	for range 5 {
		e.do("POST", "/api/login", map[string]string{"password": "guess"}, "X-Real-IP", "10.0.0.1")
	}

	if status, _ := e.do("POST", "/api/login", map[string]string{"password": password}, "X-Real-IP", "10.0.0.1"); status != http.StatusTooManyRequests {
		t.Fatalf("6th attempt: %d, want 429", status)
	}

	if status, _ := e.do("POST", "/api/login", map[string]string{"password": password}, "X-Real-IP", "10.0.0.2"); status != 200 {
		t.Fatalf("other IP: %d", status)
	}
}

func TestForgedToken(t *testing.T) {
	e := setup(t)

	e.token = "9999999999.forged"
	if status, _ := e.do("GET", "/api/expenses", nil); status != http.StatusUnauthorized {
		t.Fatalf("forged token: %d", status)
	}
}
