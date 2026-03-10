package chathook

import (
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "env-webhook")
	t.Setenv("UDP_PORT", "4012")
	t.Setenv("AVATAR_URL", "env-avatar")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}
	if cfg.WebhookURL != "env-webhook" {
		t.Fatalf("expected env webhook, got %q", cfg.WebhookURL)
	}
	if cfg.UDPPort != 4012 {
		t.Fatalf("expected env udp port, got %d", cfg.UDPPort)
	}
	if cfg.AvatarURL != "env-avatar" {
		t.Fatalf("expected env avatar, got %q", cfg.AvatarURL)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "env-webhook")
	t.Setenv("UDP_PORT", "")
	t.Setenv("CHATHOOK_LOG_LEVEL", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}
	if cfg.UDPPort != 30813 {
		t.Fatalf("expected default udp port 30813, got %d", cfg.UDPPort)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected default log level info, got %q", cfg.LogLevel)
	}
}
