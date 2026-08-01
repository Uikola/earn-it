package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transactor runs logic inside a single database transaction
type Transactor interface {
	// WithinTransaction runs a function within a database transaction.
	//
	// Transaction is propagated in the context,
	// so it is important to propagate it to underlying repositories.
	// Function commits if error is nil, and rollbacks if not.
	// It returns the same error.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type txCtxKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func extractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx)
	return tx, ok
}

type pgxTransactor struct {
	pool *pgxpool.Pool
}

func NewPgxTransactor(pool *pgxpool.Pool) Transactor {
	return &pgxTransactor{
		pool: pool,
	}
}

func (t *pgxTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err = fn(injectTx(ctx, tx)); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

type QueryExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func GetQueryExecutor(ctx context.Context, pool *pgxpool.Pool) QueryExecutor {
	if tx, ok := extractTx(ctx); ok {
		return tx
	}
	return pool
}
