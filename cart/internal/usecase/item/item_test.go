package item

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item/mocks"
	"go.uber.org/mock/gomock"
)

func TestAddItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	stocks := mocks.NewMockstocksClient(ctrl)
	service := NewItemService(repo, products, stocks)
	ctx := context.Background()

	products.EXPECT().GetProduct(ctx, uint32(123)).Return(port.ProductInfo{Name: "Shoes", Price: 500}, nil)
	stocks.EXPECT().GetStock(ctx, uint32(123)).Return(uint64(4), nil)
	repo.EXPECT().AddItem(ctx, int64(7), entity.CartItem{SKU: 123, Count: 2, Name: "Shoes", Price: 500}).Return(nil)

	if err := service.AddItem(ctx, 7, 123, 2); err != nil {
		t.Fatalf("AddItem() error = %v, want nil", err)
	}
}

func TestAddItem_ZeroCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := NewItemService(mocks.NewMockcartRepository(ctrl), mocks.NewMockproductClient(ctrl), mocks.NewMockstocksClient(ctrl))

	err := service.AddItem(context.Background(), 7, 123, 0)
	if !errors.Is(err, ErrZeroCount) {
		t.Fatalf("AddItem() error = %v, want %v", err, ErrZeroCount)
	}
}

func TestAddItem_ProductError(t *testing.T) {
	ctrl := gomock.NewController(t)
	products := mocks.NewMockproductClient(ctrl)
	service := NewItemService(mocks.NewMockcartRepository(ctrl), products, mocks.NewMockstocksClient(ctrl))
	wantErr := port.ErrProductNotFound
	products.EXPECT().GetProduct(gomock.Any(), uint32(123)).Return(port.ProductInfo{}, wantErr)

	err := service.AddItem(context.Background(), 7, 123, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("AddItem() error = %v, want %v", err, wantErr)
	}
}

func TestAddItem_StockClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	products := mocks.NewMockproductClient(ctrl)
	stocks := mocks.NewMockstocksClient(ctrl)
	service := NewItemService(mocks.NewMockcartRepository(ctrl), products, stocks)
	wantErr := errors.New("stocks unavailable")
	products.EXPECT().GetProduct(gomock.Any(), uint32(123)).Return(port.ProductInfo{}, nil)
	stocks.EXPECT().GetStock(gomock.Any(), uint32(123)).Return(uint64(0), wantErr)

	err := service.AddItem(context.Background(), 7, 123, 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("AddItem() error = %v, want %v", err, wantErr)
	}
}

func TestAddItem_InsufficientStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	products := mocks.NewMockproductClient(ctrl)
	stocks := mocks.NewMockstocksClient(ctrl)
	service := NewItemService(mocks.NewMockcartRepository(ctrl), products, stocks)
	products.EXPECT().GetProduct(gomock.Any(), uint32(123)).Return(port.ProductInfo{}, nil)
	stocks.EXPECT().GetStock(gomock.Any(), uint32(123)).Return(uint64(1), nil)

	err := service.AddItem(context.Background(), 7, 123, 2)
	if !errors.Is(err, entity.ErrInsufficientStock) {
		t.Fatalf("AddItem() error = %v, want %v", err, entity.ErrInsufficientStock)
	}
}

func TestAddItem_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	stocks := mocks.NewMockstocksClient(ctrl)
	service := NewItemService(repo, products, stocks)
	wantErr := errors.New("cart storage failed")
	products.EXPECT().GetProduct(gomock.Any(), uint32(123)).Return(port.ProductInfo{}, nil)
	stocks.EXPECT().GetStock(gomock.Any(), uint32(123)).Return(uint64(3), nil)
	repo.EXPECT().AddItem(gomock.Any(), int64(7), gomock.Any()).Return(wantErr)

	err := service.AddItem(context.Background(), 7, 123, 2)
	if !errors.Is(err, wantErr) {
		t.Fatalf("AddItem() error = %v, want %v", err, wantErr)
	}
}

func TestDeleteItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	service := NewItemService(repo, mocks.NewMockproductClient(ctrl), mocks.NewMockstocksClient(ctrl))
	repo.EXPECT().DeleteItem(gomock.Any(), int64(7), uint32(123)).Return(nil)

	if err := service.DeleteItem(context.Background(), 7, 123); err != nil {
		t.Fatalf("DeleteItem() error = %v, want nil", err)
	}
}

func TestDeleteItem_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	service := NewItemService(repo, mocks.NewMockproductClient(ctrl), mocks.NewMockstocksClient(ctrl))
	wantErr := entity.ErrCartItemNotFound
	repo.EXPECT().DeleteItem(gomock.Any(), int64(7), uint32(123)).Return(wantErr)

	err := service.DeleteItem(context.Background(), 7, 123)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DeleteItem() error = %v, want %v", err, wantErr)
	}
}
