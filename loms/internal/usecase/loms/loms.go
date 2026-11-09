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
		GetOrderForUpdate(ctx context.Context, id int64) (entity.Order, error)
		SetOrderStatus(ctx context.Context, id int64, status entity.OrderStatus) error
	}

	stocksRepository interface {
		ReserveItems(ctx context.Context, items []entity.OrderItem) error
		ReleaseItems(ctx context.Context, items []entity.OrderItem) error
	}
	transactor interface {
		WithTx(ctx context.Context, f func(context.Context) error) error
	}
	notificationsClient interface {
		SendOrderStatusChangedNotification(ctx context.Context, userId int64, orderId int64, status entity.OrderStatus) error
	}
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mock.go -package=mocks

func NewLomsService(
	orderRepository orderRepository,
	stocksRepository stocksRepository,
	transactor transactor,
	notificationsClient notificationsClient,
) *lomsService {
	return &lomsService{
		orderRepository:     orderRepository,
		stocksRepository:    stocksRepository,
		transactor:          transactor,
		notificationsClient: notificationsClient,
	}
}

type lomsService struct {
	orderRepository     orderRepository
	stocksRepository    stocksRepository
	transactor          transactor
	notificationsClient notificationsClient
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
	var created entity.Order
	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		if err := l.stocksRepository.ReserveItems(ctx, items); err != nil {
			return err
		}
		var err error
		created, err = l.orderRepository.CreateOrder(ctx, entity.Order{
			UserID: userID,
			Items:  items,
			Status: entity.OrderStatusAwaitingPayment,
		})
		return err
	})
	if err != nil {
		return 0, err
	}
	err = l.notificationsClient.SendOrderStatusChangedNotification(ctx, userID, created.ID, created.Status)
	if err != nil {
		return 0, err
	}

	return created.ID, nil
}

func (l *lomsService) GetOrder(ctx context.Context, id int64) (entity.Order, error) {
	order, err := l.orderRepository.GetOrder(ctx, id)
	if err != nil {
		return entity.Order{}, err
	}
	return order, nil
}

func (l *lomsService) PayOrder(ctx context.Context, id int64) error {
	return l.transactor.WithTx(ctx, func(ctx context.Context) error {
		order, err := l.orderRepository.GetOrderForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if order.Status != entity.OrderStatusAwaitingPayment {
			return ErrInvalidStatus
		}

		err = l.orderRepository.SetOrderStatus(ctx, id, entity.OrderStatusPaid)
		if err != nil {
			return err
		}
		return l.notificationsClient.SendOrderStatusChangedNotification(ctx, order.UserID, order.ID, entity.OrderStatusPaid)
	})
}
func (l *lomsService) CancelOrder(ctx context.Context, id int64) error {
	return l.transactor.WithTx(ctx, func(ctx context.Context) error {
		order, err := l.orderRepository.GetOrderForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if order.Status != entity.OrderStatusAwaitingPayment {
			return ErrInvalidStatus
		}
		if err := l.stocksRepository.ReleaseItems(ctx, order.Items); err != nil {
			return err
		}
		err = l.orderRepository.SetOrderStatus(ctx, id, entity.OrderStatusCancelled)
		if err != nil {
			return err
		}
		return l.notificationsClient.SendOrderStatusChangedNotification(ctx, order.UserID, order.ID, entity.OrderStatusCancelled)

	})
}
