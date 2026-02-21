package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"

	"github.com/kiry163/claw-mail-monitor/internal/config"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage system service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return controlService("install")
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return controlService("uninstall")
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return controlService("start")
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return controlService("stop")
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return controlService("restart")
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return controlService("status")
	},
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
}

type serviceProgram struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *serviceProgram) Start(_ service.Service) error {
	p.ctx, p.cancel = context.WithCancel(context.Background())
	go p.run()
	return nil
}

func (p *serviceProgram) Stop(_ service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *serviceProgram) run() {
	if configPath != "" {
		_ = os.Setenv("CLAW_MAIL_MONITOR_CONFIG", configPath)
	}

	cfg, err := config.Load()
	if err != nil {
		return
	}

	if pollInterval != "" {
		cfg.Monitoring.PollInterval = pollInterval
	}

	ctx, cancel := context.WithCancel(p.ctx)
	defer cancel()

	_ = runServer(ctx, cfg, listenAddr)
}

func controlService(action string) error {
	if configPath != "" {
		_ = os.Setenv("CLAW_MAIL_MONITOR_CONFIG", configPath)
	}

	if configPath == "" {
		if resolved, err := config.DefaultConfigPath(); err == nil {
			configPath = resolved
		}
	}

	if listenAddr == "" {
		listenAddr = "127.0.0.1:14630"
	}

	args := []string{"serve", "--listen", listenAddr}
	if pollInterval != "" {
		args = append(args, "--poll-interval", pollInterval)
	}

	serviceConfig := &service.Config{
		Name:        "claw-mail-monitor",
		DisplayName: "Claw Mail Monitor",
		Description: "Multi-account email monitoring via HTTP",
		EnvVars:     map[string]string{"CLAW_MAIL_MONITOR_CONFIG": configPath},
		Arguments:   args,
	}

	program := &serviceProgram{}
	svc, err := service.New(program, serviceConfig)
	if err != nil {
		return err
	}

	if action == "status" {
		status, err := svc.Status()
		if err != nil {
			return err
		}
		fmt.Println(statusString(status))
		return nil
	}

	if err := service.Control(svc, action); err != nil {
		return err
	}

	if action == "install" {
		fmt.Println("service installed")
	}
	if action == "uninstall" {
		fmt.Println("service uninstalled")
	}
	if action == "start" {
		fmt.Println("service started")
	}
	if action == "stop" {
		fmt.Println("service stopped")
	}
	if action == "restart" {
		fmt.Println("service restarted")
	}

	return nil
}

func statusString(status service.Status) string {
	switch status {
	case service.StatusRunning:
		return "running"
	case service.StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

func init() {
	serviceCmd.PersistentFlags().StringVar(&listenAddr, "listen", "127.0.0.1:14630", "HTTP listen address")
	serviceCmd.PersistentFlags().StringVar(&pollInterval, "poll-interval", "", "polling interval (e.g. 30s)")
}
