package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/converter"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *cartServer) CheckoutCart(ctx context.Context, req *cartpb.CheckoutCartRequest) (*cartpb.CheckoutCartResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	UserId := req.GetUserId()

	orderId, err := c.cartService.CheckoutCart(ctx, UserId)
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &cartpb.CheckoutCartResponse{OrderId: orderId}, nil
}
