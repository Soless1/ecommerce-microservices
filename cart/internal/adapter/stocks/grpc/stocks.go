package grpc

import (
	"context"

	adapterconverter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/converter"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

type stocksClient struct {
	client stockspb.StocksClient
}

func NewStocksClient(stocksClientPb stockspb.StocksClient) *stocksClient {
	return &stocksClient{stocksClientPb}
}

func (c *stocksClient) GetStock(ctx context.Context, SKU uint32) (uint64, error) {
	response, err := c.client.GetStock(ctx, &stockspb.GetStockRequest{Sku: SKU})
	if err != nil {
		return 0, adapterconverter.FromStocksError(err)
	}
	return response.Count, nil
}

func (c *stocksClient) SetStock(ctx context.Context, SKU uint32, count uint64) error {
	_, err := c.client.SetStock(ctx, &stockspb.SetStockRequest{Sku: SKU, Count: count})
	if err != nil {
		return adapterconverter.FromStocksError(err)
	}
	return nil
}
