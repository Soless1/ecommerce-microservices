package product

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	"github.com/redis/go-redis/v9"
)

type (
	productRepository interface {
		CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error)
		GetProduct(ctx context.Context, sku uint32) (entity.Product, error)
	}
)

type cachedRepository struct {
	productRepository productRepository
	cache             *redis.Client
	cfg               *config.Config
}

func NewCachedRepository(
	productRepository productRepository,
	cache *redis.Client,
	cfg *config.Config,
) *cachedRepository {
	if cfg.Redis.ProductCacheTTL <= 0 || cfg.Redis.ProductCacheTimeout <= 0 {
		cache = nil
	}

	return &cachedRepository{
		productRepository: productRepository,
		cache:             cache,
		cfg:               cfg,
	}
}

func (r *cachedRepository) CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error) {
	return r.productRepository.CreateProduct(ctx, name, price)
}

func (r *cachedRepository) GetProduct(ctx context.Context, sku uint32) (entity.Product, error) {
	if r.cache == nil {
		return r.productRepository.GetProduct(ctx, sku)
	}
	// Чтение в транзакции обходит кэш, чтобы не сохранить неподтверждённые данные.
	if _, err := transactor.ExtractTx(ctx); err == nil {
		return r.productRepository.GetProduct(ctx, sku)
	}

	key := "loms:product:v1:" + strconv.FormatUint(uint64(sku), 10)
	cacheCtx, cancel := context.WithTimeout(ctx, r.cfg.Redis.ProductCacheTimeout)
	value, err := r.cache.Get(cacheCtx, key).Bytes()
	cancel()
	if err != nil && !errors.Is(err, redis.Nil) {
		return r.productRepository.GetProduct(ctx, sku)
	}

	if len(value) > 0 {
		var product *entity.Product
		if decodeErr := json.Unmarshal(value, &product); decodeErr == nil && product != nil && product.SKU == sku {
			return *product, nil
		}
	}

	product, err := r.productRepository.GetProduct(ctx, sku)
	if err != nil {
		return product, err
	}

	value, err = json.Marshal(product)
	if err != nil {
		return product, nil
	}

	cacheCtx, cancel = context.WithTimeout(ctx, r.cfg.Redis.ProductCacheTimeout)
	defer cancel()
	_ = r.cache.Set(cacheCtx, key, value, r.cfg.Redis.ProductCacheTTL).Err()

	return product, nil
}
