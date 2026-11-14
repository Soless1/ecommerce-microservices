package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/cart/sqlc"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/transactor"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	queries *sqlc.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		queries: sqlc.New(pool),
	}
}

func (r *postgresRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *postgresRepository) AddItem(ctx context.Context, userId int64, newItem entity.CartItem) error {
	queries := r.getQueries(ctx)
	err := queries.AddItem(ctx, sqlc.AddItemParams{
		UserID: userId,
		Sku:    int64(newItem.SKU),
		Count:  int64(newItem.Count),
	})

	return err
}

func (r *postgresRepository) DeleteItem(ctx context.Context, userId int64, sku uint32) error {
	queries := r.getQueries(ctx)
	_, err := queries.DeleteItem(ctx, sqlc.DeleteItemParams{UserID: userId, Sku: int64(sku)})
	return err
}

func (r *postgresRepository) GetItemsByUserID(ctx context.Context, userId int64) ([]entity.CartItem, error) {
	queries := r.getQueries(ctx)
	rows, err := queries.GetItemsByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	items := make([]entity.CartItem, 0, len(rows))
	for _, item := range rows {
		items = append(items, entity.CartItem{SKU: uint32(item.Sku), Count: uint32(item.Count)})
	}
	return items, nil
}

func (r *postgresRepository) DeleteItemsByUserID(ctx context.Context, userId int64) error {
	queries := r.getQueries(ctx)
	_, err := queries.DeleteItemsByUserID(ctx, userId)
	if err != nil {
		return err
	}
	return nil
}
