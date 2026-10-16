//go:build js && wasm

package llm

import (
	"errors"

	"expensebot/internal/agent"

	apphttp "expensebot/internal/app/http"
)

// SDK Anthropic не влезает в лимит Worker (+2.5 МБ к 3 МБ), Claude — только на сервере.
func newAnthropic(Config, apphttp.Doer) (agent.LLM, error) {
	return nil, errors.New("llm anthropic: not available in the Cloudflare Worker build, use the server deployment")
}
