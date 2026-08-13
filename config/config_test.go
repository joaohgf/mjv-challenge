package config

import (
	"testing"
	"time"
)

func TestLoadUsesSwaggerHostEnvironment(t *testing.T) {
	t.Setenv("SWAGGER_HOST", "api.example.com")

	if got := Load().SwaggerHost; got != "api.example.com" {
		t.Fatalf("expected Swagger host from environment, got %q", got)
	}
}

func TestLoadQueueUsesEnvironment(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "amqp://queue.example")
	t.Setenv("RABBITMQ_QUEUE_NAME", "work")
	t.Setenv("RABBITMQ_DEAD_LETTER_QUEUE_NAME", "work.dlq")
	t.Setenv("RABBITMQ_PUBLISH_TIMEOUT", "3s")
	t.Setenv("OUTBOX_LEASE_DURATION", "10s")
	t.Setenv("OUTBOX_RETRY_INTERVAL", "2s")

	queue := LoadQueue()

	if queue.RabbitMQURL != "amqp://queue.example" || queue.Name != "work" || queue.DeadLetterName != "work.dlq" || queue.PublishTimeout != 3*time.Second || queue.OutboxLease != 10*time.Second || queue.OutboxRetryInterval != 2*time.Second {
		t.Fatalf("expected queue configuration from environment, got %#v", queue)
	}
}

func TestValueUsesFallbackForEmptyEnvironment(t *testing.T) {
	t.Setenv("EMPTY_VALUE", "")

	if got := value("EMPTY_VALUE", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}

func TestLoadDatabaseUsesConfiguredSaveTimeout(t *testing.T) {
	t.Setenv("MONGODB_SAVE_TIMEOUT", "3s")
	t.Setenv("MONGO_OUTBOX_COLLECTION_NAME", "events")

	database := LoadDatabase()
	if database.SaveTimeout != 3*time.Second || database.OutboxCollectionName != "events" {
		t.Fatalf("expected configured database settings, got %#v", database)
	}
}

func TestDurationUsesFallbackForInvalidValue(t *testing.T) {
	t.Setenv("INVALID_DURATION", "invalid")

	if got := duration("INVALID_DURATION", time.Second); got != time.Second {
		t.Fatalf("expected fallback duration, got %s", got)
	}
}
