package converter

import (
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	itemusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {
	switch {
	case errors.Is(err, entity.ErrCartItemNotFound):
		return status.Error(codes.NotFound, "cart item not found")
	case errors.Is(err, port.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, entity.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, "insufficient stock")
	case errors.Is(err, itemusecase.ErrZeroCount):
		return status.Error(codes.InvalidArgument, "count must be greater than zero")
	case errors.Is(err, port.ErrInvalidOrder):
		return status.Error(codes.InvalidArgument, "invalid order")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func FromItems(items []port.Item) []*cartpb.Item {
	result := make([]*cartpb.Item, len(items))
	for i, item := range items {
		result[i] = &cartpb.Item{Sku: item.SKU, Count: item.Count, Name: item.Name, Price: item.Price}
	}
	return result
}
