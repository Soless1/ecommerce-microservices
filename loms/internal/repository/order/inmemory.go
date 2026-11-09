package order

import (
	"context"
	"sync"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

// inMemoryRepository хранит заказы в памяти процесса.
type inMemoryRepository struct {
	mx     sync.RWMutex
	nextID int64
	orders map[int64]entity.Order
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		nextID: 1,
		orders: make(map[int64]entity.Order),
	}
}

// CreateOrder сохраняет заказ, назначает ему ID и выставляет время создания.
func (r *inMemoryRepository) CreateOrder(ctx context.Context, newOrder entity.Order) (entity.Order, error) {
	if err := ctx.Err(); err != nil {
		return entity.Order{}, err
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	now := time.Now()
	newOrder.ID = r.nextID
	newOrder.CreatedAt = now
	newOrder.UpdatedAt = now
	newOrder.Items = cloneItems(newOrder.Items)

	r.orders[newOrder.ID] = newOrder
	r.nextID++

	return cloneOrder(newOrder), nil
}

// GetOrder ищет заказ по ID.
func (r *inMemoryRepository) GetOrder(ctx context.Context, id int64) (entity.Order, error) {
	if err := ctx.Err(); err != nil {
		return entity.Order{}, err
	}

	r.mx.RLock()
	defer r.mx.RUnlock()

	order, ok := r.orders[id]
	if !ok {
		return entity.Order{}, entity.ErrOrderNotFound
	}

	return cloneOrder(order), nil
}

func (r *inMemoryRepository) GetOrderForUpdate(ctx context.Context, id int64) (entity.Order, error) {
	return r.GetOrder(ctx, id)
}

// SetOrderStatus меняет статус заказа и время последнего изменения.
func (r *inMemoryRepository) SetOrderStatus(ctx context.Context, id int64, status entity.OrderStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	order, ok := r.orders[id]
	if !ok {
		return entity.ErrOrderNotFound
	}

	order.Status = status
	order.UpdatedAt = time.Now()
	r.orders[id] = order
	return nil
}

func cloneOrder(order entity.Order) entity.Order {
	order.Items = cloneItems(order.Items)
	return order
}

func cloneItems(items []entity.OrderItem) []entity.OrderItem {
	return append([]entity.OrderItem(nil), items...)
}
