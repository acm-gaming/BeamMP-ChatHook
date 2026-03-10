package main

import "testing"

func TestNewUDPClientInvalidIP(t *testing.T) {
	if _, err := newUDPClient("invalid", 30813); err == nil {
		t.Fatalf("expected error for invalid ip")
	}
}

func TestNewUDPListenerInvalidIP(t *testing.T) {
	if _, err := newUDPListener("invalid", 30813); err == nil {
		t.Fatalf("expected error for invalid ip")
	}
}

func TestUDPClientRecvTimeout(t *testing.T) {
	client, err := newUDPClient("127.0.0.1", 1)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	value, ok, err := client.RecvString()
	if err != nil {
		t.Fatalf("recv string: %v", err)
	}
	if ok {
		t.Fatalf("expected no data, got %q", value)
	}
}

func TestUDPListenerRecvTimeout(t *testing.T) {
	listener, err := newUDPListener("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("create listener: %v", err)
	}
	defer listener.Close()

	value, source, ok, err := listener.RecvString()
	if err != nil {
		t.Fatalf("recv string: %v", err)
	}
	if ok {
		t.Fatalf("expected no data, got %q from %q", value, source)
	}
}
