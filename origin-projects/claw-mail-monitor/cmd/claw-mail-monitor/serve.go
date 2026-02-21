package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kiry163/claw-mail-monitor/internal/config"
	"github.com/kiry163/claw-mail-monitor/internal/httpapi"
	"github.com/kiry163/claw-mail-monitor/internal/imap"
	"github.com/kiry163/claw-mail-monitor/internal/logging"
	"github.com/kiry163/claw-mail-monitor/internal/version"
	"github.com/kiry163/claw-mail-monitor/internal/webhook"
)

var listenAddr string
var pollInterval string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server and email monitoring",
	RunE: func(cmd *cobra.Command, args []string) error {
		if configPath != "" {
			_ = os.Setenv("CLAW_MAIL_MONITOR_CONFIG", configPath)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if pollInterval != "" {
			cfg.Monitoring.PollInterval = pollInterval
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		stop := handleSignals(cancel)
		defer stop()

		return runServer(ctx, cfg, listenAddr)
	},
}

func init() {
	serveCmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:14630", "HTTP listen address")
	serveCmd.Flags().StringVar(&pollInterval, "poll-interval", "", "polling interval (e.g. 30s)")
}

func handleSignals(cancel context.CancelFunc) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-done:
		}
	}()

	return func() {
		close(done)
		signal.Stop(sigCh)
	}
}

func runServer(ctx context.Context, cfg *config.Config, addr string) error {
	logging.Init(cfg.Logging)

	webhookClient := webhook.NewClient(&cfg.Webhook)
	monitorManager := imap.NewManager(cfg, webhookClient)
	monitorManager.StartAll(ctx)

	slog.Info("http server starting", "addr", addr)
	server := httpapi.NewServer(addr, cfg, monitorManager, version.Version)
	return server.Run(ctx)
}
