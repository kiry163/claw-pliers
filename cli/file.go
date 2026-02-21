package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	endpoint string
	localKey string
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "File management commands",
}

type Config struct {
	Endpoint string `yaml:"endpoint"`
	LocalKey string `yaml:"local_key"`
}

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type FileItem struct {
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	Path         string `json:"path,omitempty"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	CreatedAt    string `json:"created_at"`
}

type FolderItem struct {
	FolderID  string `json:"folder_id"`
	Name      string `json:"name"`
	ParentID  string `json:"parent_id,omitempty"`
	Path      string `json:"path,omitempty"`
	CreatedAt string `json:"created_at"`
}

type FileListResponse struct {
	Total int        `json:"total"`
	Items []FileItem `json:"items"`
}

type FolderListResponse struct {
	Total   int          `json:"total"`
	Folders []FolderItem `json:"folders"`
}

type Client struct {
	Endpoint string
	LocalKey string
	HTTP     *http.Client
}

func loadConfig() (Config, error) {
	projectConfig := "./config/config.yaml"
	if _, err := os.Stat(projectConfig); err == nil {
		return loadConfigFromPath(projectConfig)
	}

	configDir := os.Getenv("CLAWPLIERS_CONFIG_DIR")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, err
		}
		configDir = filepath.Join(home, ".config", "claw-pliers")
	}

	userConfig := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(userConfig); err == nil {
		return loadConfigFromPath(userConfig)
	}

	return Config{}, errors.New("config file not found: ./config/config.yaml or ~/.config/claw-pliers/config.yaml")
}

func loadConfigFromPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, err
	}

	endpoint := ""
	localKey := ""

	if server, ok := raw["server"].(map[string]interface{}); ok {
		if port, ok := server["port"].(int); ok {
			endpoint = fmt.Sprintf("http://localhost:%d", port)
		} else if portFloat, ok := server["port"].(float64); ok {
			endpoint = fmt.Sprintf("http://localhost:%d", int(portFloat))
		}
	}

	if auth, ok := raw["auth"].(map[string]interface{}); ok {
		if lk, ok := auth["local_key"].(string); ok {
			localKey = lk
		}
	}

	if envEndpoint := os.Getenv("CLAWPLIERS_ENDPOINT"); envEndpoint != "" {
		endpoint = envEndpoint
	}

	if envKey := os.Getenv("CLAWPLIERS_AUTH_LOCAL_KEY"); envKey != "" {
		localKey = envKey
	}

	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}

	return Config{Endpoint: endpoint, LocalKey: localKey}, nil
}

func NewClient(cfg Config) *Client {
	return &Client{
		Endpoint: strings.TrimRight(cfg.Endpoint, "/"),
		LocalKey: cfg.LocalKey,
		HTTP: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (c *Client) attachAuth(req *http.Request) {
	if c.LocalKey != "" {
		req.Header.Set("X-Local-Key", c.LocalKey)
	}
}

func parseRemotePath(path string) (string, error) {
	path = strings.TrimPrefix(path, "claw:")
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return "/", nil
	}
	return "/" + path, nil
}

func validateRemotePath(path string) error {
	if !strings.HasPrefix(path, "claw:/") && !strings.HasPrefix(path, "/") && path != "" {
		return errors.New("路径必须以 claw:/ 开头")
	}
	return nil
}

type ProgressReader struct {
	reader   io.Reader
	total    int64
	current  int64
	lastPct  int
	progress func(int)
}

func NewProgressReader(reader io.Reader, total int64, progress func(int)) *ProgressReader {
	return &ProgressReader{
		reader:   reader,
		total:    total,
		current:  0,
		lastPct:  -1,
		progress: progress,
	}
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.current += int64(n)

	if pr.total > 0 && pr.progress != nil {
		pct := int(float64(pr.current) / float64(pr.total) * 100)
		if pct != pr.lastPct {
			pr.lastPct = pct
			pr.progress(pct)
		}
	}
	return n, err
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func printFileTable(items []FileItem, folders []FolderItem) {
	allItems := make([]string, 0)

	for _, f := range folders {
		allItems = append(allItems, fmt.Sprintf("\033[34m%s/\033[0m", f.Name))
	}
	for _, item := range items {
		allItems = append(allItems, item.OriginalName)
	}

	if len(allItems) == 0 {
		fmt.Println("(empty)")
		return
	}

	for _, item := range allItems {
		fmt.Println(item)
	}
}

func prompt(label, fallback string) string {
	reader := bufio.NewReader(os.Stdin)
	if fallback != "" {
		fmt.Printf("%s [%s]: ", label, fallback)
	} else {
		fmt.Printf("%s: ", label)
	}
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return fallback
	}
	return text
}

// ============ Commands ============

var fileLsCmd = &cobra.Command{
	Use:   "ls [claw:/path]",
	Short: "List directory contents",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := "/"
		if len(args) > 0 {
			remotePath = args[0]
			if err := validateRemotePath(remotePath); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			p, err := parseRemotePath(remotePath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			remotePath = p
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		folders, files, err := client.List(remotePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		printFileTable(files, folders)
		return nil
	},
}

var fileMkdirCmd = &cobra.Command{
	Use:   "mkdir <name> claw:/[path]",
	Short: "Create a directory",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		remotePath := "/"

		if len(args) > 1 {
			remotePath = args[1]
			if err := validateRemotePath(remotePath); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			p, err := parseRemotePath(remotePath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			remotePath = p
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		fullPath := remotePath
		if remotePath == "/" {
			fullPath = "/" + name
		} else {
			fullPath = remotePath + "/" + name
		}

		err := client.CreateFolder(fullPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		fmt.Printf("Created: %s\n", fullPath)
		return nil
	},
}

var fileRmCmd = &cobra.Command{
	Use:   "rm claw:/<path>",
	Short: "Remove file or directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		if err := validateRemotePath(remotePath); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		p, err := parseRemotePath(remotePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		isDir, err := client.IsDirectory(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		if isDir {
			err = client.DeleteFolder(p)
		} else {
			err = client.DeleteFile(p)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		fmt.Printf("Removed: %s\n", p)
		return nil
	},
}

var fileMvCmd = &cobra.Command{
	Use:   "mv claw:/<src> claw:/<dst>",
	Short: "Move or rename file or directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath := args[0]
		dstPath := args[1]

		if err := validateRemotePath(srcPath); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}
		if err := validateRemotePath(dstPath); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		src, err := parseRemotePath(srcPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}
		dst, err := parseRemotePath(dstPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		isDir, err := client.IsDirectory(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		if isDir {
			err = client.RenameFolder(src, dst)
		} else {
			err = client.MoveFile(src, dst)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		fmt.Printf("Moved: %s -> %s\n", src, dst)
		return nil
	},
}

var filePutCmd = &cobra.Command{
	Use:   "put <local> claw:/[remote]",
	Short: "Upload a file",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		localPath := args[0]
		remotePath := "/"

		if len(args) > 1 {
			remotePath = args[1]
			if err := validateRemotePath(remotePath); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			p, err := parseRemotePath(remotePath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				return nil
			}
			remotePath = p
		}

		info, err := os.Stat(localPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}
		if info.IsDir() {
			fmt.Fprintln(os.Stderr, "Error: is a directory")
			return nil
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		localFileName := filepath.Base(localPath)
		fullRemotePath := remotePath
		if remotePath == "/" {
			fullRemotePath = "/" + localFileName
		} else {
			fullRemotePath = remotePath + "/" + localFileName
		}

		fmt.Printf("Uploading %s (%s)...\n", localFileName, formatSize(info.Size()))

		file, err := client.UploadFileByPath(localPath, fullRemotePath, func(pct int) {
			fmt.Printf("\rProgress: %d%%", pct)
		})
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			return nil
		}
		fmt.Printf("\n✓ Uploaded: %s (path: %s)\n", file.OriginalName, file.Path)
		return nil
	},
}

var fileGetCmd = &cobra.Command{
	Use:   "get claw:/<remote> [local]",
	Short: "Download a file",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		localPath := ""

		if len(args) > 1 {
			localPath = args[1]
		}

		if err := validateRemotePath(remotePath); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		p, err := parseRemotePath(remotePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		fmt.Printf("Downloading %s...\n", p)

		path, err := client.DownloadFileByPath(p, localPath, func(pct int) {
			fmt.Printf("\rProgress: %d%%", pct)
		})
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			return nil
		}
		fmt.Printf("\n✓ Saved to: %s\n", path)
		return nil
	},
}

var fileInfoCmd = &cobra.Command{
	Use:   "info claw:/<path>",
	Short: "Show file details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		if err := validateRemotePath(remotePath); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		p, err := parseRemotePath(remotePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		cfg := Config{Endpoint: endpoint, LocalKey: localKey}
		if cfg.Endpoint == "" || cfg.LocalKey == "" {
			loadedCfg, err := loadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				return nil
			}
			cfg = loadedCfg
		}

		client := NewClient(cfg)

		info, err := client.GetFileInfo(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return nil
		}

		printFileInfo(info)
		return nil
	},
}

// ============ Client Methods ============

func (c *Client) List(path string) ([]FolderItem, []FileItem, error) {
	folders, err := c.ListFolders(path)
	if err != nil {
		return nil, nil, err
	}

	files, _, err := c.ListFiles(path)
	return folders, files, err
}

func (c *Client) ListFolders(path string) ([]FolderItem, error) {
	var url string
	if path == "/" || path == "" {
		url = c.Endpoint + "/api/v1/folders"
	} else {
		url = c.Endpoint + "/api/v1/folders/by-path?path=" + path
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list failed: %s", resp.Status)
	}

	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		return nil, errors.New(payload.Message)
	}

	var data FolderListResponse
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		return nil, err
	}
	return data.Folders, nil
}

func (c *Client) ListFiles(path string) ([]FileItem, int, error) {
	url := c.Endpoint + "/api/v1/files/by-path?path=" + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("list failed: %s", resp.Status)
	}

	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, 0, err
	}
	if payload.Code != 0 {
		return nil, 0, errors.New(payload.Message)
	}

	var data FileListResponse
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		return nil, 0, err
	}
	return data.Items, data.Total, nil
}

func (c *Client) CreateFolder(path string) error {
	url := c.Endpoint + "/api/v1/folders/by-path?path=" + path

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create folder failed: %s", resp.Status)
	}

	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.Code != 0 {
		return errors.New(payload.Message)
	}
	return nil
}

func (c *Client) DeleteFile(path string) error {
	url := c.Endpoint + "/api/v1/files/by-path?path=" + path

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) DeleteFolder(path string) error {
	url := c.Endpoint + "/api/v1/folders/by-path?path=" + path

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) MoveFile(srcPath, dstPath string) error {
	url := c.Endpoint + "/api/v1/files/by-path?path=" + srcPath + "&new_path=" + dstPath

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("move failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) RenameFolder(srcPath, dstPath string) error {
	url := c.Endpoint + "/api/v1/folders/by-path?path=" + srcPath + "&new_name=" + filepath.Base(dstPath)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rename failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) IsDirectory(path string) (bool, error) {
	url := c.Endpoint + "/api/v1/folders/by-path?path=" + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	url = c.Endpoint + "/api/v1/files/by-path?path=" + path
	req2, _ := http.NewRequest("GET", url, nil)
	c.attachAuth(req2)
	resp2, err := c.HTTP.Do(req2)
	if err != nil {
		return false, err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusOK {
		return false, nil
	}

	return false, errors.New("path not found")
}

func (c *Client) UploadFileByPath(localPath, remotePath string, progress func(int)) (FileItem, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return FileItem{}, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return FileItem{}, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return FileItem{}, err
	}

	progressReader := NewProgressReader(file, stat.Size(), progress)
	if _, err := io.Copy(part, progressReader); err != nil {
		return FileItem{}, err
	}

	if err := writer.Close(); err != nil {
		return FileItem{}, err
	}

	url := c.Endpoint + "/api/v1/files/by-path?path=" + remotePath

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return FileItem{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return FileItem{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FileItem{}, fmt.Errorf("upload failed: %s", resp.Status)
	}

	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return FileItem{}, err
	}
	if payload.Code != 0 {
		return FileItem{}, errors.New(payload.Message)
	}

	var data FileItem
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		return FileItem{}, err
	}
	return data, nil
}

func (c *Client) DownloadFileByPath(remotePath, localPath string, progress func(int)) (string, error) {
	url := c.Endpoint + "/api/v1/files/by-path/download?path=" + remotePath

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	filename := filepath.Base(remotePath)
	if localPath == "" || localPath == "." {
		localPath = filename
	} else if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		localPath = filepath.Join(localPath, filename)
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return "", err
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		contentLength = 0
	}

	progressReader := NewProgressReader(resp.Body, contentLength, progress)
	if _, err := io.Copy(outFile, progressReader); err != nil {
		return "", err
	}

	return localPath, nil
}

type FileInfo struct {
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	CreatedAt    string `json:"created_at"`
	DownloadLink string `json:"download_link"`
	ExpiresAt    string `json:"expires_at"`
}

func (c *Client) GetFileInfo(path string) (FileInfo, error) {
	url := c.Endpoint + "/api/v1/files/by-path/info?path=" + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return FileInfo{}, err
	}
	c.attachAuth(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return FileInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FileInfo{}, fmt.Errorf("get file info failed: %s", resp.Status)
	}

	var payload APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return FileInfo{}, err
	}
	if payload.Code != 0 {
		return FileInfo{}, errors.New(payload.Message)
	}

	var info FileInfo
	if err := json.Unmarshal(payload.Data, &info); err != nil {
		return FileInfo{}, err
	}
	return info, nil
}

func printFileInfo(info FileInfo) {
	fmt.Printf("File: %s\n", info.OriginalName)
	fmt.Printf("Path: %s\n", info.Path)
	fmt.Printf("Size: %s\n", formatSize(info.Size))
	fmt.Printf("Type: %s\n", info.MimeType)
	fmt.Printf("Created: %s\n", info.CreatedAt)
	if info.DownloadLink != "" {
		fmt.Printf("\nDownload Link (valid for 7 days):\n%s\n", info.DownloadLink)
		fmt.Printf("Expires: %s\n", info.ExpiresAt)
	}
}

// ============ Init ============

func init() {
	fileCmd.AddCommand(fileLsCmd)
	fileCmd.AddCommand(fileMkdirCmd)
	fileCmd.AddCommand(fileRmCmd)
	fileCmd.AddCommand(fileMvCmd)
	fileCmd.AddCommand(filePutCmd)
	fileCmd.AddCommand(fileGetCmd)
	fileCmd.AddCommand(fileInfoCmd)

	fileLsCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	fileLsCmd.Flags().StringVar(&localKey, "key", "", "Local key")

	fileMkdirCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	fileMkdirCmd.Flags().StringVar(&localKey, "key", "", "Local key")

	fileRmCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	fileRmCmd.Flags().StringVar(&localKey, "key", "", "Local key")

	fileMvCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	fileMvCmd.Flags().StringVar(&localKey, "key", "", "Local key")

	filePutCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	filePutCmd.Flags().StringVar(&localKey, "key", "", "Local key")

	fileGetCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	fileGetCmd.Flags().StringVar(&localKey, "key", "", "Local key")

	fileInfoCmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint")
	fileInfoCmd.Flags().StringVar(&localKey, "key", "", "Local key")
}
