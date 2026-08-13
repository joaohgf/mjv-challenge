package main

import (
	"net"
	"net/http"
	"testing"
	"time"
)

func TestShutdownWaitsForInFlightRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listening: %v", err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	server := newServer("", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-release
	}))
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String())
		if response != nil {
			response.Body.Close()
		}
		requestDone <- err
	}()
	<-started
	stopped := make(chan error, 1)
	go func() { stopped <- shutdown(server, time.Second, serverErrors) }()
	select {
	case err := <-stopped:
		t.Fatalf("shutdown finished before request: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	if err := <-requestDone; err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if err := <-stopped; err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}
}
