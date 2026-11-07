package entity

import "time"

type OrderStatus string

const (
	OrderStatusUnspecified     OrderStatus = "unspecified"
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

type Order struct {
	ID        int64
	UserID    int64
	Status    OrderStatus
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	SKU   uint32
	Count uint32
}
