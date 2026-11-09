package converter

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	notificaitonspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

func FromOrderStatus(orderStatus entity.OrderStatus) notificaitonspb.OrderStatus {
	switch orderStatus {
	case entity.OrderStatusUnspecified:
		return notificationspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	case entity.OrderStatusNew:
		return notificationspb.OrderStatus_ORDER_STATUS_NEW
	case entity.OrderStatusAwaitingPayment:
		return notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT
	case entity.OrderStatusFailed:
		return notificationspb.OrderStatus_ORDER_STATUS_FAILED
	case entity.OrderStatusPaid:
		return notificationspb.OrderStatus_ORDER_STATUS_PAID
	case entity.OrderStatusCancelled:
		return notificationspb.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return notificationspb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}
