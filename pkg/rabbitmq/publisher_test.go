package adapter

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestPublishingCreatesPersistentJSONMessage(t *testing.T) {
	published := publishing([]byte(`{"id":"order-1"}`))

	if published.ContentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", published.ContentType)
	}
	if published.DeliveryMode != amqp091.Persistent {
		t.Fatalf("expected persistent delivery, got %d", published.DeliveryMode)
	}
	if string(published.Body) != `{"id":"order-1"}` {
		t.Fatalf("expected preserved body, got %q", published.Body)
	}
}
