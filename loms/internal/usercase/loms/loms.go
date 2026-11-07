package loms

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

var (
	ErrOrderNotFound     = entity.ErrOrderNotFound
	ErrInsufficientStock = entity.ErrInsufficientStock
	ErrInvalidStatus     = entity.ErrInvalidStatus
	ErrWrongOrder        = errors.New("wrong order")
)

type (
	orderRepository interface {
		CreateOrder(ctx context.Context, newOrder entity.Order) (entity.Order, error)
		GetOrder(ctx context.Context, id int64) (entity.Order, error)
		SetOrderStatus(ctx context.Context, id int64, status entity.OrderStatus) error
		TransitionOrderStatus(
			ctx context.Context,
			id int64,
			expected entity.OrderStatus,
			next entity.OrderStatus,
		) error
	}

	stocksRepository interface {
		ReserveItems(ctx context.Context, items []entity.OrderItem) error
		ReleaseItems(ctx context.Context, items []entity.OrderItem) error
	}
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mock.go -package=mocks

func NewLomsService(
	orderRepository orderRepository,
	stocksRepository stocksRepository,
) *lomsService {
	return &lomsService{
		orderRepository:  orderRepository,
		stocksRepository: stocksRepository,
	}
}

type lomsService struct {
	orderRepository  orderRepository
	stocksRepository stocksRepository
}

func (l *lomsService) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	if len(items) == 0 {
		return 0, ErrWrongOrder
	}
	for _, item := range items {
		if item.Count == 0 {
			return 0, ErrWrongOrder
		}
	}

	if err := l.stocksRepository.ReserveItems(ctx, items); err != nil {
		return 0, err
	}

	order, err := l.orderRepository.CreateOrder(ctx, entity.Order{
		UserID: userID,
		Items:  items,
		Status: entity.OrderStatusAwaitingPayment,
	})
	if err != nil {
		releaseErr := l.stocksRepository.ReleaseItems(
			context.WithoutCancel(ctx),
			items,
		)
		if releaseErr != nil {
			return 0, errors.Join(err, releaseErr)
		}
		return 0, err
	}

	return order.ID, nil
}

func (l *lomsService) GetOrder(ctx context.Context, id int64) (entity.Order, error) {
	order, err := l.orderRepository.GetOrder(ctx, id)
	if err != nil {
		return entity.Order{}, err
	}
	return order, nil
}

func (l *lomsService) PayOrder(ctx context.Context, id int64) error {
	return l.orderRepository.TransitionOrderStatus(ctx, id, entity.OrderStatusAwaitingPayment, entity.OrderStatusPaid)
}
func (l *lomsService) CancelOrder(ctx context.Context, id int64) error {
	order, err := l.orderRepository.GetOrder(ctx, id)
	if err != nil {
		return err
	}

	// Только один параллельный запрос сможет отменить заказ.
	err = l.orderRepository.TransitionOrderStatus(
		ctx,
		id,
		entity.OrderStatusAwaitingPayment,
		entity.OrderStatusCancelled,
	)
	if err != nil {
		return err
	}

	// Отмена уже началась — завершаем возврат остатков,
	// даже если клиент успел отменить запрос.
	workCtx := context.WithoutCancel(ctx)
	err = l.stocksRepository.ReleaseItems(workCtx, order.Items)
	if err == nil {
		return nil
	}

	// Остатки не вернулись: откатываем статус,
	// чтобы отмену можно было повторить.
	rollbackErr := l.orderRepository.TransitionOrderStatus(
		workCtx,
		id,
		entity.OrderStatusCancelled,
		entity.OrderStatusAwaitingPayment,
	)
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
