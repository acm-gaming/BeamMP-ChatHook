package chathook

import (
	"fmt"
	"strings"

	"github.com/num30/config"
)

type Config struct {
	WebhookURL string `envvar:"WEBHOOK_URL"`
	UDPPort    int    `envvar:"UDP_PORT" default:"30813"`
	AvatarURL  string `envvar:"AVATAR_URL"`
	LogLevel   string `envvar:"CHATHOOK_LOG_LEVEL" default:"info"`
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
	cfg.LogLevel = strings.ToLower(strings.TrimSpace(cfg.LogLevel))

	return cfg, nil
}
