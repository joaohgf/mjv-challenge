package mapper

import (
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
	"github.com/joaohgf/mjv-challenge/internal/inbound/http/dto"
)

func TestOrderToMapsCreateRequest(t *testing.T) {
	order := new(Order).To(&dto.OrderCreate{ProductName: "Caderno", Quantity: 2})

	if order.ProductName != "Caderno" || order.Quantity != 2 {
		t.Fatalf("expected mapped request, got %#v", order)
	}
}

func TestOrderFromMapsDomainOrder(t *testing.T) {
	updatedAt := time.Now().UTC()
	source := &domain.Order{ID: "order-1", ProductName: "Caderno", Status: enum.Processado, Quantity: 2, UpdatedAt: &updatedAt}

	response := new(Order).From(source)

	if response.ID != source.ID || response.Status != string(enum.Processado) || response.UpdatedAt != &updatedAt {
		t.Fatalf("expected mapped response, got %#v", response)
	}
}
