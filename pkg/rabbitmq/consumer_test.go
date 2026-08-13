package adapter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

type (
	decodedMessage struct {
		ID string `json:"id"`
	}
	confirmationStub struct {
		acknowledged bool
		err          error
		waited       bool
	}
	confirmationEnablerStub struct {
		enabled bool
		err     error
	}
	acknowledgerStub struct {
		acknowledged bool
		nacked       bool
		rejected     bool
	}
)

func (stub *confirmationStub) WaitContext(context.Context) (bool, error) {
	stub.waited = true
	return stub.acknowledged, stub.err
}

func (stub *confirmationEnablerStub) Confirm(wait bool) error {
	stub.enabled = !wait
	return stub.err
}

func (stub *acknowledgerStub) Ack(uint64, bool) error {
	stub.acknowledged = true
	return nil
}

func (stub *acknowledgerStub) Nack(uint64, bool, bool) error {
	stub.nacked = true
	return nil
}

func (stub *acknowledgerStub) Reject(uint64, bool) error {
	stub.rejected = true
	return nil
}

func TestConsumerDecodeReadsJSON(t *testing.T) {
	message, err := new(Consumer[decodedMessage]).decode([]byte(`{"id":"message-1"}`))

	if err != nil || message.ID != "message-1" {
		t.Fatalf("expected decoded message, got err=%v message=%#v", err, message)
	}
}

func TestConsumerDecodeReturnsInvalidJSONError(t *testing.T) {
	_, err := new(Consumer[decodedMessage]).decode([]byte(`invalid`))

	if err == nil {
		t.Fatal("expected JSON decoding error")
	}
}

func TestEnableConfirmsUsesSynchronousMode(t *testing.T) {
	channel := new(confirmationEnablerStub)

	if err := enableConfirms(channel); err != nil || !channel.enabled {
		t.Fatalf("expected synchronous publisher confirms, got err=%v channel=%#v", err, channel)
	}
}

func TestWaitConfirmationRequiresBrokerAcknowledgement(t *testing.T) {
	confirmation := new(confirmationStub)

	if err := waitConfirmation(context.Background(), confirmation); err == nil || !confirmation.waited {
		t.Fatalf("expected rejected confirmation error, got err=%v confirmation=%#v", err, confirmation)
	}
}

func TestReturnedMessageReportsRoutingFailure(t *testing.T) {
	returns := make(chan amqp091.Return, 1)
	returns <- amqp091.Return{ReplyCode: 312, ReplyText: "NO_ROUTE"}
	consumer := &Consumer[string]{returns: returns}

	if err := consumer.returnedMessage(); err == nil || !strings.Contains(err.Error(), "NO_ROUTE") {
		t.Fatalf("expected routing error, got %v", err)
	}
}

func TestDeadLetterAcknowledgesOnlyAfterPublish(t *testing.T) {
	acknowledger := new(acknowledgerStub)
	confirmation := &confirmationStub{acknowledged: true}
	published := false
	consumer := &Consumer[string]{
		Client: NewClient(&config.Queue{DeadLetterName: "orders.dlq"}),
		deadLetterPublish: func(ctx context.Context, _ amqp091.Publishing) error {
			if acknowledger.acknowledged {
				t.Fatal("source message was acknowledged before the DLQ publication")
			}
			published = true
			return waitConfirmation(ctx, confirmation)
		},
	}
	delivery := amqp091.Delivery{Acknowledger: acknowledger, DeliveryTag: 1, Body: []byte("payload")}

	if err := consumer.deadLetter(context.Background(), delivery, errors.New("processing failed")); err != nil {
		t.Fatal(err)
	}
	if !published || !confirmation.waited || !acknowledger.acknowledged || acknowledger.nacked || acknowledger.rejected {
		t.Fatalf("expected confirmed publish then acknowledgement only, got %#v", acknowledger)
	}
}

func TestDeadLetterDoesNotAcknowledgeWhenPublishFails(t *testing.T) {
	acknowledger := new(acknowledgerStub)
	consumer := &Consumer[string]{
		Client: NewClient(&config.Queue{DeadLetterName: "orders.dlq"}),
		deadLetterPublish: func(context.Context, amqp091.Publishing) error {
			return errors.New("confirmation timeout")
		},
	}
	delivery := amqp091.Delivery{Acknowledger: acknowledger, DeliveryTag: 1, Body: []byte("payload")}

	if err := consumer.deadLetter(context.Background(), delivery, errors.New("processing failed")); err == nil {
		t.Fatal("expected dead-letter publication error")
	}
	if acknowledger.acknowledged || acknowledger.nacked || acknowledger.rejected {
		t.Fatalf("expected source message to remain unacknowledged, got %#v", acknowledger)
	}
}

func TestConsumerParksNilDecodedMessage(t *testing.T) {
	acknowledger := new(acknowledgerStub)
	published := false
	consumer := &Consumer[*decodedMessage]{
		Client: NewClient(&config.Queue{DeadLetterName: "orders.dlq"}),
		deadLetterPublish: func(context.Context, amqp091.Publishing) error {
			published = true
			return nil
		},
	}
	delivery := amqp091.Delivery{Acknowledger: acknowledger, DeliveryTag: 1, Body: []byte("null")}

	err := consumer.process(context.Background(), delivery, func(_ context.Context, message *decodedMessage) error {
		if message != nil {
			t.Fatalf("expected a nil decoded message, got %#v", message)
		}
		return errors.New("invalid message: envelope is required")
	})
	if err != nil || !published || !acknowledger.acknowledged {
		t.Fatalf("expected nil message parked in DLQ, got err=%v published=%t ack=%t", err, published, acknowledger.acknowledged)
	}
}

func TestConsumerParksHandlerPanic(t *testing.T) {
	acknowledger := new(acknowledgerStub)
	published := false
	consumer := &Consumer[decodedMessage]{
		Client: NewClient(&config.Queue{DeadLetterName: "orders.dlq"}),
		deadLetterPublish: func(context.Context, amqp091.Publishing) error {
			published = true
			return nil
		},
	}
	delivery := amqp091.Delivery{Acknowledger: acknowledger, DeliveryTag: 1, Body: []byte(`{"id":"message-1"}`)}

	err := consumer.process(context.Background(), delivery, func(context.Context, decodedMessage) error {
		panic("unexpected payload")
	})
	if err != nil || !published || !acknowledger.acknowledged {
		t.Fatalf("expected panic parked in DLQ, got err=%v published=%t ack=%t", err, published, acknowledger.acknowledged)
	}
}

func TestToPublishingPreservesMessageAndAddsFailureReason(t *testing.T) {
	createdAt := time.Now().UTC()
	delivery := amqp091.Delivery{
		Body: []byte("payload"), Headers: amqp091.Table{"trace-id": "trace-1"},
		ContentType: "application/json", DeliveryMode: amqp091.Persistent,
		CorrelationId: "correlation-1", Timestamp: createdAt,
	}

	published := toPublishing(context.Background(), delivery, errors.New("processing failed"))

	if string(published.Body) != "payload" || published.Headers["trace-id"] != "trace-1" {
		t.Fatalf("expected preserved delivery, got %#v", published)
	}
	if published.Headers["x-failure-reason"] != "processing failed" || published.CorrelationId != "correlation-1" {
		t.Fatalf("expected DLQ metadata, got %#v", published)
	}
	if !published.Timestamp.Equal(createdAt) || published.DeliveryMode != amqp091.Persistent {
		t.Fatalf("expected preserved AMQP properties, got %#v", published)
	}
}

func TestToPublishingMakesDeadLetterMessagePersistent(t *testing.T) {
	delivery := amqp091.Delivery{Body: []byte("payload"), DeliveryMode: amqp091.Transient}

	published := toPublishing(context.Background(), delivery, errors.New("processing failed"))

	if published.DeliveryMode != amqp091.Persistent {
		t.Fatalf("expected persistent dead-letter message, got %d", published.DeliveryMode)
	}
}
