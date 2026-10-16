package usecase

import (
	"context"
	"fmt"

	"expensebot/internal/app/apperr"
	"expensebot/internal/app/clock"
	"expensebot/internal/expenses"
	"expensebot/internal/expenses/models"
	"expensebot/internal/members"

	membersModels "expensebot/internal/members/models"
)

const (
	defaultLimit = 50
	maxLimit     = 500
)

type ExpensesUseCase struct {
	repo    expenses.ExpensesRepo
	members members.MembersRepo
	clock   clock.Clock
}

func NewExpensesUC(repo expenses.ExpensesRepo, membersRepo members.MembersRepo, clk clock.Clock) *ExpensesUseCase {
	return &ExpensesUseCase{repo: repo, members: membersRepo, clock: clk}
}

func (uc *ExpensesUseCase) Today() models.Date { return models.DateOf(uc.clock.Time()) }

func (uc *ExpensesUseCase) CreateExpense(ctx context.Context, scope models.Scope, e models.Expense) (*models.Expense, error) {
	if scope.IsAdmin() {
		// Админ добавляет траты только одобренным участникам.
		m, err := uc.members.GetMember(ctx, e.UserID)
		if apperr.Is(err, apperr.CodeNotFound) || (err == nil && m.Status != membersModels.StatusApproved) {
			return nil, expenses.ErrUnknownMember()
		} else if err != nil {
			return nil, fmt.Errorf("members.GetMember: %w", err)
		}
	} else {
		e.UserID = scope.UserID()
	}

	if err := uc.validate(&e); err != nil {
		return nil, err
	}

	if err := uc.repo.CreateExpense(ctx, &e); err != nil {
		return nil, fmt.Errorf("repo.CreateExpense: %w", err)
	}

	return &e, nil
}

func (uc *ExpensesUseCase) GetExpense(ctx context.Context, scope models.Scope, id int64) (*models.Expense, error) {
	return uc.repo.GetExpense(ctx, scope, id)
}

func (uc *ExpensesUseCase) GetExpenses(ctx context.Context, scope models.Scope, params models.GetExpensesParams) ([]models.Expense, error) {
	switch {
	case params.Limit <= 0:
		params.Limit = defaultLimit
	case params.Limit > maxLimit:
		params.Limit = maxLimit
	}

	params.Offset = max(params.Offset, 0)

	return uc.repo.GetExpenses(ctx, scope, params)
}

func (uc *ExpensesUseCase) GetSummary(ctx context.Context, scope models.Scope, params models.GetExpensesParams) (*models.Summary, error) {
	return uc.repo.GetSummary(ctx, scope, params)
}

func (uc *ExpensesUseCase) UpdateExpense(ctx context.Context, scope models.Scope, id int64, in models.UpdateExpenseIn) (*models.Expense, error) {
	e, err := uc.repo.GetExpense(ctx, scope, id)
	if err != nil {
		return nil, err
	}

	if in.IsEmpty() {
		return e, nil
	}

	e.Apply(in)

	if err := uc.validate(e); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateExpense(ctx, scope, e); err != nil {
		return nil, err
	}

	return e, nil
}

func (uc *ExpensesUseCase) DeleteExpense(ctx context.Context, scope models.Scope, id int64) (*models.Expense, error) {
	e, err := uc.repo.GetExpense(ctx, scope, id)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.DeleteExpense(ctx, scope, id); err != nil {
		return nil, err
	}

	return e, nil
}

func (uc *ExpensesUseCase) validate(e *models.Expense) error {
	if err := e.Validate(); err != nil {
		return err
	}

	if e.SpentAt.After(uc.Today()) {
		return expenses.ErrFutureDate()
	}

	return nil
}
