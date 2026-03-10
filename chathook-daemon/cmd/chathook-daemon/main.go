package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/acm-gaming/beammp-chathook/chathook-daemon/internal/chathook"
)

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
		Use:   "chathook-daemon",
		Short: "BeamMP ChatHook UDP listener and Discord webhook bridge",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := chathook.LoadConfig()
			if err != nil {
				return err
			}
			chathook.ApplyLogLevel(logger, cfg.LogLevel)

			service := chathook.NewService(cfg, logger)
			logger.Info(
				"ChatHook daemon starting",
				"version", chathook.Version,
				"protocol", chathook.ProtocolVersion,
				"udp_port", cfg.UDPPort,
			)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := service.SendStartupHello(ctx); err != nil {
				logger.Warn("startup hello failed", "error", err)
			}

			if err := service.Listen(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}

			return nil
		},
	}

	return cmd
}
