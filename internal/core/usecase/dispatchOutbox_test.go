package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
)

type (
	outboxStub struct {
		event     *domain.OutboxEvent[string]
		claimErr  error
		released  string
		published string
	}
	publisherStub struct {
		message string
		err     error
	}
)

func (stub *outboxStub) Enqueue(context.Context, string) error { return nil }

func (stub *outboxStub) Claim(context.Context) (*domain.OutboxEvent[string], error) {
	return stub.event, stub.claimErr
}

func (stub *outboxStub) MarkPublished(_ context.Context, id string) error {
	stub.published = id
	return nil
}

func (stub *outboxStub) Release(_ context.Context, id string) error {
	stub.released = id
	return nil
}

func (stub *publisherStub) Publish(_ context.Context, message string) error {
	stub.message = message
	return stub.err
}

func TestDispatchOutboxPublishesAndMarksClaimedEvent(t *testing.T) {
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", Payload: "order-1"}}
	publisher := new(publisherStub)

	dispatched, err := NewDispatchOutbox(outbox, publisher).Dispatch(context.Background())

	if err != nil || !dispatched || publisher.message != "order-1" || outbox.published != "event-1" {
		t.Fatalf("expected confirmed event dispatch, got dispatched=%t err=%v", dispatched, err)
	}
}

func TestDispatchOutboxReleasesEventAfterPublicationFailure(t *testing.T) {
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", Payload: "order-1"}}
	publisher := &publisherStub{err: errors.New("broker unavailable")}

	dispatched, err := NewDispatchOutbox(outbox, publisher).Dispatch(context.Background())

	if err == nil || !dispatched || outbox.released != "event-1" || outbox.published != "" {
		t.Fatalf("expected released event, got dispatched=%t err=%v outbox=%#v", dispatched, err, outbox)
	}
}

func TestDispatchOutboxDoesNothingWhenNoEventIsAvailable(t *testing.T) {
	dispatched, err := NewDispatchOutbox(&outboxStub{claimErr: errs.ErrNotFound}, new(publisherStub)).Dispatch(context.Background())

	if err != nil || dispatched {
		t.Fatalf("expected idle dispatch, got dispatched=%t err=%v", dispatched, err)
	}
}
