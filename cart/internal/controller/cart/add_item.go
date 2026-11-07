package cart

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/converter"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *cartServer) AddItem(ctx context.Context, req *cartpb.AddItemRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	err := c.itemService.AddItem(ctx, req.GetUserId(), req.GetSku(), req.GetCount())
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
