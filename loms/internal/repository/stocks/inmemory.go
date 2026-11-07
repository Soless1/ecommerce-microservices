package stocks

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

type inMemoryRepository struct {
	mx      sync.RWMutex
	storage map[uint32]uint64
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		storage: make(map[uint32]uint64),
	}
}

func (r *inMemoryRepository) SetStock(ctx context.Context, sku uint32, count uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mx.Lock()
	defer r.mx.Unlock()
	r.storage[sku] = count
	return nil
}

func (r *inMemoryRepository) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	r.mx.RLock()
	defer r.mx.RUnlock()
	count, ok := r.storage[sku]
	if !ok {
		return 0, entity.ErrStockNotFound
	}
	return count, nil
}

// ReserveItems сначала проверяет весь заказ, затем уменьшает остатки.
// Один Lock защищает всю операцию от параллельного резервирования.
func (r *inMemoryRepository) ReserveItems(ctx context.Context, items []entity.OrderItem) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	requested := make(map[uint32]uint64)
	for _, item := range items {
		requested[item.SKU] += uint64(item.Count)
	}

	for sku, count := range requested {
		available, ok := r.storage[sku]
		if !ok || available < count {
			return entity.ErrInsufficientStock
		}
	}

	for sku, count := range requested {
		r.storage[sku] -= count
	}

	return nil
}

// ReleaseItems возвращает в доступный остаток зарезервированные позиции.
func (r *inMemoryRepository) ReleaseItems(ctx context.Context, items []entity.OrderItem) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	requested := make(map[uint32]uint64)
	for _, item := range items {
		requested[item.SKU] += uint64(item.Count)
	}

	for sku := range requested {
		if _, ok := r.storage[sku]; !ok {
			return entity.ErrStockNotFound
		}
	}

	for sku, count := range requested {
		r.storage[sku] += count
	}

	return nil
}
