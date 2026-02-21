package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := os.ExpandEnv("$HOME/.config/claw-pliers/config.yaml")

		if _, err := os.Stat(configPath); err == nil && !forceInit {
			fmt.Printf("Config already exists at %s\n", configPath)
			fmt.Println("Use --overwrite to replace")
			return nil
		}

		dir := os.ExpandEnv("$HOME/.config/claw-pliers")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		configContent := `server:
  port: 8080
  log_level: info

auth:
  local_key: "change-me-in-production"

includes:
  - name: file
    path: "./file-config.yaml"
  - name: mail
    path: "./mail-config.yaml"
  - name: image
    path: "./image-config.yaml"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return err
		}

		fmt.Printf("Config initialized at %s\n", configPath)
		return nil
	},
}

var forceInit bool

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&forceInit, "overwrite", false, "Overwrite existing config")
}
