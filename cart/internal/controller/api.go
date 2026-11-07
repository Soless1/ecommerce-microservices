package controller

import "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"

type API struct {
	Cart cart.CartServer
}

func New(cart cart.CartServer) *API {
	return &API{Cart: cart}
}
