package telemetry

import (
	"context"
	"testing"

	"github.com/joaohgf/mjv-challenge/config"
)

func TestStartDisabledReturnsNoopShutdown(t *testing.T) {
	shutdown, err := Start(context.Background(), &config.Telemetry{Enabled: false})
	if err != nil {
		t.Fatalf("expected disabled telemetry to start, got %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("expected noop shutdown, got %v", err)
	}
}
