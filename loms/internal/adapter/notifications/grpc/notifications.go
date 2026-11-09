package grpc

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/adapter/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

type NotificationsClient struct {
	notificationsClient notificationspb.NotificationsClient
}

func NewNotificationsClient(notificationsClient notificationspb.NotificationsClient) *NotificationsClient {
	return &NotificationsClient{notificationsClient}
}

func (c *NotificationsClient) SendOrderStatusChangedNotification(ctx context.Context, userId int64, orderId int64, status entity.OrderStatus) error {
	_, err := c.notificationsClient.SendOrderStatusChangedNotification(ctx, &notificationspb.OrderStatusChangedNotificationRequest{UserId: userId, OrderId: orderId, Status: converter.FromOrderStatus(status)})
	return err
}
