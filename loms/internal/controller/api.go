package controller

import (
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
)

type API struct {
	StockServer   *stocks.Server
	ProductServer *product.Server
	LomsServer    *loms.Server
}

func New(stockServer *stocks.Server, productServer *product.Server, lomsServer *loms.Server) *API {
	return &API{StockServer: stockServer, ProductServer: productServer, LomsServer: lomsServer}
}
