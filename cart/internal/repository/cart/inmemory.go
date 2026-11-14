package cart

import (
	"context"
	"sync"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
)

type InMemoryRepository struct {
	mx    sync.RWMutex
	carts map[int64]map[uint32]entity.CartItem
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		carts: make(map[int64]map[uint32]entity.CartItem),
	}
}

func (r *InMemoryRepository) AddItem(ctx context.Context, userId int64, newItem entity.CartItem) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mx.Lock()
	defer r.mx.Unlock()

	userCart, ok := r.carts[userId]
	if !ok {
		userCart = make(map[uint32]entity.CartItem)
		r.carts[userId] = userCart
	}

	item, ok := userCart[newItem.SKU]
	if ok {
		item.Count += newItem.Count
	} else {
		item = newItem
	}

	userCart[newItem.SKU] = item
	return nil
}

func (r *InMemoryRepository) DeleteItem(ctx context.Context, userId int64, sku uint32) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mx.Lock()
	defer r.mx.Unlock()

	userCart, ok := r.carts[userId]
	if !ok {
		return entity.ErrCartItemNotFound
	}
	//if _, ok := userCart[sku]; !ok {
	//	return entity.ErrCartItemNotFound
	//}

	delete(userCart, sku)
	return nil
}

func (r *InMemoryRepository) GetItemsByUserID(ctx context.Context, userId int64) ([]entity.CartItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mx.RLock()
	defer r.mx.RUnlock()
	userCart, ok := r.carts[userId]
	if !ok {
		return []entity.CartItem{}, nil
	}
	items := make([]entity.CartItem, 0, len(userCart))
	for _, item := range userCart {
		items = append(items, item)
	}
	return items, nil
}

func (r *InMemoryRepository) DeleteItemsByUserID(ctx context.Context, userId int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mx.Lock()
	defer r.mx.Unlock()
	delete(r.carts, userId)
	return nil
}
