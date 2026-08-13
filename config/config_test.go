package config

import (
	"testing"
	"time"
)

func TestLoadUsesSwaggerHostEnvironment(t *testing.T) {
	t.Setenv("SWAGGER_HOST", "api.example.com")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "7s")

	settings := Load()
	if settings.SwaggerHost != "api.example.com" || settings.HTTPShutdownTimeout != 7*time.Second {
		t.Fatalf("expected HTTP configuration from environment, got %#v", settings)
	}
}

func TestLoadQueueUsesEnvironment(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "amqp://queue.example")
	t.Setenv("RABBITMQ_QUEUE_NAME", "work")
	t.Setenv("RABBITMQ_DEAD_LETTER_QUEUE_NAME", "work.dlq")
	t.Setenv("RABBITMQ_PUBLISH_TIMEOUT", "3s")
	t.Setenv("OUTBOX_LEASE_DURATION", "10s")
	t.Setenv("OUTBOX_RETRY_INTERVAL", "2s")
	t.Setenv("OUTBOX_MAX_ATTEMPTS", "7")

	queue := LoadQueue()

	if queue.RabbitMQURL != "amqp://queue.example" || queue.Name != "work" || queue.DeadLetterName != "work.dlq" || queue.PublishTimeout != 3*time.Second || queue.OutboxLease != 10*time.Second || queue.OutboxRetryInterval != 2*time.Second || queue.OutboxMaxAttempts != 7 {
		t.Fatalf("expected queue configuration from environment, got %#v", queue)
	}
}

func TestPositiveIntUsesFallbackForInvalidValue(t *testing.T) {
	t.Setenv("INVALID_NUMBER", "zero")

	if got := positiveInt("INVALID_NUMBER", 5); got != 5 {
		t.Fatalf("expected fallback value, got %d", got)
	}
}

func TestLoadTelemetryUsesEnvironment(t *testing.T) {
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "orders-api")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "collector:4317")

	telemetry := LoadTelemetry()
	if !telemetry.Enabled || telemetry.ServiceName != "orders-api" || telemetry.Endpoint != "collector:4317" {
		t.Fatalf("expected telemetry configuration from environment, got %#v", telemetry)
	}
}

func TestValueUsesFallbackForEmptyEnvironment(t *testing.T) {
	t.Setenv("EMPTY_VALUE", "")

	if got := value("EMPTY_VALUE", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}

func TestLoadDatabaseUsesConfiguredOperationTimeout(t *testing.T) {
	t.Setenv("MONGODB_OPERATION_TIMEOUT", "3s")
	t.Setenv("MONGO_OUTBOX_COLLECTION_NAME", "events")

	database := LoadDatabase()
	if database.OperationTimeout != 3*time.Second || database.OutboxCollectionName != "events" {
		t.Fatalf("expected configured database settings, got %#v", database)
	}
}

func TestDurationUsesFallbackForInvalidValue(t *testing.T) {
	t.Setenv("INVALID_DURATION", "invalid")

	if got := duration("INVALID_DURATION", time.Second); got != time.Second {
		t.Fatalf("expected fallback duration, got %s", got)
	}
}
