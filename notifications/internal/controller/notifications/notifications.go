package notifications

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/controller/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
)

type NotificationService interface {
	SendOrderStatusChangedNotification(ctx context.Context, userId int64, orderId int64, orderStatus entity.OrderStatus) error
}
type NotificationsServer struct {
	notificationsService NotificationService
}

func NewNotificationsServer(notificationsService NotificationService) *NotificationsServer {
	return &NotificationsServer{
		notificationsService: notificationsService,
	}
}

func (s *NotificationsServer) SendOrderStatusChangedNotification(ctx context.Context, req *notificationspb.OrderStatusChangedNotificationRequest) (*emptypb.Empty, error) {
	if req.Validate() != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	err := s.notificationsService.SendOrderStatusChangedNotification(ctx, req.UserId, req.OrderId, converter.ToOrderStatus(req.Status))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
