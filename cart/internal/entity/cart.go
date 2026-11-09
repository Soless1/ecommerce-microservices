package entity

import "errors"

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrCartItemNotFound  = errors.New("cart item not found")
)

type CartItem struct {
	SKU   uint32
	Count uint32
}
