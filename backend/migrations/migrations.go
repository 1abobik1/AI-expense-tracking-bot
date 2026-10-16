// Package migrations — схема для goose; для D1 make cf-migrate берёт только секции Up.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
