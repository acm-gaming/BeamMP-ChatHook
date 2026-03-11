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
	if cfg.ChatRateLimitCount != 6 {
		t.Fatalf("expected default chat rate limit count 6, got %d", cfg.ChatRateLimitCount)
	}
	if cfg.ChatRateLimitWindowSec != 10 {
		t.Fatalf("expected default chat rate limit window 10, got %d", cfg.ChatRateLimitWindowSec)
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

func TestLoadConfigChatRateLimitOverride(t *testing.T) {
	t.Setenv("WEBHOOK_URL", "env-webhook")
	t.Setenv("CHATHOOK_CHAT_RATE_LIMIT_COUNT", "12")
	t.Setenv("CHATHOOK_CHAT_RATE_LIMIT_WINDOW_SEC", "30")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}
	if cfg.ChatRateLimitCount != 12 {
		t.Fatalf("expected rate limit count 12, got %d", cfg.ChatRateLimitCount)
	}
	if cfg.ChatRateLimitWindowSec != 30 {
		t.Fatalf("expected rate limit window 30, got %d", cfg.ChatRateLimitWindowSec)
	}
}
