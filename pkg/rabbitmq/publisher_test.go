package adapter

import (
	"context"
	"testing"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

func TestPublishingCreatesPersistentJSONMessage(t *testing.T) {
	published := publishing(context.Background(), []byte(`{"id":"order-1"}`), nil)

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

func TestNewDeadLetterPublisherUsesParkingQueue(t *testing.T) {
	publisher := NewDeadLetterPublisher[string](&config.Queue{DeadLetterName: "orders.dlq"})

	if publisher.queue != "orders.dlq" {
		t.Fatalf("expected dead-letter queue, got %q", publisher.queue)
	}
}

func TestPublishingPreservesFailureReason(t *testing.T) {
	published := publishing(context.Background(), []byte("payload"), amqp091.Table{"x-failure-reason": "attempt limit"})

	if published.Headers["x-failure-reason"] != "attempt limit" {
		t.Fatalf("expected failure reason, got %#v", published.Headers)
	}
}
