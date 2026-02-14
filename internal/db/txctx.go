package db

import (
	"context"
	"database/sql"
)

// txKey is the unexported context key for carrying a transaction.
type txKey struct{}

// DBTX is the interface satisfied by both *sql.DB, *sql.Tx, and the
// immediateTx wrapper. Repositories use this to transparently run
// queries inside or outside a transaction.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// NewTxContext returns a new context that carries the given DBTX.
func NewTxContext(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext extracts the DBTX from ctx.
// Returns nil if no transaction is present.
func TxFromContext(ctx context.Context) DBTX {
	tx, _ := ctx.Value(txKey{}).(DBTX)
	return tx
}
