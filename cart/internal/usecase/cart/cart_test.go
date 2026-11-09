package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart/mocks"
	"go.uber.org/mock/gomock"
)

type testTransactor struct{}

func (testTransactor) WithTx(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
}

func TestListCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	service := NewCartService(repo, products, mocks.NewMocklomsClient(ctrl), testTransactor{})
	ctx := context.Background()
	stored := []entity.CartItem{{SKU: 123, Count: 2}}
	wantItems := []port.Item{{SKU: 123, Count: 2, Name: "New name", Price: 150}}
	repo.EXPECT().GetItemsByUserID(ctx, int64(7)).Return(stored, nil)
	products.EXPECT().GetProduct(ctx, uint32(123)).Return(port.ProductInfo{Name: "New name", Price: 150}, nil)

	items, total, err := service.ListCart(ctx, 7)
	if err != nil {
		t.Fatalf("ListCart() error = %v, want nil", err)
	}
	if len(items) != 1 || items[0] != wantItems[0] {
		t.Fatalf("ListCart() items = %+v, want %+v", items, wantItems)
	}
	if total != 300 {
		t.Fatalf("ListCart() total = %d, want 300", total)
	}
}

func TestListCart_RepositoryReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	service := NewCartService(repo, mocks.NewMockproductClient(ctrl), mocks.NewMocklomsClient(ctrl), testTransactor{})
	wantErr := errors.New("cart storage failed")
	repo.EXPECT().GetItemsByUserID(gomock.Any(), int64(7)).Return(nil, wantErr)

	_, _, err := service.ListCart(context.Background(), 7)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListCart() error = %v, want %v", err, wantErr)
	}
}

func TestListCart_ProductError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	service := NewCartService(repo, products, mocks.NewMocklomsClient(ctrl), testTransactor{})
	wantErr := port.ErrProductNotFound
	repo.EXPECT().GetItemsByUserID(gomock.Any(), int64(7)).Return([]entity.CartItem{{SKU: 123, Count: 1}}, nil)
	products.EXPECT().GetProduct(gomock.Any(), uint32(123)).Return(port.ProductInfo{}, wantErr)

	_, _, err := service.ListCart(context.Background(), 7)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListCart() error = %v, want %v", err, wantErr)
	}
}

func TestCheckoutCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	loms := mocks.NewMocklomsClient(ctrl)
	service := NewCartService(repo, products, loms, testTransactor{})
	ctx := context.Background()
	stored := []entity.CartItem{{SKU: 123, Count: 2}}
	orderItems := []port.OrderItem{{SKU: 123, Count: 2}}
	repo.EXPECT().GetItemsByUserID(ctx, int64(7)).Return(stored, nil)
	products.EXPECT().GetProduct(ctx, uint32(123)).Return(port.ProductInfo{Name: "Shoes", Price: 500}, nil)
	loms.EXPECT().CreateOrder(ctx, int64(7), orderItems).Return(int64(42), nil)
	repo.EXPECT().DeleteItemsByUserID(ctx, int64(7)).Return(nil)

	id, err := service.CheckoutCart(ctx, 7)
	if err != nil || id != 42 {
		t.Fatalf("CheckoutCart() = (%d, %v), want (42, nil)", id, err)
	}
}

func TestCheckoutCart_OrderCreationErrorLeavesCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	loms := mocks.NewMocklomsClient(ctrl)
	service := NewCartService(repo, products, loms, testTransactor{})
	ctx := context.Background()
	items := []entity.CartItem{{SKU: 123, Count: 1}}
	wantErr := errors.New("LOMS unavailable")
	repo.EXPECT().GetItemsByUserID(ctx, int64(7)).Return(items, nil)
	products.EXPECT().GetProduct(ctx, uint32(123)).Return(port.ProductInfo{Name: "Shoes", Price: 500}, nil)
	loms.EXPECT().CreateOrder(ctx, int64(7), []port.OrderItem{{SKU: 123, Count: 1}}).Return(int64(0), wantErr)

	_, err := service.CheckoutCart(ctx, 7)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CheckoutCart() error = %v, want %v", err, wantErr)
	}
}

func TestCheckoutCart_ClearError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	products := mocks.NewMockproductClient(ctrl)
	loms := mocks.NewMocklomsClient(ctrl)
	service := NewCartService(repo, products, loms, testTransactor{})
	ctx := context.Background()
	items := []entity.CartItem{{SKU: 123, Count: 1}}
	wantErr := errors.New("cart clear failed")
	repo.EXPECT().GetItemsByUserID(ctx, int64(7)).Return(items, nil)
	products.EXPECT().GetProduct(ctx, uint32(123)).Return(port.ProductInfo{Name: "Shoes", Price: 500}, nil)
	loms.EXPECT().CreateOrder(ctx, int64(7), []port.OrderItem{{SKU: 123, Count: 1}}).Return(int64(42), nil)
	repo.EXPECT().DeleteItemsByUserID(ctx, int64(7)).Return(wantErr)

	_, err := service.CheckoutCart(ctx, 7)
	if !errors.Is(err, wantErr) {
		t.Fatalf("CheckoutCart() error = %v, want %v", err, wantErr)
	}
}

func TestClearCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	service := NewCartService(repo, mocks.NewMockproductClient(ctrl), mocks.NewMocklomsClient(ctrl), testTransactor{})
	repo.EXPECT().DeleteItemsByUserID(gomock.Any(), int64(7)).Return(nil)

	err := service.ClearCart(context.Background(), 7)
	if err != nil {
		t.Fatalf("ClearCart() error = %v, want nil", err)
	}
}

func TestClearCart_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockcartRepository(ctrl)
	service := NewCartService(repo, mocks.NewMockproductClient(ctrl), mocks.NewMocklomsClient(ctrl), testTransactor{})
	wantErr := errors.New("delete failed")
	repo.EXPECT().DeleteItemsByUserID(gomock.Any(), int64(7)).Return(wantErr)

	err := service.ClearCart(context.Background(), 7)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ClearCart() error = %v, want %v", err, wantErr)
	}
}
