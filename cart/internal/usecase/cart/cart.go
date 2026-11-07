package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
)

type (
	cartRepository interface {
		GetItemsByUserID(ctx context.Context, userId int64) ([]entity.CartItem, error)
		DeleteItemsByUserID(ctx context.Context, userId int64) error
		SetItemsByUserID(ctx context.Context, userId int64, items []entity.CartItem) error
	}

	productClient interface {
		GetProduct(ctx context.Context, sku uint32) (port.ProductInfo, error)
	}

	lomsClient interface {
		CreateOrder(ctx context.Context, userId int64, items []port.OrderItem) (int64, error)
	}
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mock.go -package=mocks

type cartService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
}

func NewCartService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient) *cartService {
	return &cartService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
	}
}

func (s *cartService) ListCart(ctx context.Context, userId int64) ([]entity.CartItem, uint32, error) {
	items, err := s.cartRepository.GetItemsByUserID(ctx, userId)
	if err != nil {
		return nil, 0, err
	}
	totalPrice := uint32(0)
	for i, item := range items {
		product, err2 := s.productClient.GetProduct(ctx, item.SKU)
		if err2 != nil {
			return nil, 0, err2
		}
		items[i].Name = product.Name
		items[i].Price = product.Price
		totalPrice += product.Price * item.Count
	}
	if err := s.cartRepository.SetItemsByUserID(ctx, userId, items); err != nil {
		return nil, 0, err
	}
	return items, totalPrice, nil
}
func (s *cartService) CheckoutCart(ctx context.Context, id int64) (int64, error) {
	cart, _, err := s.ListCart(ctx, id)
	if err != nil {
		return 0, err
	}
	orderItems := make([]port.OrderItem, 0, len(cart))
	for _, item := range cart {
		orderItems = append(orderItems, port.OrderItem{SKU: item.SKU, Count: item.Count})
	}
	orderId, err := s.lomsClient.CreateOrder(ctx, id, orderItems)
	if err != nil {
		return 0, err
	}
	err = s.cartRepository.DeleteItemsByUserID(ctx, id)
	if err != nil {
		return 0, err
	}
	return orderId, nil
}

func (s *cartService) ClearCart(ctx context.Context, id int64) error {
	return s.cartRepository.DeleteItemsByUserID(ctx, id)
}
