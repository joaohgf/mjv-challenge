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
		event           *domain.OutboxEvent[string]
		claimErr        error
		released        string
		releaseToken    string
		published       string
		publishToken    string
		publishErr      error
		deadLettered    string
		deadLetterToken string
	}
	publisherStub struct {
		message string
		err     error
		context context.Context
		cause   error
	}
)

func (stub *outboxStub) Enqueue(context.Context, string) error { return nil }

func (stub *outboxStub) Claim(context.Context) (*domain.OutboxEvent[string], error) {
	return stub.event, stub.claimErr
}

func (stub *outboxStub) MarkPublished(_ context.Context, id, leaseToken string) error {
	stub.published = id
	stub.publishToken = leaseToken
	return stub.publishErr
}

func (stub *outboxStub) MarkDeadLettered(_ context.Context, id, leaseToken string) error {
	stub.deadLettered = id
	stub.deadLetterToken = leaseToken
	return nil
}

func (stub *outboxStub) Release(_ context.Context, id, leaseToken string) error {
	stub.released = id
	stub.releaseToken = leaseToken
	return nil
}

func (stub *publisherStub) Publish(ctx context.Context, message string) error {
	stub.context = ctx
	stub.message = message
	return stub.err
}

func (stub *publisherStub) DeadLetter(ctx context.Context, message string, cause error) error {
	stub.context = ctx
	stub.message = message
	stub.cause = cause
	return stub.err
}

func TestDispatchOutboxUsesRestoredEventContext(t *testing.T) {
	restored := context.WithValue(context.Background(), "trace", "request-1")
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", LeaseToken: "lease-1", Payload: "order-1", Context: restored}}
	publisher := new(publisherStub)

	_, err := NewDispatchOutbox(outbox, publisher, publisher, 5).Dispatch(context.Background())

	if err != nil || publisher.context.Value("trace") != "request-1" {
		t.Fatalf("expected restored event context, got err=%v context=%v", err, publisher.context)
	}
}

func TestDispatchOutboxPublishesAndMarksClaimedEvent(t *testing.T) {
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", LeaseToken: "lease-1", Payload: "order-1"}}
	publisher := new(publisherStub)

	dispatched, err := NewDispatchOutbox(outbox, publisher, publisher, 5).Dispatch(context.Background())

	if err != nil || !dispatched || publisher.message != "order-1" || outbox.published != "event-1" || outbox.publishToken != "lease-1" {
		t.Fatalf("expected confirmed event dispatch, got dispatched=%t err=%v", dispatched, err)
	}
}

func TestDispatchOutboxReportsLeaseLossAfterPublication(t *testing.T) {
	outbox := &outboxStub{
		event:      &domain.OutboxEvent[string]{ID: "event-1", LeaseToken: "expired", Payload: "order-1"},
		publishErr: errs.ErrLeaseLost,
	}
	dispatched, err := NewDispatchOutbox(outbox, new(publisherStub), new(publisherStub), 5).Dispatch(context.Background())

	if !dispatched || !errors.Is(err, errs.ErrLeaseLost) {
		t.Fatalf("expected reported lease loss, got dispatched=%t err=%v", dispatched, err)
	}
}

func TestDispatchOutboxReleasesEventAfterPublicationFailure(t *testing.T) {
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", LeaseToken: "lease-1", Payload: "order-1"}}
	publisher := &publisherStub{err: errors.New("broker unavailable")}

	dispatched, err := NewDispatchOutbox(outbox, publisher, publisher, 5).Dispatch(context.Background())

	if err == nil || !dispatched || outbox.released != "event-1" || outbox.releaseToken != "lease-1" || outbox.published != "" {
		t.Fatalf("expected released event, got dispatched=%t err=%v outbox=%#v", dispatched, err, outbox)
	}
}

func TestDispatchOutboxDoesNothingWhenNoEventIsAvailable(t *testing.T) {
	publisher := new(publisherStub)
	dispatched, err := NewDispatchOutbox(&outboxStub{claimErr: errs.ErrNotFound}, publisher, publisher, 5).Dispatch(context.Background())

	if err != nil || dispatched {
		t.Fatalf("expected idle dispatch, got dispatched=%t err=%v", dispatched, err)
	}
}

func TestDispatchOutboxParksEventAfterFiveFailedAttempts(t *testing.T) {
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", LeaseToken: "lease-1", Payload: "order-1", Attempts: 5}}
	publisher := &publisherStub{err: errors.New("broker unavailable")}
	deadLetter := new(publisherStub)

	dispatched, err := NewDispatchOutbox(outbox, publisher, deadLetter, 5).Dispatch(context.Background())

	if err != nil || !dispatched || deadLetter.message != "order-1" || deadLetter.cause == nil || outbox.deadLettered != "event-1" || outbox.deadLetterToken != "lease-1" || outbox.released != "" {
		t.Fatalf("expected parked event, got dispatched=%t err=%v outbox=%#v", dispatched, err, outbox)
	}
}

func TestDispatchOutboxParksEventPastAttemptLimit(t *testing.T) {
	outbox := &outboxStub{event: &domain.OutboxEvent[string]{ID: "event-1", LeaseToken: "lease-1", Payload: "order-1", Attempts: 6}}
	publisher, deadLetter := new(publisherStub), new(publisherStub)

	dispatched, err := NewDispatchOutbox(outbox, publisher, deadLetter, 5).Dispatch(context.Background())

	if err != nil || !dispatched || publisher.message != "" || deadLetter.cause == nil || outbox.deadLettered != "event-1" || outbox.deadLetterToken != "lease-1" {
		t.Fatalf("expected event parked without publishing, got dispatched=%t err=%v outbox=%#v", dispatched, err, outbox)
	}
}
