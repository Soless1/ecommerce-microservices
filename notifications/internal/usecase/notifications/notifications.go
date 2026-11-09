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
func (n *NotificationsService) SendOrderStatusChangedNotification(ctx context.Context, userID int64, orderID int64, status entity.OrderStatus) error {
	time.Sleep(time.Millisecond * 400)

	n.logger.Info("Order status changed",
		zap.String("user_id", strconv.FormatInt(userID, 10)),
		zap.String("order_id", strconv.FormatInt(orderID, 10)),
		zap.String("status", string(status)))
	return nil
}
