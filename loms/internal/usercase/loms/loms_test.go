package loms

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)

	ctx := context.Background()
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	wantOrderID := int64(42)
	wantOrder := entity.Order{
		ID:     wantOrderID,
		UserID: 7,
		Items:  items,
		Status: entity.OrderStatusAwaitingPayment,
	}

	reserveCall := stocks.EXPECT().ReserveItems(ctx, items).Return(nil)
	createOrderCall := orders.EXPECT().CreateOrder(ctx, entity.Order{
		UserID: 7,
		Items:  items,
		Status: entity.OrderStatusAwaitingPayment,
	}).Return(wantOrder, nil)
	gomock.InOrder(reserveCall, createOrderCall)

	gotOrderID, err := service.CreateOrder(ctx, 7, items)
	if err != nil {
		t.Fatalf("CreateOrder() returned unexpected error: %v", err)
	}
	if gotOrderID != wantOrderID {
		t.Fatalf("CreateOrder() order ID = %d, want %d", gotOrderID, wantOrderID)
	}
}

func TestCreateOrder_RejectsInvalidItems(t *testing.T) {
	tests := []struct {
		name  string
		items []entity.OrderItem
	}{
		{name: "empty items"},
		{name: "zero count", items: []entity.OrderItem{{SKU: 123, Count: 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orders := mocks.NewMockorderRepository(ctrl)
			stocks := mocks.NewMockstocksRepository(ctrl)
			service := NewLomsService(orders, stocks)

			_, err := service.CreateOrder(context.Background(), 7, tt.items)
			if !errors.Is(err, ErrWrongOrder) {
				t.Fatalf("CreateOrder() error = %v, want %v", err, ErrWrongOrder)
			}
		})
	}
}

func TestCreateOrder_ReserveFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)

	wantErr := errors.New("not enough stock")
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	stocks.EXPECT().ReserveItems(gomock.Any(), items).Return(wantErr)

	id, err := service.CreateOrder(context.Background(), 7, items)
	if id != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("CreateOrder() = (%d, %v), want (0, %v)", id, err, wantErr)
	}
}

func TestCreateOrder_OrderFailureReleasesStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)

	ctx := context.Background()
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	createErr := errors.New("storage failure")
	stocks.EXPECT().ReserveItems(ctx, items).Return(nil)
	orders.EXPECT().CreateOrder(ctx, gomock.Any()).Return(entity.Order{}, createErr)
	stocks.EXPECT().ReleaseItems(gomock.Any(), items).Return(nil)

	id, err := service.CreateOrder(ctx, 7, items)
	if id != 0 || !errors.Is(err, createErr) {
		t.Fatalf("CreateOrder() = (%d, %v), want (0, %v)", id, err, createErr)
	}
}

func TestCreateOrder_OrderAndReleaseFailureJoinsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)

	ctx := context.Background()
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	createErr := errors.New("storage failure")
	releaseErr := errors.New("release failure")
	stocks.EXPECT().ReserveItems(ctx, items).Return(nil)
	orders.EXPECT().CreateOrder(ctx, gomock.Any()).Return(entity.Order{}, createErr)
	stocks.EXPECT().ReleaseItems(gomock.Any(), items).Return(releaseErr)

	id, err := service.CreateOrder(ctx, 7, items)
	if id != 0 || !errors.Is(err, createErr) || !errors.Is(err, releaseErr) {
		t.Fatalf("CreateOrder() = (%d, %v), want joined create and release errors", id, err)
	}
}

func TestGetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl))

	ctx := context.Background()
	want := entity.Order{ID: 42, UserID: 7, Status: entity.OrderStatusAwaitingPayment}
	orders.EXPECT().GetOrder(ctx, int64(42)).Return(want, nil)

	got, err := service.GetOrder(ctx, 42)
	if err != nil || got.ID != want.ID || got.UserID != want.UserID || got.Status != want.Status {
		t.Fatalf("GetOrder() = (%+v, %v), want (%+v, nil)", got, err, want)
	}
}

func TestGetOrder_ReturnsRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl))
	wantErr := entity.ErrOrderNotFound
	orders.EXPECT().GetOrder(gomock.Any(), int64(404)).Return(entity.Order{}, wantErr)

	_, err := service.GetOrder(context.Background(), 404)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetOrder() error = %v, want %v", err, wantErr)
	}
}

func TestPayOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl))
	ctx := context.Background()
	orders.EXPECT().TransitionOrderStatus(ctx, int64(42), entity.OrderStatusAwaitingPayment, entity.OrderStatusPaid).Return(nil)

	if err := service.PayOrder(ctx, 42); err != nil {
		t.Fatalf("PayOrder() error = %v, want nil", err)
	}
}

func TestPayOrder_ReturnsRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl))
	wantErr := entity.ErrInvalidStatus
	orders.EXPECT().TransitionOrderStatus(gomock.Any(), int64(42), entity.OrderStatusAwaitingPayment, entity.OrderStatusPaid).Return(wantErr)

	err := service.PayOrder(context.Background(), 42)
	if !errors.Is(err, wantErr) {
		t.Fatalf("PayOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)
	ctx := context.Background()
	order := entity.Order{ID: 42, Items: []entity.OrderItem{{SKU: 123, Count: 2}}}

	orders.EXPECT().GetOrder(ctx, int64(42)).Return(order, nil)
	orders.EXPECT().TransitionOrderStatus(ctx, int64(42), entity.OrderStatusAwaitingPayment, entity.OrderStatusCancelled).Return(nil)
	stocks.EXPECT().ReleaseItems(gomock.Any(), order.Items).Return(nil)

	if err := service.CancelOrder(ctx, 42); err != nil {
		t.Fatalf("CancelOrder() error = %v, want nil", err)
	}
}

func TestCancelOrder_GetOrderFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl))
	wantErr := entity.ErrOrderNotFound
	orders.EXPECT().GetOrder(gomock.Any(), int64(404)).Return(entity.Order{}, wantErr)

	err := service.CancelOrder(context.Background(), 404)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelOrder_TransitionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)
	ctx := context.Background()
	order := entity.Order{ID: 42}
	wantErr := entity.ErrInvalidStatus
	orders.EXPECT().GetOrder(ctx, int64(42)).Return(order, nil)
	orders.EXPECT().TransitionOrderStatus(ctx, int64(42), entity.OrderStatusAwaitingPayment, entity.OrderStatusCancelled).Return(wantErr)

	err := service.CancelOrder(ctx, 42)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelOrder_ReleaseFailureRollsStatusBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)
	ctx := context.Background()
	order := entity.Order{ID: 42, Items: []entity.OrderItem{{SKU: 123, Count: 2}}}
	releaseErr := errors.New("release failure")
	orders.EXPECT().GetOrder(ctx, int64(42)).Return(order, nil)
	orders.EXPECT().TransitionOrderStatus(ctx, int64(42), entity.OrderStatusAwaitingPayment, entity.OrderStatusCancelled).Return(nil)
	stocks.EXPECT().ReleaseItems(gomock.Any(), order.Items).Return(releaseErr)
	orders.EXPECT().TransitionOrderStatus(gomock.Any(), int64(42), entity.OrderStatusCancelled, entity.OrderStatusAwaitingPayment).Return(nil)

	err := service.CancelOrder(ctx, 42)
	if !errors.Is(err, releaseErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, releaseErr)
	}
}

func TestCancelOrder_ReleaseAndRollbackFailureJoinsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks)
	ctx := context.Background()
	order := entity.Order{ID: 42, Items: []entity.OrderItem{{SKU: 123, Count: 2}}}
	releaseErr := errors.New("release failure")
	rollbackErr := errors.New("rollback failure")
	orders.EXPECT().GetOrder(ctx, int64(42)).Return(order, nil)
	orders.EXPECT().TransitionOrderStatus(ctx, int64(42), entity.OrderStatusAwaitingPayment, entity.OrderStatusCancelled).Return(nil)
	stocks.EXPECT().ReleaseItems(gomock.Any(), order.Items).Return(releaseErr)
	orders.EXPECT().TransitionOrderStatus(gomock.Any(), int64(42), entity.OrderStatusCancelled, entity.OrderStatusAwaitingPayment).Return(rollbackErr)

	err := service.CancelOrder(ctx, 42)
	if !errors.Is(err, releaseErr) || !errors.Is(err, rollbackErr) {
		t.Fatalf("CancelOrder() error = %v, want joined release and rollback errors", err)
	}
}
