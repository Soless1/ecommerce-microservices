package product

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

// inMemoryRepository хранит товары в памяти процесса.
type inMemoryRepository struct {
	mu       sync.RWMutex
	nextSKU  uint32
	products map[uint32]entity.Product
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		nextSKU:  1,
		products: make(map[uint32]entity.Product),
	}
}

// CreateProduct создаёт товар и назначает ему следующий свободный SKU.
func (r *inMemoryRepository) CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error) {
	if err := ctx.Err(); err != nil {
		return entity.Product{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	product := entity.Product{
		SKU:   r.nextSKU,
		Name:  name,
		Price: price,
	}
	r.products[product.SKU] = product
	r.nextSKU++

	return product, nil
}

// GetProduct ищет товар по SKU.
func (r *inMemoryRepository) GetProduct(ctx context.Context, sku uint32) (entity.Product, error) {
	if err := ctx.Err(); err != nil {
		return entity.Product{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	product, ok := r.products[sku]
	if !ok {
		return entity.Product{}, entity.ErrProductNotFound
	}
	return product, nil
}
