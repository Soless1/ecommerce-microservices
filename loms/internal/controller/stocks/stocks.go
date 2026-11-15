package stocks

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/converter"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type stocksService interface {
	SetStock(ctx context.Context, sku uint32, count uint64) error
	GetStock(ctx context.Context, sku uint32) (uint64, error)
}

type Server struct {
	stocksService stocksService
}

func NewStocksServer(stocksService stocksService) *Server {
	return &Server{stocksService: stocksService}
}

func (s *Server) SetStock(ctx context.Context, req *stockspb.SetStockRequest) (*emptypb.Empty, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	if err := s.stocksService.SetStock(ctx, req.GetSku(), req.GetCount()); err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetStock(ctx context.Context, req *stockspb.GetStockRequest) (*stockspb.GetStockResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	count, err := s.stocksService.GetStock(ctx, req.GetSku())
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &stockspb.GetStockResponse{Count: count}, nil
}
