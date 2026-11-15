package cart

import (
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/converter"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *cartServer) ListCart(req *cartpb.ListCartRequest, listCartServer cartpb.Cart_ListCartServer) error {
	if err := req.Validate(); err != nil {
		return status.Errorf(codes.InvalidArgument, "%v", err)
	}
	userID := req.GetUserId()
	list, totalPrice, err := c.cartService.ListCart(listCartServer.Context(), userID)
	if len(list) == 0 {
		return nil
	}
	if err != nil {
		return converter.ToGRPCError(err)
	}
	items := converter.FromItems(list)
	return listCartServer.Send(&cartpb.ListCartResponse{Items: items, TotalPrice: totalPrice})
}
