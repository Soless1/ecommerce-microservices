package converter

import (
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FromProductServiceError(err error) error {
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.NotFound {
		return port.ErrProductNotFound
	}
	return err
}

func FromStocksError(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.NotFound, codes.FailedPrecondition:
		return entity.ErrInsufficientStock
	default:
		return err
	}
}

func FromLOMSError(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.NotFound:
		return port.ErrProductNotFound
	case codes.FailedPrecondition:
		return entity.ErrInsufficientStock
	case codes.InvalidArgument:
		return port.ErrInvalidOrder
	default:
		return err
	}
}
