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
	}

	productClient interface {
		GetProduct(ctx context.Context, sku uint32) (port.ProductInfo, error)
	}

	lomsClient interface {
		CreateOrder(ctx context.Context, userId int64, items []port.OrderItem) (int64, error)
	}
	transactor interface {
		WithTx(ctx context.Context, f func(context.Context) error) error
	}
)

//go:generate mockgen -source=cart.go -destination=mocks/cart_mock.go -package=mocks

type cartService struct {
	cartRepository cartRepository
	productClient  productClient
	lomsClient     lomsClient
	transactor     transactor
}

func NewCartService(cartRepository cartRepository, productClient productClient, lomsClient lomsClient, transactor transactor) *cartService {
	return &cartService{
		cartRepository: cartRepository,
		productClient:  productClient,
		lomsClient:     lomsClient,
		transactor:     transactor,
	}
}

func (s *cartService) ListCart(ctx context.Context, userId int64) ([]port.Item, uint32, error) {
	items, err := s.cartRepository.GetItemsByUserID(ctx, userId)
	if err != nil {
		return nil, 0, err
	}
	totalPrice := uint32(0)
	result := make([]port.Item, 0, len(items))
	for _, item := range items {
		product, err2 := s.productClient.GetProduct(ctx, item.SKU)
		if err2 != nil {
			return nil, 0, err2
		}
		totalPrice += product.Price * item.Count
		result = append(result, port.Item{
			SKU:   item.SKU,
			Count: item.Count,
			Name:  product.Name,
			Price: product.Price,
		})
	}
	return result, totalPrice, nil
}
func (s *cartService) CheckoutCart(ctx context.Context, id int64) (int64, error) {
	var orderID int64
	err := s.transactor.WithTx(ctx, func(ctx context.Context) error {
		cart, _, err := s.ListCart(ctx, id)
		if err != nil {
			return err
		}
		orderItems := make([]port.OrderItem, 0, len(cart))
		for _, item := range cart {
			orderItems = append(orderItems, port.OrderItem{SKU: item.SKU, Count: item.Count})
		}
		orderID, err = s.lomsClient.CreateOrder(ctx, id, orderItems)
		if err != nil {
			return err
		}
		return s.cartRepository.DeleteItemsByUserID(ctx, id)
	})
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func (s *cartService) ClearCart(ctx context.Context, id int64) error {
	return s.cartRepository.DeleteItemsByUserID(ctx, id)
}
