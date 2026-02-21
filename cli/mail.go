package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

type MailConfig struct {
	Accounts []MailAccount `yaml:"accounts"`
}

type MailAccount struct {
	Provider  string `yaml:"provider"`
	Email     string `yaml:"email"`
	ImapHost  string `yaml:"imap_host"`
	ImapPort  int    `yaml:"imap_port"`
	SmtpHost  string `yaml:"smtp_host"`
	SmtpPort  int    `yaml:"smtp_port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"-"`
	AuthToken string `yaml:"auth_token"`
	Enabled   bool   `yaml:"enabled"`
}

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Mail management commands",
}

var mailSendCmd = &cobra.Command{
	Use:   "send --from <email> --to <email> --subject <subject> --body <body>",
	Short: "Send an email",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")

		if from == "" || to == "" || subject == "" || body == "" {
			fmt.Println("Error: --from, --to, --subject and --body are required")
			return nil
		}

		config, err := loadMailConfig()
		if err != nil || len(config.Accounts) == 0 {
			fmt.Println("Error: no accounts configured")
			return nil
		}

		var account *MailAccount
		for i := range config.Accounts {
			if config.Accounts[i].Email == from {
				account = &config.Accounts[i]
				break
			}
		}

		if account == nil {
			fmt.Printf("Error: account %s not found in local config\n", from)
			fmt.Println("Note: Account must be configured in server config")
			return nil
		}

		_, err = callMailAPIWithResponse("send", map[string]string{
			"from":    from,
			"to":      to,
			"subject": subject,
			"body":    body,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil
		}

		fmt.Println("✓ Email sent successfully!")
		return nil
	},
}

var mailListCmd = &cobra.Command{
	Use:   "list",
	Short: "List mail accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := callMailAPIWithResponse("accounts", nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil
		}

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return nil
		}

		accounts, ok := response["accounts"].([]interface{})
		if !ok || len(accounts) == 0 {
			fmt.Println("No accounts configured on server")
			return nil
		}

		fmt.Println("Configured accounts:")
		for _, acc := range accounts {
			accMap, ok := acc.(map[string]interface{})
			if !ok {
				continue
			}
			email, _ := accMap["email"].(string)
			provider, _ := accMap["provider"].(string)
			fmt.Printf("  - %s (%s)\n", email, provider)
		}
		return nil
	},
}

var mailAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "Mail account management (local config)",
}

var mailAccountAddCmd = &cobra.Command{
	Use:   "add --provider <provider> --email <email> --username <user> --password <pass> [--auth-token <token>]",
	Short: "Add a mail account to local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		email, _ := cmd.Flags().GetString("email")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		authToken, _ := cmd.Flags().GetString("auth-token")

		if provider == "" || email == "" || username == "" || password == "" {
			fmt.Println("Error: --provider, --email, --username and --password are required")
			return nil
		}

		imapHost, smtpHost := getProviderSettings(provider)
		if imapHost == "" {
			fmt.Printf("Error: unknown provider %s\n", provider)
			return nil
		}

		account := MailAccount{
			Provider:  provider,
			Email:     email,
			Username:  username,
			Password:  password,
			ImapHost:  imapHost,
			ImapPort:  993,
			SmtpHost:  smtpHost,
			SmtpPort:  465,
			AuthToken: authToken,
			Enabled:   true,
		}

		config, err := loadMailConfig()
		if err != nil {
			config = &MailConfig{}
		}

		for _, acc := range config.Accounts {
			if acc.Email == email {
				fmt.Printf("Error: account %s already exists\n", email)
				return nil
			}
		}

		config.Accounts = append(config.Accounts, account)

		if err := saveMailConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return nil
		}

		fmt.Printf("Added account: %s (%s)\n", email, provider)
		fmt.Println("Note: Ensure the account is also configured in server config")
		return nil
	},
}

var mailAccountRemoveCmd = &cobra.Command{
	Use:   "remove --email <email>",
	Short: "Remove a mail account",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			fmt.Println("Error: --email is required")
			return nil
		}

		config, err := loadMailConfig()
		if err != nil {
			fmt.Println("No accounts configured")
			return nil
		}

		found := false
		newAccounts := []MailAccount{}
		for _, acc := range config.Accounts {
			if acc.Email == email {
				found = true
				continue
			}
			newAccounts = append(newAccounts, acc)
		}

		if !found {
			fmt.Printf("Error: account %s not found\n", email)
			return nil
		}

		config.Accounts = newAccounts
		if err := saveMailConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return nil
		}

		fmt.Printf("Removed account: %s\n", email)
		return nil
	},
}

func getMailConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "claw-pliers")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "mail.yaml")
}

func loadMailConfig() (*MailConfig, error) {
	path := getMailConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config MailConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func saveMailConfig(config *MailConfig) error {
	path := getMailConfigPath()
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func getProviderSettings(provider string) (imapHost, smtpHost string) {
	switch provider {
	case "163":
		return "imap.163.com", "smtp.163.com"
	case "126":
		return "imap.126.com", "smtp.126.com"
	case "qq":
		return "imap.qq.com", "smtp.qq.com"
	case "gmail":
		return "imap.gmail.com", "smtp.gmail.com"
	case "outlook":
		return "outlook.office365.com", "smtp.office365.com"
	default:
		return "", ""
	}
}

var mailTestConnectionCmd = &cobra.Command{
	Use:   "test-connection --email <email>",
	Short: "Test mail account connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			fmt.Println("Error: --email is required")
			return nil
		}

		fmt.Printf("Testing IMAP connection to %s...\n", email)

		result, err := callMailAPIWithResponse("test-connection?email="+email, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil
		}

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return nil
		}

		latency, _ := response["latency"].(float64)
		status, _ := response["status"].(string)

		if status == "ok" {
			fmt.Printf("✓ Connection successful! Latency: %.0f ms\n", latency)
		} else {
			msg, _ := response["message"].(string)
			fmt.Printf("Error: %s\n", msg)
		}

		return nil
	},
}

var mailLatestCmd = &cobra.Command{
	Use:   "latest --count <n> --email <email>",
	Short: "Get latest emails",
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := cmd.Flags().GetInt("count")
		email, _ := cmd.Flags().GetString("email")

		if count <= 0 {
			count = 5
		}

		if email == "" {
			config, err := loadMailConfig()
			if err != nil || len(config.Accounts) == 0 {
				fmt.Println("Error: no accounts configured")
				return nil
			}
			email = config.Accounts[0].Email
		}

		fmt.Printf("Fetching latest %d emails from %s...\n", count, email)

		result, err := callMailAPIWithResponse(fmt.Sprintf("latest?email=%s&count=%d", email, count), nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil
		}

		var response map[string]interface{}
		if err := json.Unmarshal([]byte(result), &response); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return nil
		}

		emails, ok := response["emails"].([]interface{})
		if !ok || len(emails) == 0 {
			fmt.Println("No emails found")
			return nil
		}

		fmt.Printf("\n=== Latest %d emails ===\n\n", len(emails))
		for i, e := range emails {
			emailMap, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			from, _ := emailMap["from"].(string)
			subject, _ := emailMap["subject"].(string)
			date, _ := emailMap["date"].(string)
			preview, _ := emailMap["preview"].(string)

			fmt.Printf("[%d] From: %s\n", i+1, from)
			fmt.Printf("    Subject: %s\n", subject)
			fmt.Printf("    Date: %s\n", date)
			fmt.Printf("    Preview: %s\n", preview)
			fmt.Println("---")
		}

		return nil
	},
}

var mailMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Mail monitoring commands",
}

var mailMonitorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show monitor status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadMailConfig()
		if err != nil || len(config.Accounts) == 0 {
			fmt.Println("No accounts configured")
			return nil
		}

		fmt.Printf("Monitor Status: Running\n")
		fmt.Printf("Watching %d account(s):\n", len(config.Accounts))
		for _, acc := range config.Accounts {
			fmt.Printf("  - %s (%s)\n", acc.Email, acc.Provider)
		}

		return nil
	},
}

var mailMonitorStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start mail monitoring",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Starting mail monitor...")
		fmt.Println("Note: Background monitoring is handled by the server")
		return nil
	},
}

func init() {
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailListCmd)
	mailCmd.AddCommand(mailAccountCmd)
	mailAccountCmd.AddCommand(mailAccountAddCmd)
	mailAccountCmd.AddCommand(mailAccountRemoveCmd)
	mailCmd.AddCommand(mailTestConnectionCmd)
	mailCmd.AddCommand(mailLatestCmd)
	mailCmd.AddCommand(mailMonitorCmd)
	mailMonitorCmd.AddCommand(mailMonitorStatusCmd)
	mailMonitorCmd.AddCommand(mailMonitorStartCmd)

	mailAccountAddCmd.Flags().String("provider", "", "Email provider (163, 126, qq, gmail, outlook)")
	mailAccountAddCmd.Flags().String("email", "", "Email address")
	mailAccountAddCmd.Flags().String("username", "", "Username (usually email)")
	mailAccountAddCmd.Flags().String("password", "", "Password or app password")
	mailAccountAddCmd.Flags().String("auth-token", "", "Auth token (optional)")

	mailAccountRemoveCmd.Flags().String("email", "", "Email address to remove")
	mailTestConnectionCmd.Flags().String("email", "", "Email address to test")
	mailLatestCmd.Flags().Int("count", 5, "Number of emails to fetch")
	mailLatestCmd.Flags().String("email", "", "Email account (optional)")
	mailSendCmd.Flags().String("from", "", "From email address")
	mailSendCmd.Flags().String("to", "", "To email address")
	mailSendCmd.Flags().String("subject", "", "Email subject")
	mailSendCmd.Flags().String("body", "", "Email body")
}

func callMailAPIWithResponse(endpoint string, params map[string]string) (string, error) {
	// Load server config (endpoint and localKey) from file.go's loadConfig
	serverCfg, err := loadConfig()
	if err != nil {
		// If config loading fails, use defaults
		serverCfg = Config{Endpoint: "http://localhost:8080", LocalKey: ""}
	}

	// Override with environment variables if set
	if envEndpoint := os.Getenv("CLAWPLIERS_ENDPOINT"); envEndpoint != "" {
		serverCfg.Endpoint = envEndpoint
	}
	if envKey := os.Getenv("CLAWPLIERS_AUTH_LOCAL_KEY"); envKey != "" {
		serverCfg.LocalKey = envKey
	}

	// Load mail account config (for potential future use)
	_, _ = loadMailConfig()

	apiURL := fmt.Sprintf("%s/api/v1/mail/%s", serverCfg.Endpoint, endpoint)

	var req *http.Request
	if params != nil && len(params) > 0 {
		form := url.Values{}
		for k, v := range params {
			form.Add(k, v)
		}
		req, err = http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return "", err
		}
	}

	req.Header.Set("X-Local-Key", serverCfg.LocalKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	return string(body), nil
}

func mergeMailConfig(cfg *MailConfig, configPath string) *MailConfig {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return cfg
	}

	if server, ok := raw["server"].(map[string]interface{}); ok {
		if port, ok := server["port"].(int); ok {
			os.Setenv("CLAWPLIERS_ENDPOINT", fmt.Sprintf("http://localhost:%d", port))
		} else if portFloat, ok := server["port"].(float64); ok {
			os.Setenv("CLAWPLIERS_ENDPOINT", fmt.Sprintf("http://localhost:%d", int(portFloat)))
		}
	}

	if auth, ok := raw["auth"].(map[string]interface{}); ok {
		if lk, ok := auth["local_key"].(string); ok {
			os.Setenv("CLAWPLIERS_AUTH_LOCAL_KEY", lk)
		}
	}

	return cfg
}
