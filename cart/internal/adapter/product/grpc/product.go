package grpc

import (
	"context"

	adapterconverter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/converter"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
)

type productClient struct {
	client productpb.ProductServiceClient
}

func NewProductClient(productClientPb productpb.ProductServiceClient) *productClient {
	return &productClient{productClientPb}
}

func (c *productClient) GetProduct(ctx context.Context, sku uint32) (port.ProductInfo, error) {
	response, err := c.client.GetProduct(ctx, &productpb.GetProductRequest{Sku: sku})
	if err != nil {
		return port.ProductInfo{}, adapterconverter.FromProductServiceError(err)
	}
	return port.ProductInfo{Price: response.GetPrice(), Name: response.GetName()}, nil
}
