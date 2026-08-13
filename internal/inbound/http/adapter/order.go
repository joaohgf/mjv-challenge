package adapter

import (
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
	"github.com/joaohgf/mjv-challenge/internal/inbound/http/dto"
)

// OrderHandler translates HTTP requests and responses at the API boundary.
type OrderHandler struct {
	toMapper   port.To[*dto.OrderCreate, *domain.Order]
	fromMapper port.From[*domain.Order, *dto.OrderResponse]
	creator    port.Creator[*domain.Order]
	getter     port.Getter[*domain.Order]
}

// NewOrderHandler receives all ports required by the HTTP order routes.
func NewOrderHandler(
	toMapper port.To[*dto.OrderCreate, *domain.Order],
	fromMapper port.From[*domain.Order, *dto.OrderResponse],
	creator port.Creator[*domain.Order],
	getter port.Getter[*domain.Order],
) *OrderHandler {
	return &OrderHandler{
		toMapper: toMapper, fromMapper: fromMapper, creator: creator, getter: getter,
	}
}
