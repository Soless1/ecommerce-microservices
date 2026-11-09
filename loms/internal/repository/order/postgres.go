package order

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/sqlc"
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

func (r *postgresRepository) CreateOrder(ctx context.Context, newOrder entity.Order) (entity.Order, error) {
	queries := r.getQueries(ctx)

	row, err := queries.CreateOrder(ctx, sqlc.CreateOrderParams{UserID: newOrder.UserID, Status: statusToPostgres(newOrder.Status)})
	if err != nil {
		return entity.Order{}, err
	}
	newOrder.ID = row.OrderID
	newOrder.UserID = row.UserID
	newOrder.Status = statusFromPostgres(row.Status)
	newOrder.CreatedAt = row.CreatedAt.Time
	newOrder.UpdatedAt = row.UpdatedAt.Time

	for _, item := range newOrder.Items {
		if item.Count == 0 {
			return entity.Order{}, fmt.Errorf("order item count must be positive")
		}
		if err := queries.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID: newOrder.ID,
			Sku:     int64(item.SKU),
			Count:   int64(item.Count),
		}); err != nil {
			return entity.Order{}, err
		}
	}
	newOrder.Items = append([]entity.OrderItem(nil), newOrder.Items...)
	return newOrder, nil
}

func (r *postgresRepository) GetOrder(ctx context.Context, id int64) (entity.Order, error) {
	queries := r.getQueries(ctx)
	row, err := queries.GetOrder(ctx, id)
	return loadOrder(ctx, queries, id, row, err)
}

func (r *postgresRepository) GetOrderForUpdate(ctx context.Context, id int64) (entity.Order, error) {
	queries := r.getQueries(ctx)
	row, err := queries.GetOrderForUpdate(ctx, id)
	return loadOrder(ctx, queries, id, row, err)
}

func loadOrder(ctx context.Context, queries *sqlc.Queries, id int64, row sqlc.LomsOrder, err error) (entity.Order, error) {
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Order{}, entity.ErrOrderNotFound
		}
		return entity.Order{}, err
	}
	rows, err := queries.GetOrderItems(ctx, id)
	if err != nil {
		return entity.Order{}, err
	}
	items := make([]entity.OrderItem, 0, len(rows))
	for _, item := range rows {
		items = append(items, entity.OrderItem{SKU: uint32(item.Sku), Count: uint32(item.Count)})
	}

	return entity.Order{
		ID: row.OrderID, UserID: row.UserID, Status: statusFromPostgres(row.Status),
		Items: items, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (r *postgresRepository) SetOrderStatus(ctx context.Context, id int64, status entity.OrderStatus) error {
	queries := r.getQueries(ctx)

	rows, err := queries.SetOrderStatus(ctx, sqlc.SetOrderStatusParams{OrderID: id, Status: statusToPostgres(status)})
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrOrderNotFound
	}
	return nil
}

func statusToPostgres(status entity.OrderStatus) sqlc.LomsOrderStatus {
	return sqlc.LomsOrderStatus(strings.ReplaceAll(string(status), "_", " "))
}

func statusFromPostgres(status sqlc.LomsOrderStatus) entity.OrderStatus {
	return entity.OrderStatus(strings.ReplaceAll(string(status), " ", "_"))
}
