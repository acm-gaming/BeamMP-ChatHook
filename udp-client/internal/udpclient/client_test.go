package udpclient

import (
	"bytes"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestRunWithArgumentPayload(t *testing.T) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer listener.Close()

	readBuffer := make([]byte, 1024)
	if err := listener.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	port := listener.LocalAddr().(*net.UDPAddr).Port
	config := Config{BindStart: 3400, BindEnd: 3499}
	if err := Run([]string{"127.0.0.1", strconv.Itoa(port), "payload"}, bytes.NewReader(nil), config); err != nil {
		t.Fatalf("run helper: %v", err)
	}

	readBytes, _, err := listener.ReadFromUDP(readBuffer)
	if err != nil {
		t.Fatalf("read udp payload: %v", err)
	}
	if got := string(readBuffer[:readBytes]); got != "payload" {
		t.Fatalf("unexpected payload %q", got)
	}
}

func TestReadPayloadFallsBackToStdin(t *testing.T) {
	payload, err := ReadPayload([]string{"127.0.0.1", "30813"}, bytes.NewBufferString("stdin-data"))
	if err != nil {
		t.Fatalf("read payload: %v", err)
	}
	if got := string(payload); got != "stdin-data" {
		t.Fatalf("unexpected payload %q", got)
	}
}
