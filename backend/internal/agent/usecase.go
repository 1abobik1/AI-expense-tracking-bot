package agent

import "context"

// AgentUseCase: модель только разбирает текст в намерение, остальное делает код в Scope автора.
type AgentUseCase interface {
	Handle(ctx context.Context, userID int64, text string) (string, error)
}
