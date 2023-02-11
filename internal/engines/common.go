package engines

import (
	"database/sql"
)

type Engine interface {
	Query(sql string, args ...any) (*sql.Rows, error)
	Execute(sql string, args ...any) error
	WithTx(fn func() error) error
}
