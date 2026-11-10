package grpc

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

type NotificationsClient struct {
	notificationsClient notificationspb.NotificationsClient
}

func NewNotificationsClient(notificationsClient notificationspb.NotificationsClient) *NotificationsClient {
	return &NotificationsClient{notificationsClient}
}

func (c *NotificationsClient) SendOrderStatusChangedNotification(ctx context.Context, order port.OrderStatusChangedNotification) error {
	_, err := c.notificationsClient.SendOrderStatusChangedNotification(ctx, &notificationspb.OrderStatusChangedNotificationRequest{UserId: order.UserID, OrderId: order.OrderID, Status: FromOrderStatus(order.Status)})
	return err
}
