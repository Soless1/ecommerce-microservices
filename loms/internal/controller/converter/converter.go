package converter

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"errors"
)

func ToOrderStatus(status lomspb.OrderStatus) entity.OrderStatus {
	switch status {
	case lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED:
		return entity.OrderStatusUnspecified
	case lomspb.OrderStatus_ORDER_STATUS_NEW:
		return entity.OrderStatusNew
	case lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT:
		return entity.OrderStatusAwaitingPayment
	case lomspb.OrderStatus_ORDER_STATUS_FAILED:
		return entity.OrderStatusFailed
	case lomspb.OrderStatus_ORDER_STATUS_PAID:
		return entity.OrderStatusPaid
	case lomspb.OrderStatus_ORDER_STATUS_CANCELLED:
		return entity.OrderStatusCancelled
	default:
		return entity.OrderStatusUnspecified
	}
}

func FromOrderStatus(orderStatus entity.OrderStatus) lomspb.OrderStatus {
	switch orderStatus {
	case entity.OrderStatusUnspecified:
		return lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	case entity.OrderStatusNew:
		return lomspb.OrderStatus_ORDER_STATUS_NEW
	case entity.OrderStatusAwaitingPayment:
		return lomspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case entity.OrderStatusFailed:
		return lomspb.OrderStatus_ORDER_STATUS_FAILED
	case entity.OrderStatusPaid:
		return lomspb.OrderStatus_ORDER_STATUS_PAID
	case entity.OrderStatusCancelled:
		return lomspb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return lomspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func ToItem(item *lomspb.Item) entity.OrderItem {
	return entity.OrderItem{SKU: item.GetSku(), Count: item.GetCount()}
}
func FromItem(item entity.OrderItem) *lomspb.Item {
	return &lomspb.Item{Sku: item.SKU, Count: item.Count}
}

func ToItems(items []*lomspb.Item) []entity.OrderItem {
	result := make([]entity.OrderItem, len(items))
	for i, item := range items {
		result[i] = ToItem(item)
	}
	return result
}

func FromItems(items []entity.OrderItem) []*lomspb.Item {
	result := make([]*lomspb.Item, len(items))
	for i, item := range items {
		result[i] = FromItem(item)
	}
	return result
}

func ToGRPCError(err error) error {
	switch {
	case errors.Is(err, entity.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, entity.ErrOrderNotFound):
		return status.Error(codes.NotFound, "order not found")
	case errors.Is(err, entity.ErrStockNotFound):
		return status.Error(codes.NotFound, "stock not found")
	case errors.Is(err, entity.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, "insufficient stock")
	case errors.Is(err, entity.ErrInvalidStatus):
		return status.Error(codes.FailedPrecondition, "invalid order status")
	case errors.Is(err, loms.ErrWrongOrder):
		return status.Error(codes.InvalidArgument, "invalid order")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
