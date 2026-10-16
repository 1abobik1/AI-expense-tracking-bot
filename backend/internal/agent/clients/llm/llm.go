// Package llm — адаптеры языковых моделей для порта agent.LLM.
package llm

import (
	"fmt"
	"strings"

	"expensebot/internal/agent"

	apphttp "expensebot/internal/app/http"
)

const (
	ProviderOpenAI     = "openai"     // любой OpenAI-совместимый API: OpenAI, DeepSeek, Groq, Ollama, Workers AI по REST
	ProviderOpenRouter = "openrouter" // сотни моделей через один ключ; тот же API, адрес по умолчанию свой
	ProviderAnthropic  = "anthropic"  // Claude через официальный SDK
	ProviderWorkersAI  = "workersai"  // Cloudflare Workers AI через binding; только внутри Worker
)

type Config struct {
	Provider string
	Model    string
	BaseURL  string
	APIKey   string
	// ExtraBody добавляется в запрос как есть: параметры конкретной модели.
	ExtraBody map[string]any
}

func New(cfg Config, httpClient apphttp.Doer) (agent.LLM, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("llm: model is not set")
	}

	switch strings.ToLower(cfg.Provider) {
	case ProviderOpenAI:
		return newOpenAI(cfg, httpClient), nil
	case ProviderOpenRouter:
		if cfg.BaseURL == "" {
			cfg.BaseURL = openRouterBaseURL
		}

		return newOpenAI(cfg, httpClient), nil
	case ProviderAnthropic:
		return newAnthropic(cfg, httpClient)
	case ProviderWorkersAI:
		return newWorkersAI(cfg)
	}

	return nil, fmt.Errorf("llm: unknown provider %q (want %s, %s, %s or %s)",
		cfg.Provider, ProviderOpenAI, ProviderOpenRouter, ProviderAnthropic, ProviderWorkersAI)
}

func truncate(b []byte) string {
	const limit = 300
	if len(b) > limit {
		return string(b[:limit]) + "…"
	}

	return string(b)
}
