package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
)

type (
	transactionStub struct{ err error }
	orderStoreStub  struct {
		saved *domain.Order
		err   error
	}
	outboxStoreStub struct {
		enqueued *domain.Order
		err      error
	}
)

func (stub transactionStub) Transaction(ctx context.Context, operation func(context.Context) error) error {
	if stub.err != nil {
		return stub.err
	}
	return operation(ctx)
}

func (stub *orderStoreStub) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	stub.saved = order
	return order, stub.err
}

func (stub *outboxStoreStub) Enqueue(_ context.Context, order *domain.Order) error {
	stub.enqueued = order
	return stub.err
}

func (stub *outboxStoreStub) Claim(context.Context) (*domain.OutboxEvent[*domain.Order], error) {
	return nil, nil
}
func (stub *outboxStoreStub) MarkPublished(context.Context, string, string) error    { return nil }
func (stub *outboxStoreStub) MarkDeadLettered(context.Context, string, string) error { return nil }
func (stub *outboxStoreStub) Release(context.Context, string, string) error          { return nil }

func TestCreatorCreatesAndEnqueuesWithinTransaction(t *testing.T) {
	orders := new(orderStoreStub)
	outbox := new(outboxStoreStub)
	order := &domain.Order{ID: "order-1"}

	created, err := NewCreator(transactionStub{}, orders, outbox).Create(context.Background(), order)

	if err != nil || created != order || orders.saved != order || outbox.enqueued != order {
		t.Fatalf("expected transactional creation and enqueue, got err=%v", err)
	}
}

func TestCreatorDoesNotEnqueueWhenCreationFails(t *testing.T) {
	orders := &orderStoreStub{err: errors.New("database unavailable")}
	outbox := new(outboxStoreStub)

	_, err := NewCreator(transactionStub{}, orders, outbox).Create(context.Background(), &domain.Order{})

	if err == nil || outbox.enqueued != nil {
		t.Fatalf("expected creation error without enqueue, got err=%v outbox=%#v", err, outbox)
	}
}
