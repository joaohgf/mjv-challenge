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

func TestWithOperationTimeoutUsesConfiguredBudget(t *testing.T) {
	repository := NewRepository[timeoutModel](&config.Database{OperationTimeout: time.Second})
	context, cancel := repository.WithOperationTimeout(context.Background())
	defer cancel()

	deadline, ok := context.Deadline()
	if !ok || time.Until(deadline) > time.Second || time.Until(deadline) < 900*time.Millisecond {
		t.Fatalf("expected one-second deadline, got %v", deadline)
	}
}

func TestWithOperationTimeoutPreservesEarlierParentDeadline(t *testing.T) {
	repository := NewRepository[timeoutModel](&config.Database{OperationTimeout: time.Second})
	parent, parentCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer parentCancel()
	child, cancel := repository.WithOperationTimeout(parent)
	defer cancel()

	deadline, ok := child.Deadline()
	if !ok || time.Until(deadline) > 50*time.Millisecond {
		t.Fatalf("expected parent deadline to be preserved, got %v", deadline)
	}
}

func TestPingFailsBeforeRepositoryConnects(t *testing.T) {
	repository := NewRepository[timeoutModel](&config.Database{OperationTimeout: time.Second})

	if err := repository.Ping(context.Background()); err == nil {
		t.Fatal("expected disconnected repository error")
	}
}
