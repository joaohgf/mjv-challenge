package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/config"
)

type timeoutModel struct{}

func (timeoutModel) GetID() string      { return "" }
func (timeoutModel) GetIDField() string { return "id" }

func TestWithSaveTimeoutUsesConfiguredBudget(t *testing.T) {
	repository := NewRepository[timeoutModel](&config.Database{SaveTimeout: time.Second})
	context, cancel := repository.withSaveTimeout(context.Background())
	defer cancel()

	deadline, ok := context.Deadline()
	if !ok || time.Until(deadline) > time.Second || time.Until(deadline) < 900*time.Millisecond {
		t.Fatalf("expected one-second deadline, got %v", deadline)
	}
}
