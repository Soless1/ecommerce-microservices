package grpc

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	notificaitonspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

func FromOrderStatus(orderStatus port.OrderStatus) notificaitonspb.OrderStatus {
	switch orderStatus {
	case port.OrderStatusNew:
		return notificaitonspb.OrderStatus_ORDER_STATUS_NEW
	case port.OrderStatusAwaitingPayment:
		return notificaitonspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case port.OrderStatusFailed:
		return notificaitonspb.OrderStatus_ORDER_STATUS_FAILED
	case port.OrderStatusPaid:
		return notificaitonspb.OrderStatus_ORDER_STATUS_PAID
	case port.OrderStatusCancelled:
		return notificaitonspb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return notificaitonspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
