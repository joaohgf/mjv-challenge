package mapper

import (
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
)

func TestOrderMapperRoundTripsDomainData(t *testing.T) {
	updatedAt := time.Now().UTC()
	source := &domain.Order{
		ID: "order-1", ProductName: "Caderno", Status: enum.Processado, Quantity: 2,
		CreatedAt: updatedAt, UpdatedAt: &updatedAt,
	}
	mapper := new(OrderMapper)

	result := mapper.From(mapper.To(source))

	if result.ID != source.ID || result.Status != source.Status || result.UpdatedAt != source.UpdatedAt {
		t.Fatalf("expected full domain round trip, got %#v", result)
	}
}
