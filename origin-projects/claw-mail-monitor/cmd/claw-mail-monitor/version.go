package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kiry163/claw-mail-monitor/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}
