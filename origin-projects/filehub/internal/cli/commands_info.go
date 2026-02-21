package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <path>",
	Short: "Show file or folder information",
	Long: `Show file or folder information.

Examples:
  filehub-cli info /test.txt
  filehub-cli info /documents/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		path := args[0]

		if path == "/" || path == "" {
			return showRootInfo(client)
		}

		if path[0] == '/' || path == "filehub:/" || path == "filehub://" {
			info, err := resolvePath(client, path)
			if err != nil {
				return err
			}
			if info.IsFile {
				return showFileInfo(client, info.FileID)
			}
			return showFolderInfo(client, info.FolderID)
		}

		return showFolderInfo(client, path)
	},
}

func showRootInfo(client *Client) error {
	contents, err := client.GetRootContents()
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		data, _ := json.MarshalIndent(contents, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Name:          📁 /\n")
	fmt.Printf("Path:          filehub:/\n")
	fmt.Printf("Folders:       %d\n", contents.Stats.FolderCount)
	fmt.Printf("Files:         %d\n", contents.Stats.FileCount)
	fmt.Printf("Total Size:    %s\n", formatSize(contents.Stats.TotalSize))
	return nil
}

func showFileInfo(client *Client, fileID string) error {
	file, err := client.GetFile(fileID)
	if err != nil {
		return fmt.Errorf("get file failed: %w", err)
	}

	if outputFormat == "json" {
		data, _ := json.MarshalIndent(file, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	path := file.Path
	if path == "" {
		path = "/"
	}
	fmt.Printf("Name:          %s\n", file.OriginalName)
	fmt.Printf("Path:          filehub:%s\n", path)
	fmt.Printf("Size:          %s\n", formatSize(file.Size))
	fmt.Printf("MIME Type:     %s\n", file.MimeType)
	fmt.Printf("Created:       %s\n", formatDate(file.CreatedAt))
	fmt.Printf("Download:      %s\n", file.DownloadURL)
	if file.ViewURL != "" {
		fmt.Printf("View:          %s\n", file.ViewURL)
	}
	return nil
}

func showFolderInfo(client *Client, folderIDOrPath string) error {
	var folderID string
	var err error

	if folderIDOrPath == "/" || folderIDOrPath == "" {
		contents, err := client.GetRootContents()
		if err != nil {
			return err
		}
		return showRootContents(contents)
	}

	info, err := resolvePath(client, folderIDOrPath)
	if err != nil {
		return err
	}
	folderID = info.FolderID

	contents, err := client.GetFolderContents(folderID)
	if err != nil {
		return fmt.Errorf("get folder contents failed: %w", err)
	}

	if outputFormat == "json" {
		data, _ := json.MarshalIndent(contents, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	folderPath := info.Path
	if folderPath == "" {
		folderPath = "/"
	}
	fmt.Printf("Name:          📁 %s\n", contents.Name)
	fmt.Printf("Path:          filehub:%s\n", folderPath)
	fmt.Printf("View:          %s\n", contents.FolderID)
	fmt.Printf("Folders:       %d\n", contents.Stats.FolderCount)
	fmt.Printf("Files:         %d\n", contents.Stats.FileCount)
	fmt.Printf("Total Size:    %s\n", formatSize(contents.Stats.TotalSize))
	return nil
}

func showRootContents(contents *FolderContents) error {
	fmt.Printf("Name:          📁 /\n")
	fmt.Printf("Path:          filehub:/\n")
	fmt.Printf("Folders:       %d\n", contents.Stats.FolderCount)
	fmt.Printf("Files:         %d\n", contents.Stats.FileCount)
	fmt.Printf("Total Size:    %s\n", formatSize(contents.Stats.TotalSize))
	return nil
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
