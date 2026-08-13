package adapter

import (
	"errors"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type decodedMessage struct {
	ID string `json:"id"`
}

func TestConsumerDecodeReadsJSON(t *testing.T) {
	message, err := new(Consumer[decodedMessage]).decode([]byte(`{"id":"message-1"}`))

	if err != nil || message.ID != "message-1" {
		t.Fatalf("expected decoded message, got err=%v message=%#v", err, message)
	}
}

func TestConsumerDecodeReturnsInvalidJSONError(t *testing.T) {
	_, err := new(Consumer[decodedMessage]).decode([]byte(`invalid`))

	if err == nil {
		t.Fatal("expected JSON decoding error")
	}
}

func TestToPublishingPreservesMessageAndAddsFailureReason(t *testing.T) {
	createdAt := time.Now().UTC()
	delivery := amqp091.Delivery{
		Body: []byte("payload"), Headers: amqp091.Table{"trace-id": "trace-1"},
		ContentType: "application/json", DeliveryMode: amqp091.Persistent,
		CorrelationId: "correlation-1", Timestamp: createdAt,
	}

	published := toPublishing(delivery, errors.New("processing failed"))

	if string(published.Body) != "payload" || published.Headers["trace-id"] != "trace-1" {
		t.Fatalf("expected preserved delivery, got %#v", published)
	}
	if published.Headers["x-failure-reason"] != "processing failed" || published.CorrelationId != "correlation-1" {
		t.Fatalf("expected DLQ metadata, got %#v", published)
	}
	if !published.Timestamp.Equal(createdAt) || published.DeliveryMode != amqp091.Persistent {
		t.Fatalf("expected preserved AMQP properties, got %#v", published)
	}
}
