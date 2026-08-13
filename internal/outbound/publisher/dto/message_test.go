package dto

import "testing"

func TestNewMessageCreatesEmptyEnvelope(t *testing.T) {
	if message := NewMessage[string](); message == nil || message.Type != "" {
		t.Fatalf("expected an empty message envelope, got %#v", message)
	}
}
