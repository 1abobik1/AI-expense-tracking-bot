package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"expensebot/internal/agent"

	apphttp "expensebot/internal/app/http"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	openRouterBaseURL    = "https://openrouter.ai/api/v1"
)

type openAI struct {
	cfg    Config
	client apphttp.Doer
}

func newOpenAI(cfg Config, client apphttp.Doer) *openAI {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultOpenAIBaseURL
	}

	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	return &openAI{cfg: cfg, client: client}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func chatBody(cfg Config, req agent.LLMRequest) map[string]any {
	body := map[string]any{
		"model": cfg.Model,
		"messages": []chatMessage{
			{Role: "system", Content: req.System},
			{Role: "user", Content: req.User},
		},
		"max_tokens":  req.MaxTokens,
		"temperature": 0,
	}
	maps.Copy(body, cfg.ExtraBody)

	return body
}

func (c *openAI) Complete(ctx context.Context, req agent.LLMRequest) (string, error) {
	payload, err := json.Marshal(chatBody(c.cfg, req))
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("llm openai: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm openai: HTTP %d: %s", resp.StatusCode, truncate(body))
	}

	var out struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("llm openai: decode response: %w", err)
	}

	if len(out.Choices) == 0 {
		return "", errors.New("llm openai: empty choices")
	}

	return out.Choices[0].Message.Content, nil
}
