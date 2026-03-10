package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/num30/config"
	"github.com/spf13/cobra"

	"github.com/acm-gaming/beammp-chathook/udp-client/internal/udpclient"
)

type clientConfig struct {
	BindStart int    `default:"3400" envvar:"UDP_CLIENT_BIND_START"`
	BindEnd   int    `default:"3499" envvar:"UDP_CLIENT_BIND_END"`
	LogLevel  string `default:"info" envvar:"UDP_CLIENT_LOG_LEVEL"`
}

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: true})
	cmd := newRootCmd(logger)
	if err := cmd.Execute(); err != nil {
		logger.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func newRootCmd(logger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udp-client <ip> <port> [payload]",
		Short: "Send a UDP payload to the ChatHook daemon",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig(cmd)
			applyLogLevel(logger, cfg.LogLevel)
			return udpclient.Run(args, os.Stdin, udpclient.Config{
				BindStart: cfg.BindStart,
				BindEnd:   cfg.BindEnd,
			})
		},
	}

	cmd.Flags().Int("bind-start", 3400, "First UDP source port to try")
	cmd.Flags().Int("bind-end", 3499, "Last UDP source port to try")
	cmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")

	return cmd
}

func loadConfig(cmd *cobra.Command) clientConfig {
	var cfg clientConfig
	// num30/config parses os.Args for flags; Cobra already handled CLI flags.
	savedArgs := os.Args
	os.Args = []string{savedArgs[0]}
	_ = config.NewConfReader("udp-client").Read(&cfg)
	os.Args = savedArgs

	if cmd != nil {
		if cmd.Flags().Changed("bind-start") {
			value, _ := cmd.Flags().GetInt("bind-start")
			cfg.BindStart = value
		}
		if cmd.Flags().Changed("bind-end") {
			value, _ := cmd.Flags().GetInt("bind-end")
			cfg.BindEnd = value
		}
		if cmd.Flags().Changed("log-level") {
			value, _ := cmd.Flags().GetString("log-level")
			cfg.LogLevel = value
		}
	}

	return cfg
}

func applyLogLevel(logger *log.Logger, level string) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "warn", "warning":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}
}
