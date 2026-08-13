package mapper

import (
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
)

func TestOrderToBuildsOrderMessage(t *testing.T) {
	createdAt := time.Now().UTC()
	source := &domain.Order{
		ID: "order-1", ProductName: "Caderno", Status: enum.Criado, Quantity: 2, CreatedAt: createdAt,
	}

	message := new(Order).To(source)

	if message.ID != source.ID || message.Type != "order" || message.Payload.ID != source.ID {
		t.Fatalf("expected order message, got %#v", message)
	}
	if message.Payload.Status != string(enum.Criado) || !message.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected mapped message metadata, got %#v", message)
	}
}
