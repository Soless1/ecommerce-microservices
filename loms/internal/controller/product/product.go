package product

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	productService interface {
		CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error)
		GetProduct(ctx context.Context, sku uint32) (entity.Product, error)
	}
)

type Server struct {
	productService productService
}

func NewProductServer(productService productService) *Server {
	return &Server{
		productService: productService,
	}
}

// type ProductServiceServer interface {
//	CreateProduct(context.Context, *CreateProductRequest) (*CreateProductResponse, error)
//	GetProduct(context.Context, *GetProductRequest) (*GetProductResponse, error)
// }

func (p *Server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	product, err := p.productService.CreateProduct(ctx, req.GetName(), req.GetPrice())
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &productpb.CreateProductResponse{Sku: product.SKU}, nil
}
func (p *Server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	product, err := p.productService.GetProduct(ctx, req.GetSku())
	if err != nil {
		return nil, converter.ToGRPCError(err)
	}
	return &productpb.GetProductResponse{Price: product.Price, Name: product.Name}, nil
}
