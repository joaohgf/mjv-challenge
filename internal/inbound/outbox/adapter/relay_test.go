package adapter

import (
	"context"
	"errors"
	"testing"
	"time"
)

type dispatchStub struct {
	dispatched bool
	err        error
	calls      int
}

func (stub *dispatchStub) Dispatch(context.Context) (bool, error) {
	stub.calls++
	return stub.dispatched, stub.err
}

func TestRelayStopsWhenContextIsCancelled(t *testing.T) {
	context, cancel := context.WithCancel(context.Background())
	stub := new(dispatchStub)
	cancel()

	err := NewRelay(stub, time.Millisecond).Run(context)

	if !errors.Is(err, context.Err()) || stub.calls != 1 {
		t.Fatalf("expected one cancelled dispatch, got err=%v calls=%d", err, stub.calls)
	}
}
