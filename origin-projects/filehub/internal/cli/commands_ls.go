package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List files and folders",
	Long: `List files and folders.

Examples:
  filehub-cli ls
  filehub-cli ls /
  filehub-cli ls /documents/
  filehub-cli ls documents/
  filehub-cli ls -d`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		dirsOnly, _ := cmd.Flags().GetBool("dirs")
		longFormat, _ := cmd.Flags().GetBool("long")

		path := ""
		if len(args) > 0 {
			path = args[0]
		}

		var folderID *string
		if path != "" {
			id, err := parsePath(client, path)
			if err != nil {
				return fmt.Errorf("resolve folder path failed: %w", err)
			}
			if id != "" {
				folderID = &id
			}
		}

		contents, err := getFolderContents(client, folderID)
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(contents, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(contents.Folders) == 0 && len(contents.Files) == 0 {
			fmt.Println("(empty)")
			return nil
		}

		if !dirsOnly && len(contents.Folders) > 0 {
			for _, folder := range contents.Folders {
				printFolderItem(folder)
			}
		}

		if !dirsOnly && len(contents.Files) > 0 {
			for _, file := range contents.Files {
				printFileItem(file)
			}
		}

		if dirsOnly {
			for _, folder := range contents.Folders {
				printFolderItem(folder)
			}
		}

		if longFormat && len(contents.Files) > 0 {
			var totalSize int64
			for _, f := range contents.Files {
				totalSize += f.Size
			}
			fmt.Printf("\n%d files, %s\n", len(contents.Files), formatSize(totalSize))
		}

		return nil
	},
}

func getFolderContents(client *Client, folderID *string) (*FolderContents, error) {
	if folderID == nil {
		return client.GetFolderContents("")
	}
	return client.GetFolderContents(*folderID)
}

func init() {
	lsCmd.Flags().BoolP("dirs", "d", false, "List directories only")
	lsCmd.Flags().BoolP("long", "l", false, "Long format with statistics")
	rootCmd.AddCommand(lsCmd)
}
