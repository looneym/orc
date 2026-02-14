package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/example/orc/internal/db"
	"github.com/example/orc/internal/ports/secondary"
)

// Transactor implements secondary.Transactor using SQLite BEGIN IMMEDIATE.
type Transactor struct {
	db *sql.DB
}

// NewTransactor creates a new SQLite Transactor.
func NewTransactor(database *sql.DB) *Transactor {
	return &Transactor{db: database}
}

// WithImmediateTx executes fn within a BEGIN IMMEDIATE transaction.
// The transaction is carried in the context so repositories can detect it
// via db.TxFromContext and run queries on the same transaction.
func (t *Transactor) WithImmediateTx(ctx context.Context, fn func(ctx context.Context) error) error {
	conn, err := t.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("begin immediate: %w", err)
	}

	// Wrap the conn in an immediateTx so it satisfies *sql.Tx-like usage
	// through the context. We store the conn as a *sql.Tx would not let
	// us issue BEGIN IMMEDIATE. Instead, we use a thin wrapper.
	itx := &immediateTx{conn: conn}
	txCtx := db.NewTxContext(ctx, itx)

	if err := fn(txCtx); err != nil {
		_, _ = conn.ExecContext(ctx, "ROLLBACK") // best-effort
		return err
	}

	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		_, _ = conn.ExecContext(ctx, "ROLLBACK") // best-effort
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// immediateTx wraps a *sql.Conn to satisfy the db.DBTX interface.
// This allows repositories to use ExecContext/QueryContext/QueryRowContext
// on the connection that owns the BEGIN IMMEDIATE transaction.
type immediateTx struct {
	conn *sql.Conn
}

func (t *immediateTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.conn.ExecContext(ctx, query, args...)
}

func (t *immediateTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.conn.QueryContext(ctx, query, args...)
}

func (t *immediateTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.conn.QueryRowContext(ctx, query, args...)
}

// Ensure Transactor implements the interface.
var _ secondary.Transactor = (*Transactor)(nil)
