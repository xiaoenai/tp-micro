package sqlx

import (
	"context"
	"database/sql"
)

// DbOrTx contains all the exportable methods of *DB
type DbOrTx interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*Rows, error)
	// NamedQueryContext(ctx context.Context, query string, arg interface{}) (*Rows, error)
	PrepareNamed(query string) (*NamedStmt, error)
	// PrepareNamedContext(ctx context.Context, query string) (*NamedStmt, error)
	Preparex(query string) (*Stmt, error)
	// PreparexContext(ctx context.Context, query string) (*Stmt, error)
	QueryRowx(query string, args ...interface{}) *Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *Row
	Queryx(query string, args ...interface{}) (*Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
