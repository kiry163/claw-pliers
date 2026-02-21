package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

func resolveFolderPath(client *Client, path string) (string, error) {
	if path == "" || path == "/" {
		return "", nil
	}
	path = normalizePath(path)
	path = strings.TrimSuffix(path, "/")
	resp, err := client.GetFolderByPath(path)
	if err != nil {
		return "", err
	}
	return resp.FolderID, nil
}

func normalizePath(path string) string {
	path = strings.TrimPrefix(path, "filehub:/")
	path = strings.TrimPrefix(path, "filehub://")
	return path
}

type FilePathInfo struct {
	FileID   string
	Path     string
	IsFile   bool
	IsFolder bool
	FolderID string
	Folder   *FolderItem
	File     *FileItem
}

func resolvePath(client *Client, value string) (*FilePathInfo, error) {
	value = normalizePath(value)
	value = strings.TrimSuffix(value, "/")

	if value == "" {
		return &FilePathInfo{
			IsFile:   false,
			IsFolder: true,
			Path:     "/",
		}, nil
	}

	path := "/" + value

	file, err := client.GetFileByPath(path)
	if err == nil {
		return &FilePathInfo{
			FileID:   file.FileID,
			Path:     path,
			IsFile:   true,
			IsFolder: false,
			File:     file,
		}, nil
	}

	folder, err := client.GetFolderByPath(path)
	if err == nil {
		folderItem := FolderItem{
			FolderID: folder.FolderID,
			Name:     folder.Name,
		}
		return &FilePathInfo{
			Path:     path,
			IsFile:   false,
			IsFolder: true,
			FolderID: folder.FolderID,
			Folder:   &folderItem,
		}, nil
	}

	return nil, fmt.Errorf("path not found: %s", value)
}

func parseFilehubURL(client *Client, value string) (string, error) {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "filehub:/") || strings.HasPrefix(value, "filehub://") {
		info, err := resolvePath(client, value)
		if err != nil {
			return "", err
		}
		if !info.IsFile {
			return "", fmt.Errorf("path is a folder, not a file")
		}
		return info.FileID, nil
	}

	if strings.HasPrefix(value, "/") {
		info, err := resolvePath(client, value)
		if err != nil {
			return "", err
		}
		if !info.IsFile {
			return "", fmt.Errorf("path is a folder, not a file")
		}
		return info.FileID, nil
	}

	if strings.Contains(value, "/") || strings.HasSuffix(value, "/") {
		return "", fmt.Errorf("invalid path format: %s", value)
	}

	value = strings.TrimPrefix(value, "filehub://")
	if value == "" {
		return "", fmt.Errorf("invalid filehub URL")
	}
	return value, nil
}

func parsePath(client *Client, value string) (string, error) {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "filehub:/") || strings.HasPrefix(value, "filehub://") {
		info, err := resolvePath(client, value)
		if err != nil {
			return "", err
		}
		if info.IsFolder {
			return info.FolderID, nil
		}
		return "", fmt.Errorf("path is a file, not a folder")
	}

	if strings.HasPrefix(value, "/") || value == "" {
		info, err := resolvePath(client, value)
		if err != nil {
			return "", err
		}
		if info.IsFolder {
			return info.FolderID, nil
		}
		return "", fmt.Errorf("path is a file, not a folder")
	}

	folder, err := client.GetFolderByPath(value)
	if err == nil {
		return folder.FolderID, nil
	}

	return "", fmt.Errorf("folder not found: %s", value)
}

func printFileItem(item FileItem) {
	path := item.Path
	if path == "" {
		path = "/"
	}
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(item, "", "  ")
		fmt.Println(string(data))
	case "short":
		fmt.Printf("📄 %s %s\n", path, item.OriginalName)
	default:
		date := formatDate(item.CreatedAt)
		fmt.Printf("📄 %-35s %8s  %s\n", path, formatSize(item.Size), date)
	}
}

func printFolderItem(folder FolderItem) {
	path := folder.Path
	if path == "" {
		path = "/"
	}
	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(folder, "", "  ")
		fmt.Println(string(data))
	default:
		fmt.Printf("📁 %s/\n", path)
	}
}

func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
}

func formatDate(date string) string {
	if len(date) >= 10 {
		return date[:10]
	}
	return date
}
