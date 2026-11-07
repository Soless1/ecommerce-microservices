package entity

import "errors"

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrStockNotFound     = errors.New("stock not found")
	ErrInvalidStatus     = errors.New("invalid order status for operation")
)
