package item

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
)

type (
	cartRepository interface {
		AddItem(ctx context.Context, userId int64, item entity.CartItem) error
		DeleteItem(ctx context.Context, userId int64, sku uint32) error
	}

	productClient interface {
		GetProduct(ctx context.Context, sku uint32) (port.ProductInfo, error)
	}

	stocksClient interface {
		GetStock(ctx context.Context, sku uint32) (uint64, error)
		SetStock(ctx context.Context, sku uint32, count uint64) error
	}
)

//go:generate mockgen -source=item.go -destination=mocks/item_mock.go -package=mocks

var (
	ErrInsufficientStock = entity.ErrInsufficientStock
	ErrZeroCount         = errors.New("zero items adding")
)

type itemService struct {
	cartRepository cartRepository
	productClient  productClient
	stocksClient   stocksClient
}

func NewItemService(cartRepository cartRepository, productClient productClient, stocksClient stocksClient) *itemService {
	return &itemService{
		cartRepository: cartRepository,
		productClient:  productClient,
		stocksClient:   stocksClient,
	}
}

func (s *itemService) AddItem(ctx context.Context, userId int64, Sku uint32, count uint32) error {
	if count == 0 {
		return ErrZeroCount
	}
	_, err := s.productClient.GetProduct(ctx, Sku)
	if err != nil {
		return err
	}

	stock, err := s.stocksClient.GetStock(ctx, Sku)
	if err != nil {
		return err
	}
	if stock < uint64(count) {
		return ErrInsufficientStock
	}
	return s.cartRepository.AddItem(ctx, userId, entity.CartItem{SKU: Sku, Count: count})
}

func (s *itemService) DeleteItem(ctx context.Context, id int64, sku uint32) error {

	err := s.cartRepository.DeleteItem(ctx, id, sku)
	if err != nil {
		return err
	}
	return nil
}
