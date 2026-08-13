package mapper

import (
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/enum"
	"github.com/joaohgf/mjv-challenge/internal/inbound/consumer/dto"
)

func TestOrderToPreservesMessageIdentity(t *testing.T) {
	updatedAt := time.Now().UTC()
	source := &dto.Order{
		ID: "order-1", ProductName: "Caderno", Status: string(enum.Criado), Quantity: 2,
		CreatedAt: updatedAt, UpdatedAt: &updatedAt,
	}

	result := new(Order).To(source)

	if result.ID != source.ID || result.Status != enum.Criado || result.UpdatedAt == nil {
		t.Fatalf("expected full order identity, got %#v", result)
	}
}
