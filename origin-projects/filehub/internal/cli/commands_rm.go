package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Delete files or folders",
	Long: `Delete files or folders.

Examples:
  filehub-cli rm /test.txt
  filehub-cli rm /documents/
  filehub-cli rm /documents/ -r`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		recursive, _ := cmd.Flags().GetBool("recursive")
		path := args[0]

		if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "filehub:/") || strings.HasPrefix(path, "filehub://") {
			fileID, err := parseFilehubURL(client, path)
			if err != nil {
				return err
			}
			if err := client.DeleteFile(fileID); err != nil {
				return fmt.Errorf("delete file failed: %w", err)
			}
			fmt.Printf("🗑️  Deleted: %s\n", path)
			return nil
		}

		folder, err := client.GetFolderByPath(path)
		if err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}

		if !recursive {
			return fmt.Errorf("cannot delete non-empty folder (use -r flag)")
		}

		if err := client.DeleteFolder(folder.FolderID, recursive); err != nil {
			return fmt.Errorf("delete folder failed: %w", err)
		}
		fmt.Printf("🗑️  Deleted: %s/\n", path)
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolP("recursive", "r", false, "Delete folders recursively")
	rootCmd.AddCommand(rmCmd)
}
