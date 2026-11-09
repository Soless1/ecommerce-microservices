package entity

type OrderStatus string

const (
	OrderStatusUnspecified     OrderStatus = "unspecified"
	OrderStatusNew             OrderStatus = "new"
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatusFailed          OrderStatus = "failed"
	OrderStatusPaid            OrderStatus = "paid"
	OrderStatusCancelled       OrderStatus = "cancelled"
)
