package stocks

import (
	"context"
)

type (
	stocksRepository interface {
		SetStock(ctx context.Context, sku uint32, count uint64) error
		GetStock(ctx context.Context, sku uint32) (uint64, error)
	}
)

//go:generate mockgen -source=stocks.go -destination=mocks/stocks_mock.go -package=mocks

type stocksService struct {
	stocksRepository stocksRepository
}

func NewStocksService(stocksRepository stocksRepository) *stocksService {
	return &stocksService{
		stocksRepository: stocksRepository,
	}
}

func (s *stocksService) SetStock(ctx context.Context, sku uint32, count uint64) error {
	return s.stocksRepository.SetStock(ctx, sku, count)
}

func (s *stocksService) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	return s.stocksRepository.GetStock(ctx, sku)
}
