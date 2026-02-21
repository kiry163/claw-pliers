package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put <local_path...> [folder]",
	Short: "Upload files to FileHub",
	Long: `Upload files to FileHub.

Examples:
  filehub-cli put file.txt
  filehub-cli put file.txt /
  filehub-cli put file.txt /documents/
  filehub-cli put file.txt documents/
  filehub-cli put folder/ -r`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		recursive, _ := cmd.Flags().GetBool("recursive")
		targetPath := ""
		var paths []string

		if len(args) > 1 {
			targetPath = args[len(args)-1]
			paths = args[:len(args)-1]
		} else {
			paths = args
		}

		var folderID *string
		if targetPath != "" {
			cleanPath := normalizePath(targetPath)
			folderIDStr, err := resolveFolderPath(client, cleanPath)
			if err != nil {
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
					parts := strings.Split(strings.TrimSuffix(cleanPath, "/"), "/")
					if len(parts) == 1 && parts[0] == "" {
						folderID = nil
					} else {
						var parentID *string
						for _, part := range parts {
							if part == "" {
								continue
							}
							folder, err := client.CreateFolder(part, parentID)
							if err != nil {
								return fmt.Errorf("create folder %s failed: %w", part, err)
							}
							fmt.Printf("Created folder: %s\n", part)
							parentID = &folder.FolderID
						}
						if parentID != nil {
							folderID = parentID
						}
					}
				} else {
					return fmt.Errorf("resolve folder path failed: %w", err)
				}
			} else if folderIDStr != "" {
				folderID = &folderIDStr
			}
		}

		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s: %v\n", path, err)
				continue
			}

			if info.IsDir() && recursive {
				if err := uploadDirectory(client, path, folderID); err != nil {
					fmt.Fprintf(os.Stderr, "Error uploading %s: %v\n", path, err)
				}
			} else if !info.IsDir() {
				if err := uploadFile(client, path, folderID); err != nil {
					fmt.Fprintf(os.Stderr, "Error uploading %s: %v\n", path, err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Skipping directory %s (use -r flag)\n", path)
			}
		}
		return nil
	},
}

func uploadFile(client *Client, path string, folderID *string) error {
	fmt.Printf("Uploading %s...\n", path)
	file, err := client.UploadFile(path, folderID, func(progress int) {
		fmt.Printf("\rProgress: %d%%", progress)
	})
	if err != nil {
		return err
	}
	fmt.Printf("\nUploaded: 📄 /%s -> filehub:/%s\n", file.OriginalName, file.Path)
	return nil
}

func uploadDirectory(client *Client, dirPath string, parentFolderID *string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	currentFolderID := parentFolderID
	if parentFolderID == nil {
		folderName := filepath.Base(dirPath)
		folder, err := client.CreateFolder(folderName, nil)
		if err != nil {
			return fmt.Errorf("create folder %s failed: %w", folderName, err)
		}
		currentFolderID = &folder.FolderID
		fmt.Printf("Created folder: 📁 %s\n", folderName)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			folder, err := client.CreateFolder(entry.Name(), currentFolderID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating folder %s: %v\n", entry.Name(), err)
				continue
			}
			if err := uploadDirectory(client, fullPath, &folder.FolderID); err != nil {
				fmt.Fprintf(os.Stderr, "Error uploading %s: %v\n", fullPath, err)
			}
		} else {
			if err := uploadFile(client, fullPath, currentFolderID); err != nil {
				fmt.Fprintf(os.Stderr, "Error uploading %s: %v\n", fullPath, err)
			}
		}
	}
	return nil
}

func init() {
	putCmd.Flags().BoolP("recursive", "r", false, "Upload directories recursively")
	rootCmd.AddCommand(putCmd)
}
