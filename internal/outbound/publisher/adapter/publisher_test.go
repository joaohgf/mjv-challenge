package adapter

import (
	"context"
	"errors"
	"testing"
)

type (
	messageMapper struct{}
	messageSender struct {
		message string
		cause   error
		err     error
	}
)

func (messageMapper) To(value int) string { return "message" }

func (sender *messageSender) Publish(_ context.Context, message string) error {
	sender.message = message
	return sender.err
}

func (sender *messageSender) DeadLetter(_ context.Context, message string, cause error) error {
	sender.message, sender.cause = message, cause
	return sender.err
}

func TestPublisherMapsAndSendsMessage(t *testing.T) {
	sender := new(messageSender)
	err := NewPublisher[int, string](sender, messageMapper{}).Publish(context.Background(), 1)

	if err != nil || sender.message != "message" {
		t.Fatalf("expected mapped message to be sent, got err=%v sender=%#v", err, sender)
	}
}

func TestDeadLetterPublisherMapsAndParksMessage(t *testing.T) {
	sender := &messageSender{err: errors.New("queue unavailable")}
	cause := errors.New("publish failed")
	err := NewDeadLetterPublisher[int, string](sender, messageMapper{}).DeadLetter(context.Background(), 1, cause)

	if !errors.Is(err, sender.err) || sender.message != "message" || !errors.Is(sender.cause, cause) {
		t.Fatalf("expected mapped message and original cause, got err=%v sender=%#v", err, sender)
	}
}
