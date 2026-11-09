package stocks

import (
	"context"
	"errors"
	"fmt"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks/sqlc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	queries *sqlc.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{queries: sqlc.New(pool)}
}

func (r *postgresRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *postgresRepository) SetStock(ctx context.Context, sku uint32, count uint64) error {
	queries := r.getQueries(ctx)
	return queries.SetStock(ctx, sqlc.SetStockParams{Sku: int64(sku), Count: int64(count)})
}

func (r *postgresRepository) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	queries := r.getQueries(ctx)
	count, err := queries.GetStock(ctx, int64(sku))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, entity.ErrStockNotFound
		}
		return 0, err
	}
	return uint64(count), nil
}

func (r *postgresRepository) ReserveItems(ctx context.Context, items []entity.OrderItem) error {
	queries := r.getQueries(ctx)
	requested, err := aggregateItems(items)
	if err != nil {
		return err
	}

	for sku, count := range requested {
		rows, err := queries.ReserveStock(ctx, sqlc.ReserveStockParams{Sku: int64(sku), Count: count})
		if err != nil {
			return err
		}
		if rows == 0 {
			return entity.ErrInsufficientStock
		}
	}
	return nil
}

func (r *postgresRepository) ReleaseItems(ctx context.Context, items []entity.OrderItem) error {
	queries := r.getQueries(ctx)

	requested, err := aggregateItems(items)
	if err != nil {
		return err
	}

	for sku, count := range requested {
		rows, err := queries.ReleaseStock(ctx, sqlc.ReleaseStockParams{Sku: int64(sku), Count: count})
		if err != nil {
			return err
		}
		if rows == 0 {
			return entity.ErrStockNotFound
		}
	}
	return nil
}

func aggregateItems(items []entity.OrderItem) (map[uint32]int64, error) {
	requested := make(map[uint32]uint64, len(items))
	for _, item := range items {
		if item.Count == 0 {
			return nil, fmt.Errorf("stock count must be positive")
		}
		requested[item.SKU] += uint64(item.Count)
	}

	result := make(map[uint32]int64, len(requested))
	for sku, count := range requested {
		result[sku] = int64(count)
	}
	return result, nil
}
