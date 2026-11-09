package transactor

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

var ErrTxNotFound = errors.New("transaction not found in context")

type transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *transactor {
	return &transactor{pool: pool}
}

func ExtractTx(ctx context.Context) (pgx.Tx, error) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return nil, ErrTxNotFound
	}
	return tx, nil
}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func (t *transactor) WithTx(ctx context.Context, f func(context.Context) error) error {
	if _, err := ExtractTx(ctx); err == nil {
		return f(ctx)
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	txCtx := injectTx(ctx, tx)
	if err := f(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
