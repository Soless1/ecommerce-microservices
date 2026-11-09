package converter

import (
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/entity"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

func ToOrderStatus(status notificationspb.OrderStatus) entity.OrderStatus {
	switch status {
	case notificationspb.OrderStatus_ORDER_STATUS_NEW:
		return entity.OrderStatusNew
	case notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT:
		return entity.OrderStatusAwaitingPayment
	case notificationspb.OrderStatus_ORDER_STATUS_PAID:
		return entity.OrderStatusPaid
	case notificationspb.OrderStatus_ORDER_STATUS_CANCELLED:
		return entity.OrderStatusCancelled
	case notificationspb.OrderStatus_ORDER_STATUS_FAILED:
		return entity.OrderStatusFailed
	default:
		return entity.OrderStatusUnspecified
	}
}
