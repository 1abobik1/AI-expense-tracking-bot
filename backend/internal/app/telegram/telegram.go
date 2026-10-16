// Package telegram — минимальный клиент Bot API.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	apphttp "expensebot/internal/app/http"
)

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"` // private, group, supergroup, channel
}

type Client struct {
	base string
	http apphttp.Doer
}

func New(token, apiBase string, httpClient apphttp.Doer) *Client {
	if apiBase == "" {
		apiBase = "https://api.telegram.org"
	}

	return &Client{base: strings.TrimRight(apiBase, "/") + "/bot" + token + "/", http: httpClient}
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.call(ctx, "sendMessage", map[string]any{"chat_id": chatID, "text": text}, nil)
}

func (c *Client) SendTyping(ctx context.Context, chatID int64) error {
	return c.call(ctx, "sendChatAction", map[string]any{"chat_id": chatID, "action": "typing"}, nil)
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	var updates []Update
	err := c.call(ctx, "getUpdates", map[string]any{
		"offset": offset, "timeout": timeoutSec, "allowed_updates": []string{"message"},
	}, &updates)

	return updates, err
}

// DeleteWebhook нужен перед polling: при вебхуке getUpdates не работает.
func (c *Client) DeleteWebhook(ctx context.Context) error {
	return c.call(ctx, "deleteWebhook", map[string]any{}, nil)
}

func (c *Client) call(ctx context.Context, method string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+method, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// В ошибке URL с токеном.
		return fmt.Errorf("telegram %s: request failed: %w", method, redact(err, c.base))
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}

	var out struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return fmt.Errorf("telegram %s: HTTP %d: bad response", method, resp.StatusCode)
	}

	if !out.OK {
		return fmt.Errorf("telegram %s: %s", method, out.Description)
	}

	if result != nil {
		return json.Unmarshal(out.Result, result)
	}

	return nil
}

type redactedError struct{ msg string }

func (e redactedError) Error() string { return e.msg }

func redact(err error, secret string) error {
	return redactedError{strings.ReplaceAll(err.Error(), secret, "https://api.telegram.org/bot***/")}
}
