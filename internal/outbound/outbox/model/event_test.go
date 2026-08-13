package model

import (
	"testing"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/enum"
	"go.mongodb.org/mongo-driver/bson"
)

func TestEventPersistsLeaseToken(t *testing.T) {
	now := time.Now().UTC()
	source := Event[string]{
		ID: "event-1", Payload: "order-1", Status: enum.OutboxProcessing,
		CreatedAt: now, UpdatedAt: now, LeaseToken: "lease-1",
	}
	body, err := bson.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	var document bson.M
	if err := bson.Unmarshal(body, &document); err != nil {
		t.Fatal(err)
	}
	if document["lease_token"] != "lease-1" {
		t.Fatalf("expected persisted lease token, got %#v", document)
	}
}
