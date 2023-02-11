package engines

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Compile-time proof of interface implementation.
var _ Engine = (*SQLiteEngine)(nil)

type SQLiteEngine struct {
	connection *sql.DB
}

func CreateSQLite(url string) (*SQLiteEngine, error) {
	sqliteDatabase, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}

	return &SQLiteEngine{connection: sqliteDatabase}, nil
}

func (e *SQLiteEngine) Query(sql string, args ...any) (*sql.Rows, error) {
	result, err := e.connection.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (e *SQLiteEngine) Execute(sql string, args ...any) error {
	_, err := e.connection.Exec(sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (e *SQLiteEngine) WithTx(fn func() error) error {
	tx, err := e.connection.Begin()
	if err != nil {
		return err
	}

	if err = fn(); err != nil {
		tx.Rollback()
		return err
	} else {
		tx.Commit()
		return nil
	}
}
