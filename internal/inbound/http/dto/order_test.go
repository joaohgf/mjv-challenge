package dto

import "testing"

func TestNewOrderResponseCreatesEmptyResponse(t *testing.T) {
	if response := NewOrderResponse(); response == nil || response.ID != "" {
		t.Fatalf("expected an empty order response, got %#v", response)
	}
}
