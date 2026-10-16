package limits

import "context"

// CountersRepo хранит счётчики в БД: у Worker нет памяти между запросами.
type CountersRepo interface {
	Incr(ctx context.Context, key, bucket string) (int, error)
	DeleteBefore(ctx context.Context, bucket string) error
}
