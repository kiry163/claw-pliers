package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "claw-pliers",
	Short: "Unified CLI for file, mail, and image services",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func main() {
	rootCmd.AddCommand(fileCmd)
	rootCmd.AddCommand(mailCmd)
	rootCmd.AddCommand(imageCmd)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
