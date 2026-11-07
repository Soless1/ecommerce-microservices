package loms

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	lomsService interface {
		CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error)
		GetOrder(ctx context.Context, id int64) (entity.Order, error)
		PayOrder(ctx context.Context, id int64) error
		CancelOrder(ctx context.Context, id int64) error
	}
)

type LomsServer struct {
	lomsService lomsService
}

func NewLomsServer(lomsService lomsService) *LomsServer {
	return &LomsServer{
		lomsService: lomsService,
	}
}

//type LomsServer interface {
//	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
//	GetOrder(context.Context, *GetOrderRequest) (*GetOrderResponse, error)
//	PayOrder(context.Context, *PayOrderRequest) (*emptypb.Empty, error)
//	CancelOrder(context.Context, *CancelOrderRequest) (*emptypb.Empty, error)
//}

func (l *LomsServer) CreateOrder(ctx context.Context, req *lomspb.CreateOrderRequest) (*lomspb.CreateOrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	items := converter.ToItems(req.Items)
	orderID, err := l.lomsService.CreateOrder(ctx, req.UserId, items)
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &lomspb.CreateOrderResponse{OrderId: orderID}, nil
}

func (l *LomsServer) GetOrder(ctx context.Context, req *lomspb.GetOrderRequest) (*lomspb.GetOrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	order, err := l.lomsService.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &lomspb.GetOrderResponse{
		Status:    converter.FromOrderStatus(order.Status),
		UserId:    order.UserID,
		Items:     converter.FromItems(order.Items),
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}, nil
}

func (l *LomsServer) PayOrder(ctx context.Context, req *lomspb.PayOrderRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	err := l.lomsService.PayOrder(ctx, req.OrderId)
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
func (l *LomsServer) CancelOrder(ctx context.Context, req *lomspb.CancelOrderRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	err := l.lomsService.CancelOrder(ctx, req.OrderId)
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
