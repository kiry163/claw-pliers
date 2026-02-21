package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var mvCmd = &cobra.Command{
	Use:   "mv <source> <destination>",
	Short: "Move or rename files and folders",
	Long: `Move or rename files and folders.

Examples:
  filehub-cli mv /test.txt /documents/
  filehub-cli mv filehub:/test.txt /documents/
  filehub-cli mv /folder/ /newfolder/
  filehub-cli mv /file.txt newname.txt`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}
		client := NewClient(cfg)

		src := args[0]
		dst := args[1]

		if strings.HasPrefix(src, "/") || strings.HasPrefix(src, "filehub:/") || strings.HasPrefix(src, "filehub://") {
			return moveFile(client, src, dst)
		}

		return moveFolder(client, src, dst)
	},
}

func moveFile(client *Client, srcURL, dst string) error {
	fileID, err := parseFilehubURL(client, srcURL)
	if err != nil {
		return err
	}

	file, err := client.GetFile(fileID)
	if err != nil {
		return fmt.Errorf("get file failed: %w", err)
	}

	dst = normalizePath(dst)
	dst = strings.TrimSuffix(dst, "/")

	if strings.Contains(dst, "/") || (dst != "" && !strings.Contains(dst, ".")) {
		folderID, err := resolveFolderPath(client, dst)
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("destination folder not found: %s", dst)
			}
			return fmt.Errorf("resolve destination path failed: %w", err)
		}
		var folderIDPtr *string
		if folderID != "" {
			folderIDPtr = &folderID
		}
		if err := client.MoveFile(fileID, folderIDPtr); err != nil {
			return fmt.Errorf("move file failed: %w", err)
		}
		fmt.Printf("Moved 📄 %s -> 📁 %s/\n", file.OriginalName, dst)
		return nil
	}

	fmt.Printf("Note: File rename is not supported by server. Moving to root instead.\n")
	if err := client.MoveFile(fileID, nil); err != nil {
		return fmt.Errorf("move file failed: %w", err)
	}
	fmt.Printf("Moved 📄 %s -> /\n", file.OriginalName)
	return nil
}

func moveFolder(client *Client, srcPath, dstPath string) error {
	srcPath = normalizePath(srcPath)
	dstPath = normalizePath(dstPath)
	srcPath = strings.TrimSuffix(srcPath, "/")
	dstPath = strings.TrimSuffix(dstPath, "/")

	srcFolder, err := client.GetFolderByPath(srcPath)
	if err != nil {
		return fmt.Errorf("source folder not found: %w", err)
	}

	if strings.Contains(dstPath, "/") {
		parts := strings.Split(dstPath, "/")
		parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
		newName := parts[len(parts)-1]

		parentFolder, err := client.GetFolderByPath(parentPath)
		if err != nil {
			return fmt.Errorf("destination parent folder not found: %w", err)
		}

		if err := client.MoveFolder(srcFolder.FolderID, &parentFolder.FolderID); err != nil {
			return fmt.Errorf("move folder failed: %w", err)
		}
		if err := client.RenameFolder(srcFolder.FolderID, newName); err != nil {
			return fmt.Errorf("rename folder failed: %w", err)
		}
		fmt.Printf("Moved 📁 %s -> 📁 %s/\n", srcPath, dstPath)
		return nil
	}

	if err := client.RenameFolder(srcFolder.FolderID, dstPath); err != nil {
		return fmt.Errorf("rename folder failed: %w", err)
	}
	fmt.Printf("Renamed 📁 %s -> %s\n", srcFolder.Name, dstPath)
	return nil
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
