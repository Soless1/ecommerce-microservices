package port

import (
	"errors"

	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
)

var ErrInvalidOrder = errors.New("invalid order")

type OrderItem struct {
	SKU   uint32
	Count uint32
}

func ToOrderItem(item *lomspb.Item) OrderItem {
	return OrderItem{SKU: item.GetSku(), Count: item.GetCount()}
}
func FromOrderItem(item OrderItem) *lomspb.Item {
	return &lomspb.Item{Sku: item.SKU, Count: item.Count}
}

func ToOrderItems(items []*lomspb.Item) []OrderItem {
	result := make([]OrderItem, len(items))
	for i, item := range items {
		result[i] = ToOrderItem(item)
	}
	return result
}

func FromOrderItems(items []OrderItem) []*lomspb.Item {
	result := make([]*lomspb.Item, len(items))
	for i, item := range items {
		result[i] = FromOrderItem(item)
	}
	return result
}
