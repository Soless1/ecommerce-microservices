package product

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
)

var ErrProductNotFound = entity.ErrProductNotFound

type (
	productRepository interface {
		CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error)
		GetProduct(ctx context.Context, sku uint32) (entity.Product, error)
	}
)

//go:generate mockgen -source=product.go -destination=mocks/product_mock.go -package=mocks

type productService struct {
	productRepository productRepository
}

func NewProductService(productRepository productRepository) *productService {
	return &productService{
		productRepository: productRepository,
	}
}

func (s *productService) CreateProduct(ctx context.Context, name string, price uint32) (entity.Product, error) {
	return s.productRepository.CreateProduct(ctx, name, price)
}

func (s *productService) GetProduct(ctx context.Context, sku uint32) (entity.Product, error) {
	return s.productRepository.GetProduct(ctx, sku)
}
