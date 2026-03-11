package chathook

import (
	"fmt"
	"strings"

	"github.com/num30/config"
)

type Config struct {
	WebhookURL             string `envvar:"WEBHOOK_URL"`
	UDPPort                int    `envvar:"UDP_PORT" default:"30813"`
	AvatarURL              string `envvar:"AVATAR_URL"`
	LogLevel               string `envvar:"CHATHOOK_LOG_LEVEL" default:"info"`
	ChatRateLimitCount     int    `envvar:"CHATHOOK_CHAT_RATE_LIMIT_COUNT" default:"6"`
	ChatRateLimitWindowSec int    `envvar:"CHATHOOK_CHAT_RATE_LIMIT_WINDOW_SEC" default:"10"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := config.NewConfReader("chathook-daemon").Read(&cfg); err != nil {
		return Config{}, err
	}

	if strings.TrimSpace(cfg.WebhookURL) == "" {
		return Config{}, fmt.Errorf("webhook url is required (set WEBHOOK_URL)")
	}
	if cfg.UDPPort < 1 || cfg.UDPPort > 65535 {
		return Config{}, fmt.Errorf("udp port must be between 1 and 65535")
	}
	if cfg.ChatRateLimitCount < 0 {
		return Config{}, fmt.Errorf("chat rate limit count must be >= 0")
	}
	if cfg.ChatRateLimitCount > 0 && (cfg.ChatRateLimitWindowSec < 1 || cfg.ChatRateLimitWindowSec > 3600) {
		return Config{}, fmt.Errorf("chat rate limit window must be between 1 and 3600 seconds when enabled")
	}
	cfg.LogLevel = strings.ToLower(strings.TrimSpace(cfg.LogLevel))

	return cfg, nil
}
