package product

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product/sqlc"
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

func (r *postgresRepository) CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error) {
	queries := r.getQueries(ctx)
	row, err := queries.CreateProduct(ctx, sqlc.CreateProductParams{Name: name, Price: int64(price)})
	if err != nil {
		return entity.Product{}, err
	}
	return productFromRow(row.Sku, row.Name, row.Price), nil
}

func (r *postgresRepository) GetProduct(ctx context.Context, sku uint32) (entity.Product, error) {
	queries := r.getQueries(ctx)
	row, err := queries.GetProduct(ctx, int64(sku))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Product{}, entity.ErrProductNotFound
		}
		return entity.Product{}, err
	}
	return productFromRow(row.Sku, row.Name, row.Price), nil
}

func productFromRow(sku int64, name string, price int64) entity.Product {
	return entity.Product{SKU: uint32(sku), Name: name, Price: uint32(price)}
}
