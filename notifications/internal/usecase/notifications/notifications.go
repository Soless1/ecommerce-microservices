package notifications

import (
	"context"
	"strconv"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/entity"
	"go.uber.org/zap"
)

type NotificationsService struct {
	logger *zap.Logger
}

func NewNotificationsServer(logger *zap.Logger) *NotificationsService {
	return &NotificationsService{
		logger: logger}
}
func (n *NotificationsService) SendOrderStatusChangedNotification(ctx context.Context, body entity.OrderStatusChangedNotification) error {
	time.Sleep(time.Second * 30)

	n.logger.Info("Order status changed",
		zap.String("user_id", strconv.FormatInt(body.UserID, 10)),
		zap.String("order_id", strconv.FormatInt(body.OrderID, 10)),
		zap.String("status", string(body.Status)))
	return nil
}
