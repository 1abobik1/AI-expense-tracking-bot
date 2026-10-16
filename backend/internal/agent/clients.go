package agent

import "context"

// LLM — языковая модель; провайдер выбирается конфигом (clients/llm).
type LLM interface {
	Complete(ctx context.Context, req LLMRequest) (string, error)
}

// Messenger — ответы пользователю в Telegram.
type Messenger interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
	SendTyping(ctx context.Context, chatID int64) error
}

type LLMRequest struct {
	System    string
	User      string
	MaxTokens int
}
