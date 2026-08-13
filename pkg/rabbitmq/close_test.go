package adapter

import "testing"

func TestClientCloseIsIdempotentWithoutResources(t *testing.T) {
	rabbit := new(Client)

	if err := rabbit.Close(); err != nil {
		t.Fatalf("expected first empty close to succeed, got %v", err)
	}
	if err := rabbit.Close(); err != nil {
		t.Fatalf("expected second empty close to succeed, got %v", err)
	}
}

func TestClientWithoutResourcesIsClosed(t *testing.T) {
	if !new(Client).IsClosed() {
		t.Fatal("expected disconnected rabbitmq adapter to be closed")
	}
}
