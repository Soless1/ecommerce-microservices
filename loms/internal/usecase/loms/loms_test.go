package loms

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms/mocks"
	"go.uber.org/mock/gomock"
)

type testTransactor struct {
	commitErr error
}

func (t testTransactor) WithTx(ctx context.Context, f func(context.Context) error) error {
	if err := f(ctx); err != nil {
		return err
	}
	return t.commitErr
}

func TestCreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks, testTransactor{})

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
			service := NewLomsService(orders, stocks, testTransactor{})

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
	service := NewLomsService(orders, stocks, testTransactor{})

	wantErr := errors.New("not enough stock")
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	stocks.EXPECT().ReserveItems(gomock.Any(), items).Return(wantErr)

	id, err := service.CreateOrder(context.Background(), 7, items)
	if id != 0 || !errors.Is(err, wantErr) {
		t.Fatalf("CreateOrder() = (%d, %v), want (0, %v)", id, err, wantErr)
	}
}

func TestCreateOrder_OrderFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks, testTransactor{})

	ctx := context.Background()
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	createErr := errors.New("storage failure")
	stocks.EXPECT().ReserveItems(ctx, items).Return(nil)
	orders.EXPECT().CreateOrder(ctx, gomock.Any()).Return(entity.Order{}, createErr)

	id, err := service.CreateOrder(ctx, 7, items)
	if id != 0 || !errors.Is(err, createErr) {
		t.Fatalf("CreateOrder() = (%d, %v), want (0, %v)", id, err, createErr)
	}
}

func TestCreateOrder_CommitFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	commitErr := errors.New("commit failure")
	service := NewLomsService(orders, stocks, testTransactor{commitErr: commitErr})

	ctx := context.Background()
	items := []entity.OrderItem{{SKU: 123, Count: 2}}
	stocks.EXPECT().ReserveItems(ctx, items).Return(nil)
	orders.EXPECT().CreateOrder(ctx, gomock.Any()).Return(entity.Order{ID: 42}, nil)

	id, err := service.CreateOrder(ctx, 7, items)
	if id != 0 || !errors.Is(err, commitErr) {
		t.Fatalf("CreateOrder() = (%d, %v), want (0, %v)", id, err, commitErr)
	}
}

func TestGetOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl), testTransactor{})

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
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl), testTransactor{})
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
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl), testTransactor{})
	ctx := context.Background()
	orders.EXPECT().GetOrderForUpdate(ctx, int64(42)).Return(entity.Order{Status: entity.OrderStatusAwaitingPayment}, nil)
	orders.EXPECT().SetOrderStatus(ctx, int64(42), entity.OrderStatusPaid).Return(nil)

	if err := service.PayOrder(ctx, 42); err != nil {
		t.Fatalf("PayOrder() error = %v, want nil", err)
	}
}

func TestPayOrder_ReturnsRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl), testTransactor{})
	wantErr := entity.ErrInvalidStatus
	orders.EXPECT().GetOrderForUpdate(gomock.Any(), int64(42)).Return(entity.Order{Status: entity.OrderStatusCancelled}, nil)

	err := service.PayOrder(context.Background(), 42)
	if !errors.Is(err, wantErr) {
		t.Fatalf("PayOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks, testTransactor{})
	ctx := context.Background()
	order := entity.Order{ID: 42, Status: entity.OrderStatusAwaitingPayment, Items: []entity.OrderItem{{SKU: 123, Count: 2}}}

	orders.EXPECT().GetOrderForUpdate(ctx, int64(42)).Return(order, nil)
	stocks.EXPECT().ReleaseItems(gomock.Any(), order.Items).Return(nil)
	orders.EXPECT().SetOrderStatus(ctx, int64(42), entity.OrderStatusCancelled).Return(nil)

	if err := service.CancelOrder(ctx, 42); err != nil {
		t.Fatalf("CancelOrder() error = %v, want nil", err)
	}
}

func TestCancelOrder_GetOrderFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	service := NewLomsService(orders, mocks.NewMockstocksRepository(ctrl), testTransactor{})
	wantErr := entity.ErrOrderNotFound
	orders.EXPECT().GetOrderForUpdate(gomock.Any(), int64(404)).Return(entity.Order{}, wantErr)

	err := service.CancelOrder(context.Background(), 404)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelOrder_InvalidStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks, testTransactor{})
	ctx := context.Background()
	order := entity.Order{ID: 42, Status: entity.OrderStatusPaid}
	wantErr := entity.ErrInvalidStatus
	orders.EXPECT().GetOrderForUpdate(ctx, int64(42)).Return(order, nil)

	err := service.CancelOrder(ctx, 42)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelOrder_ReleaseFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	service := NewLomsService(orders, stocks, testTransactor{})
	ctx := context.Background()
	order := entity.Order{ID: 42, Status: entity.OrderStatusAwaitingPayment, Items: []entity.OrderItem{{SKU: 123, Count: 2}}}
	releaseErr := errors.New("release failure")
	orders.EXPECT().GetOrderForUpdate(ctx, int64(42)).Return(order, nil)
	stocks.EXPECT().ReleaseItems(gomock.Any(), order.Items).Return(releaseErr)

	err := service.CancelOrder(ctx, 42)
	if !errors.Is(err, releaseErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, releaseErr)
	}
}

func TestCancelOrder_CommitFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	orders := mocks.NewMockorderRepository(ctrl)
	stocks := mocks.NewMockstocksRepository(ctrl)
	commitErr := errors.New("commit failure")
	service := NewLomsService(orders, stocks, testTransactor{commitErr: commitErr})
	ctx := context.Background()
	order := entity.Order{ID: 42, Status: entity.OrderStatusAwaitingPayment, Items: []entity.OrderItem{{SKU: 123, Count: 2}}}
	orders.EXPECT().GetOrderForUpdate(ctx, int64(42)).Return(order, nil)
	stocks.EXPECT().ReleaseItems(ctx, order.Items).Return(nil)
	orders.EXPECT().SetOrderStatus(ctx, int64(42), entity.OrderStatusCancelled).Return(nil)

	err := service.CancelOrder(ctx, 42)
	if !errors.Is(err, commitErr) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, commitErr)
	}
}
