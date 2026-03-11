package chathook

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/charmbracelet/log"
)

func TestSendWebhook(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		svc := NewService(Config{WebhookURL: server.URL}, log.New(nil))
		svc.httpClient = server.Client()

		err := svc.sendWebhook(context.Background(), webhookPayload{Content: "hello"})
		if err != nil {
			t.Fatalf("send webhook should succeed: %v", err)
		}
	})

	t.Run("non_2xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer server.Close()

		svc := NewService(Config{WebhookURL: server.URL}, log.New(nil))
		svc.httpClient = server.Client()

		err := svc.sendWebhook(context.Background(), webhookPayload{Content: "hello"})
		if err == nil || !strings.Contains(err.Error(), "bad request") {
			t.Fatalf("expected bad request error, got %v", err)
		}
	})
}

func TestProfilePictureCaching(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/u/Alice.json" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"user":{"avatar_template":"/user_avatar/forum.beammp.com/alice/{size}/1_2.png"}}`))
	}))
	defer server.Close()

	svc := NewService(Config{}, log.New(nil))
	svc.httpClient = server.Client()
	svc.forumBaseURL = server.URL

	first := svc.profilePicture("Alice")
	second := svc.profilePicture("Alice")

	if first == "" || second == "" {
		t.Fatalf("expected profile urls to be populated")
	}
	if first != second {
		t.Fatalf("expected cached value to match, first=%q second=%q", first, second)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one forum API call due to cache, got %d", calls.Load())
	}
	if !strings.Contains(first, "/144/") {
		t.Fatalf("expected avatar size replacement in %q", first)
	}
}

func TestIPAPIResult(t *testing.T) {
	t.Run("success_with_proxy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/json/8.8.8.8" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if fields := r.URL.Query().Get("fields"); fields != "status,countryCode,proxy,hosting" {
				t.Fatalf("unexpected fields query: %q", fields)
			}
			_, _ = w.Write([]byte(`{"status":"success","countryCode":"US","proxy":true,"hosting":false}`))
		}))
		defer server.Close()

		svc := NewService(Config{}, log.New(nil))
		svc.httpClient = server.Client()
		svc.ipAPIBaseURL = server.URL

		flag, isVPN := svc.ipAPIResult("8.8.8.8")
		if flag == "" {
			t.Fatalf("expected non-empty country flag")
		}
		if !isVPN {
			t.Fatalf("expected VPN/proxy detection to be true")
		}
	})

	t.Run("failed_status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"status":"fail"}`))
		}))
		defer server.Close()

		svc := NewService(Config{}, log.New(nil))
		svc.httpClient = server.Client()
		svc.ipAPIBaseURL = server.URL

		flag, isVPN := svc.ipAPIResult("1.1.1.1")
		if flag != "" || isVPN {
			t.Fatalf("expected empty/non-vpn for failed status, got flag=%q vpn=%v", flag, isVPN)
		}
	})
}

func TestHandlePacketBatchesScriptMessages(t *testing.T) {
	var requestCount atomic.Int32
	var payload webhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	svc := NewService(Config{WebhookURL: server.URL}, log.New(nil))
	svc.httpClient = server.Client()

	wire := packet{
		ServerName:  "Srv",
		PlayerCount: 2,
		PlayerMax:   8,
		PlayerDif:   1,
		Version:     ProtocolVersion,
		Contents: []content{
			{Type: 6, Content: json.RawMessage(`{"script_ref":"","chat_message":"hello @all"}`)},
			{Type: 6, Content: json.RawMessage(`{"script_ref":"Logic","chat_message":"started"}`)},
		},
	}

	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal wire packet: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(body)

	if err := svc.HandlePacket(context.Background(), []byte(encoded)); err != nil {
		t.Fatalf("handle packet: %v", err)
	}

	if requestCount.Load() != 1 {
		t.Fatalf("expected exactly one webhook call, got %d", requestCount.Load())
	}
	if payload.Username != "Srv" {
		t.Fatalf("unexpected username: %q", payload.Username)
	}
	expected := "hello all\n> - ⚙️ **Logic:** started"
	if payload.Content != expected {
		t.Fatalf("unexpected batched content:\nexpected: %q\nactual:   %q", expected, payload.Content)
	}
}
