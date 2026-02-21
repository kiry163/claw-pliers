package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <folder_path>",
	Short: "Create a folder",
	Long: `Create a folder. Supports nested paths like a/b/c.

Examples:
  filehub-cli mkdir documents
  filehub-cli mkdir /documents
  filehub-cli mkdir /a/b/c`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		folderPath := args[0]
		folderPath = normalizePath(folderPath)
		folderPath = strings.TrimSuffix(folderPath, "/")

		parts := strings.Split(folderPath, "/")

		var parentID *string
		for i, part := range parts {
			if part == "" {
				continue
			}
			folder, err := client.CreateFolder(part, parentID)
			if err != nil {
				return fmt.Errorf("create folder %s failed: %w", part, err)
			}
			currentPath := "/" + strings.TrimPrefix(strings.Join(parts[:i+1], "/"), "/")
			fmt.Printf("✅ Created: 📁 %s/\n", currentPath)
			parentID = &folder.FolderID
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkdirCmd)
}
