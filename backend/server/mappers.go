package server

import (
	"expensebot/config"
	"expensebot/internal/agent/clients/llm"
	agentUsecase "expensebot/internal/agent/usecase"
)

func mapLLMConfig(c config.LLMConfig) llm.Config {
	return llm.Config{Provider: c.Provider, Model: c.Model, BaseURL: c.BaseURL, APIKey: c.APIKey, ExtraBody: c.ExtraBody}
}

func mapLimits(c config.LimitsConfig) agentUsecase.Limits {
	return agentUsecase.Limits{
		MessagesPerUserPerDay: c.MessagesPerUserPerDay,
		LLMCallsPerDay:        c.LLMCallsPerDay,
		MaxMessageLength:      c.MaxMessageLength,
	}
}
