package limits

import "context"

type Window int

const (
	PerHour Window = iota
	PerDay
)

type Limiter interface {
	// Allow засчитывает попытку; limit <= 0 — без ограничения и без записи в БД.
	Allow(ctx context.Context, key string, limit int, w Window) (bool, error)
}
