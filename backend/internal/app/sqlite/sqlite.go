// Package sqlite — помощники репозиториев без драйвера: работают и с SQLite, и с D1.
package sqlite

import "time"

// TimeLayout — фиксированная ширина, чтобы строки сортировались как время.
const TimeLayout = "2006-01-02T15:04:05.000000Z"

func FormatTime(t time.Time) string { return t.UTC().Format(TimeLayout) }

func ParseTime(s string) (time.Time, error) { return time.Parse(time.RFC3339Nano, s) }

// Now — текущее время с точностью хранения в БД.
func Now() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }
