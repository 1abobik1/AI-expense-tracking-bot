package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"unicode/utf8"

	"expensebot/internal/agent"
	"expensebot/internal/app/apperr"
	"expensebot/internal/expenses"
	"expensebot/internal/expenses/models"
	"expensebot/internal/limits"
)

type Limits struct {
	MessagesPerUserPerDay int // 0 — без лимита
	LLMCallsPerDay        int // на весь бот: бережёт бесплатную квоту ИИ
	MaxMessageLength      int // в символах
}

const maxResponseTokens = 700 // хватает на несколько трат в одном сообщении

type AgentUseCase struct {
	expensesUC expenses.ExpensesUseCase
	llm        agent.LLM
	limiter    limits.Limiter
	limits     Limits
	log        *slog.Logger
}

func NewAgentUC(expensesUC expenses.ExpensesUseCase, llm agent.LLM, limiter limits.Limiter, l Limits, log *slog.Logger) *AgentUseCase {
	return &AgentUseCase{expensesUC: expensesUC, llm: llm, limiter: limiter, limits: l, log: log}
}

func (a *AgentUseCase) Handle(ctx context.Context, userID int64, text string) (string, error) {
	if a.limits.MaxMessageLength > 0 && utf8.RuneCountInString(text) > a.limits.MaxMessageLength {
		return fmt.Sprintf("Слишком длинное сообщение — уложись в %d символов.", a.limits.MaxMessageLength), nil
	}

	if ok, err := a.limiter.Allow(ctx, fmt.Sprintf("messages:%d", userID), a.limits.MessagesPerUserPerDay, limits.PerDay); err != nil {
		return "", err
	} else if !ok {
		return "На сегодня лимит сообщений исчерпан — продолжим завтра.", nil
	}

	if ok, err := a.limiter.Allow(ctx, "llm", a.limits.LLMCallsPerDay, limits.PerDay); err != nil {
		return "", err
	} else if !ok {
		return "ИИ на сегодня исчерпал бесплатный лимит — попробуй завтра.", nil
	}

	today := a.expensesUC.Today()

	raw, err := a.llm.Complete(ctx, agent.LLMRequest{System: systemPrompt(today), User: normalizeAmounts(text), MaxTokens: maxResponseTokens})
	if err != nil {
		a.log.Error("llm request failed", "err", err)
		return "ИИ сейчас недоступен, попробуй чуть позже.", nil
	}

	intent, ok := parseIntent(raw)
	if !ok {
		a.log.Warn("llm returned no intent", "response", raw)
		return HelpText, nil
	}

	reply, err := a.execute(ctx, models.UserScope(userID), intent, today)
	if e, ok := apperr.As(err); ok && e.Code == apperr.CodeValidation {
		return "Не получилось: " + e.Message + ".", nil
	}

	return reply, err
}

func (a *AgentUseCase) execute(ctx context.Context, scope models.Scope, in Intent, today models.Date) (string, error) {
	switch in.Action {
	case actionAdd:
		return a.add(ctx, scope, in, today)
	case actionList:
		return a.list(ctx, scope, in, today)
	case actionUpdate:
		return a.update(ctx, scope, in, today)
	case actionDelete:
		return a.delete(ctx, scope, in, today)
	}

	return HelpText, nil
}

func (a *AgentUseCase) add(ctx context.Context, scope models.Scope, in Intent, today models.Date) (string, error) {
	if len(in.Items) == 0 {
		return "Не понял сумму. Напиши, например: «кофе 350».", nil
	}

	var created []models.Expense
	for _, item := range in.Items {
		e, err := a.expensesUC.CreateExpense(ctx, scope, mapItemToExpense(item, today))
		if err != nil {
			return "", err
		}

		created = append(created, *e)
	}

	return formatAdded(created, today), nil
}

func (a *AgentUseCase) list(ctx context.Context, scope models.Scope, in Intent, today models.Date) (string, error) {
	from, to, label := resolvePeriod(in, today)
	category, _ := models.ParseCategory(in.Category)
	params := models.GetExpensesParams{From: from, To: to, Category: category, Limit: listPreview}

	summary, err := a.expensesUC.GetSummary(ctx, scope, params)
	if err != nil {
		return "", err
	}

	items, err := a.expensesUC.GetExpenses(ctx, scope, params)
	if err != nil {
		return "", err
	}

	return formatList(label, category, *summary, items, today), nil
}

func (a *AgentUseCase) update(ctx context.Context, scope models.Scope, in Intent, today models.Date) (string, error) {
	upd := mapChangesToUpdate(in.Changes)
	if upd.IsEmpty() {
		return "Не понял, что изменить. Например: «последнее было не 350, а 450».", nil
	}

	target, reply, err := a.findTarget(ctx, scope, in.Target, today)
	if target == nil {
		return reply, err
	}

	updated, err := a.expensesUC.UpdateExpense(ctx, scope, target.ID, upd)
	if err != nil {
		return "", err
	}

	return formatUpdated(*target, *updated, today), nil
}

func (a *AgentUseCase) delete(ctx context.Context, scope models.Scope, in Intent, today models.Date) (string, error) {
	target, reply, err := a.findTarget(ctx, scope, in.Target, today)
	if target == nil {
		return reply, err
	}

	deleted, err := a.expensesUC.DeleteExpense(ctx, scope, target.ID)
	if err != nil {
		return "", err
	}

	return "🗑 Удалил: " + formatExpense(*deleted, today), nil
}
