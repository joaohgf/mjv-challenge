package adapter

import (
	"context"
	"testing"

	"github.com/joaohgf/mjv-challenge/config"
)

func TestCloseWithoutClientDoesNothing(t *testing.T) {
	repository := NewRepository[timeoutModel](&config.Database{})

	if err := repository.Close(context.Background()); err != nil {
		t.Fatalf("expected no close error, got %v", err)
	}
}
