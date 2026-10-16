//go:build !(js && wasm)

package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"expensebot/internal/agent"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	apphttp "expensebot/internal/app/http"
)

// temperature не передаётся: новые модели Claude принимают только значение по умолчанию.
type claude struct {
	client anthropic.Client
	model  string
}

func newAnthropic(cfg Config, httpClient apphttp.Doer) (*claude, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("llm anthropic: API key is not set")
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey), option.WithHTTPClient(httpClient)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	for key, value := range cfg.ExtraBody {
		opts = append(opts, option.WithJSONSet(key, value))
	}

	return &claude{client: anthropic.NewClient(opts...), model: cfg.Model}, nil
}

func (c *claude) Complete(ctx context.Context, req agent.LLMRequest) (string, error) {
	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: int64(req.MaxTokens),
		System:    []anthropic.TextBlockParam{{Text: req.System}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(req.User))},
	})
	if err != nil {
		return "", fmt.Errorf("llm anthropic: %w", err)
	}

	if resp.StopReason == anthropic.StopReasonRefusal {
		return "", errors.New("llm anthropic: request refused")
	}

	var text strings.Builder
	for i := range resp.Content {
		if b, ok := resp.Content[i].AsAny().(anthropic.TextBlock); ok {
			text.WriteString(b.Text)
		}
	}

	return text.String(), nil
}
