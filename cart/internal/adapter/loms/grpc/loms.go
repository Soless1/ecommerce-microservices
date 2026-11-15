package grpc

import (
	"context"

	adapterconverter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
)

type lomsClient struct {
	client lomspb.LomsClient
}

func NewLOMSClient(client lomspb.LomsClient) *lomsClient {
	return &lomsClient{client: client}
}

func (c *lomsClient) CreateOrder(ctx context.Context, userID int64, items []port.OrderItem) (int64, error) {
	response, err := c.client.CreateOrder(ctx, &lomspb.CreateOrderRequest{UserId: userID, Items: port.FromOrderItems(items)})
	if err != nil {
		return 0, adapterconverter.FromLOMSError(err)
	}
	return response.GetOrderId(), nil
}

// GetOrder(ctx context.Context, in *GetOrderRequest) (*GetOrderResponse, error)
// PayOrder(ctx context.Context, in *PayOrderRequest ) (*emptypb.Empty, error)
// CancelOrder(ctx context.Context, in *CancelOrderRequest) (*emptypb.Empty, error)
