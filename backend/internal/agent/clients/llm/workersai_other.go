//go:build !(js && wasm)

package llm

import (
	"errors"

	"expensebot/internal/agent"
)

func newWorkersAI(Config) (agent.LLM, error) {
	return nil, errors.New("llm workersai: the AI binding exists only inside a Cloudflare Worker; " +
		"outside it use provider openai with the Workers AI REST endpoint")
}
