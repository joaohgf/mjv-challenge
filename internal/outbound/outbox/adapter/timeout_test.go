package adapter

import (
	"context"
	"testing"
	"time"
)

func TestStoreUsesInjectedOperationTimeout(t *testing.T) {
	called := false
	store := NewStore[string, string](nil, stringMapper{}, time.Second, func(ctx context.Context) (context.Context, context.CancelFunc) {
		called = true
		return context.WithTimeout(ctx, time.Second)
	})

	context, cancel := store.withTimeout(context.Background())
	defer cancel()
	if !called {
		t.Fatal("expected injected timeout to be used")
	}
	if _, ok := context.Deadline(); !ok {
		t.Fatal("expected timeout context deadline")
	}
}

type stringMapper struct{}

func (stringMapper) To(value string) string   { return value }
func (stringMapper) From(value string) string { return value }
