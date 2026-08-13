package domain

import "testing"

func TestNewOrderCreatesEmptyOrder(t *testing.T) {
	if order := NewOrder(); order == nil || order.ID != "" {
		t.Fatalf("expected an empty order, got %#v", order)
	}
}
