package loms

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/adapter/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox"
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
		SendOrderStatusChangedNotification(ctx context.Context, body port.OrderStatusChangedNotification) error
	}
	outboxRepository interface {
		SendMessage(
			ctx context.Context,
			idempotencyKey string,
			kind outboxrepo.Kind,
			message []byte,
		) error
	}
)

//go:generate mockgen -source=loms.go -destination=mocks/loms_mock.go -package=mocks

func NewLomsService(
	orderRepository orderRepository,
	stocksRepository stocksRepository,
	transactor transactor,
	notificationsClient notificationsClient,
	outboxRepository outboxRepository,
) *lomsService {
	return &lomsService{
		orderRepository:     orderRepository,
		stocksRepository:    stocksRepository,
		transactor:          transactor,
		notificationsClient: notificationsClient,
		outboxRepository:    outboxRepository,
	}
}

type lomsService struct {
	orderRepository     orderRepository
	stocksRepository    stocksRepository
	transactor          transactor
	notificationsClient notificationsClient
	outboxRepository    outboxRepository
}

func (s *lomsService) createIdempotencyKey(orderID int64, status entity.OrderStatus) string {
	return strconv.FormatInt(orderID, 10) + "_" + string(status)
}

func (s *lomsService) createOutboxOrderStatusChangedNotification(ctx context.Context, order *entity.Order) error {
	key := s.createIdempotencyKey(order.ID, order.Status)
	body := port.OrderStatusChangedNotification{
		UserID:  order.UserID,
		OrderID: order.ID,
		Status:  converter.FromOrderStatus(order.Status),
	}

	rawBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return s.outboxRepository.SendMessage(
		ctx,
		key,
		outboxrepo.KindNotification,
		rawBody,
	)
}

func (s *lomsService) OrderStatusChangedNotificationKindHandler(ctx context.Context, data []byte) error {
	var body port.OrderStatusChangedNotification
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}
	return s.notificationsClient.SendOrderStatusChangedNotification(ctx, body)
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
		if err != nil {
			return err
		}
		return l.createOutboxOrderStatusChangedNotification(ctx, &created)
	})
	if err != nil {
		return 0, err
	}

	return created.ID, err
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
		order.Status = entity.OrderStatusPaid

		return l.createOutboxOrderStatusChangedNotification(ctx, &order)
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
		order.Status = entity.OrderStatusCancelled

		return l.createOutboxOrderStatusChangedNotification(ctx, &order)
	})
}
