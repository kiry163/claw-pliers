package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <path> [local_path]",
	Short: "Download files from FileHub",
	Long: `Download files from FileHub.

Examples:
  filehub-cli get /test.txt
  filehub-cli get /documents/file.pdf
  filehub-cli get /file.pdf ./downloads/
  filehub-cli get /file.pdf myfile.txt`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		url := args[0]
		var fileID string
		var originalName string

		if strings.HasPrefix(url, "/") || strings.HasPrefix(url, "filehub:/") {
			info, err := resolvePath(client, url)
			if err != nil {
				return err
			}
			if !info.IsFile {
				return fmt.Errorf("path is a folder, not a file")
			}
			fileID = info.FileID
			originalName = info.File.OriginalName
		} else {
			var err error
			fileID, err = parseFilehubURL(client, url)
			if err != nil {
				return err
			}
			file, _ := client.GetFile(fileID)
			if file != nil {
				originalName = file.OriginalName
			}
		}

		outputPath := ""
		if len(args) > 1 {
			outputPath = args[1]
		}

		fmt.Printf("Downloading 📄 %s...\n", originalName)
		path, err := client.DownloadFile(fileID, outputPath, func(progress int) {
			fmt.Printf("\rProgress: %d%%", progress)
		})
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		fmt.Printf("\n✅ Saved to: %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
