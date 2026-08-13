package model

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestOrderUsesRequiredMongoDocumentFields(t *testing.T) {
	updatedAt := time.Now().UTC()
	source := &Order{
		ID: "order-1", ProductName: "Caderno", Status: "CRIADO", Quantity: 2,
		CreatedAt: updatedAt, UpdatedAt: &updatedAt,
	}
	body, err := bson.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	var document bson.M
	if err := bson.Unmarshal(body, &document); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"order_id", "product", "quantity", "status", "created_at", "updatedAt"} {
		if _, ok := document[field]; !ok {
			t.Fatalf("expected field %q in document: %#v", field, document)
		}
	}
	if source.GetIDField() != "order_id" || source.GetID() != "order-1" {
		t.Fatalf("expected order identifier metadata, got field=%q id=%q", source.GetIDField(), source.GetID())
	}
}

func TestNewOrderCreatesEmptyDocument(t *testing.T) {
	if order := NewOrder(); order == nil || order.ID != "" {
		t.Fatalf("expected an empty order document, got %#v", order)
	}
}
