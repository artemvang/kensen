package engines

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Compile-time proof of interface implementation.
var _ Engine = (*MySQLEngine)(nil)

type MySQLEngine struct {
	connection *sql.DB
}

func CreateMySQL(url string) (*SQLiteEngine, error) {
	sqliteDatabase, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	return &SQLiteEngine{connection: sqliteDatabase}, nil
}

func (e *MySQLEngine) Query(sql string, args ...any) (*sql.Rows, error) {
	result, err := e.connection.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (e *MySQLEngine) Execute(sql string, args ...any) error {
	_, err := e.connection.Exec(sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (e *MySQLEngine) WithTx(fn func() error) error {
	tx, err := e.connection.Begin()
	if err != nil {
		return err
	}

	if err = fn(); err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}
