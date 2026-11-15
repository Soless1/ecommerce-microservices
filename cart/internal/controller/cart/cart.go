package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
)

// TODO: mocks

type (
	ItemService interface {
		AddItem(ctx context.Context, userID int64, sku uint32, count uint32) error
		DeleteItem(ctx context.Context, id int64, sku uint32) error
	}

	Service interface {
		ListCart(ctx context.Context, userID int64) ([]port.Item, uint32, error)
		CheckoutCart(ctx context.Context, id int64) (int64, error)
		ClearCart(ctx context.Context, id int64) error
	}
)

type cartServer struct {
	itemService ItemService
	cartService Service
}

func NewCartServer(
	itemService ItemService,
	cartService Service,
) *cartServer {
	return &cartServer{
		itemService: itemService,
		cartService: cartService,
	}
}

// type CartServer interface {
//	AddItem(context.Context, *AddItemRequest) (*emptypb.Empty, error)
//	DeleteItem(context.Context, *DeleteItemRequest) (*emptypb.Empty, error)
//	ListCart(*ListCartRequest, Cart_ListCartServer) error
//	ClearCart(context.Context, *ClearCartRequest) (*emptypb.Empty, error)
//	CheckoutCart(context.Context, *CheckoutCartRequest) (*CheckoutCartResponse, error)
// }
