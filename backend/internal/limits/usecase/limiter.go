package usecase

import (
	"context"
	"time"

	"expensebot/internal/app/clock"
	"expensebot/internal/limits"
)

const counterRetentionDays = 2

type Limiter struct {
	repo  limits.CountersRepo
	clock clock.Clock
}

func NewLimiter(repo limits.CountersRepo, clk clock.Clock) *Limiter {
	return &Limiter{repo: repo, clock: clk}
}

func (l *Limiter) Allow(ctx context.Context, key string, limit int, w limits.Window) (bool, error) {
	if limit <= 0 {
		return true, nil
	}

	now := l.clock.Time()

	n, err := l.repo.Incr(ctx, key, bucket(now, w))
	if err != nil {
		return false, err
	}

	if n == 1 && w == limits.PerDay {
		// Раз в сутки чистим старые счётчики.
		_ = l.repo.DeleteBefore(ctx, bucket(now.AddDate(0, 0, -counterRetentionDays), limits.PerDay))
	}

	return n <= limit, nil
}

// bucket сортируется как время: «2026-10-15» < «2026-10-15T13».
func bucket(t time.Time, w limits.Window) string {
	if w == limits.PerHour {
		return t.Format("2006-01-02T15")
	}

	return t.Format("2006-01-02")
}
