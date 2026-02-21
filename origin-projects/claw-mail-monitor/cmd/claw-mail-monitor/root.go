package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	baseURL    string
)

var rootCmd = &cobra.Command{
	Use:   "claw-mail-monitor",
	Short: "Claw Mail Monitor HTTP/CLI",
}

func Execute() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(testConnCmd)
	rootCmd.AddCommand(latestCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "http://127.0.0.1:14630", "HTTP API base URL")

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
