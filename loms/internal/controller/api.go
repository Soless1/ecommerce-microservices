package controller

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
)

type API struct {
	StockServer   *stocks.StocksServer
	ProductServer *product.ProductServer
	LomsServer    *loms.LomsServer
}

func New(stockServer *stocks.StocksServer, productServer *product.ProductServer, lomsServer *loms.LomsServer) *API {
	return &API{StockServer: stockServer, ProductServer: productServer, LomsServer: lomsServer}
}
