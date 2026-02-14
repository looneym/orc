package secondary

import "context"

// Transactor defines the secondary port for transactional execution.
// Implementations wrap the provided function in a database transaction.
type Transactor interface {
	// WithImmediateTx executes fn within a BEGIN IMMEDIATE transaction.
	// If fn returns nil, the transaction is committed; otherwise it is rolled back.
	// The context passed to fn carries the active transaction so that
	// repositories can detect and reuse it via db.TxFromContext.
	WithImmediateTx(ctx context.Context, fn func(ctx context.Context) error) error
}
