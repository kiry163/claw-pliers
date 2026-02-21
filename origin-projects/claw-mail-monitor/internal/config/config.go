package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	envConfigPath = "CLAW_MAIL_MONITOR_CONFIG"
	defaultConfig = "config.yaml"
)

type Config struct {
	Accounts   []Account  `yaml:"accounts"`
	Webhook    Webhook    `yaml:"webhook"`
	Monitoring Monitoring `yaml:"monitoring"`
	Logging    Logging    `yaml:"logging"`
	ConfigPath string     `yaml:"-"`
}

type Account struct {
	Provider  string `yaml:"provider"`
	Email     string `yaml:"email"`
	AuthToken string `yaml:"auth_token"`
	Enabled   bool   `yaml:"enabled"`
	IMAPHost  string `yaml:"imap_host"`
	SMTPHost  string `yaml:"smtp_host"`
}

type Webhook struct {
	URL           string `yaml:"url"`
	Token         string `yaml:"token"`
	To            string `yaml:"to"`
	SessionKey    string `yaml:"session_key"`
	CustomPayload string `yaml:"custom_payload"`
	Enable        bool   `yaml:"enable"`
}

type Logging struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type Monitoring struct {
	PollInterval string `yaml:"poll_interval"`
}

func Load() (*Config, error) {
	path, err := resolveConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	cfg.ConfigPath = path
	return &cfg, nil
}

func DefaultConfigPath() (string, error) {
	return resolveConfigPath()
}

func ExpandPath(path string) string {
	if path == "" {
		return path
	}
	if path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	trimmed := strings.TrimPrefix(path, "~")
	trimmed = strings.TrimPrefix(trimmed, string(filepath.Separator))
	return filepath.Join(home, trimmed)
}

func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("nil config")
	}

	path := cfg.ConfigPath
	if path == "" {
		var err error
		path, err = resolveConfigPath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir failed: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config failed: %w", err)
	}

	cfg.ConfigPath = path
	return nil
}

func resolveConfigPath() (string, error) {
	if v := os.Getenv(envConfigPath); v != "" {
		return v, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir failed: %w", err)
	}

	return filepath.Join(home, ".config", "claw-mail-monitor", defaultConfig), nil
}

func CachePath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir failed: %w", err)
	}

	return filepath.Join(home, ".cache", "claw-mail-monitor", name), nil
}

func (c *Config) GetProviderHosts(provider string) (string, string) {
	hosts := map[string]struct {
		imap string
		smtp string
	}{
		"qq":    {"imap.qq.com:993", "smtp.qq.com:465"},
		"163":   {"imap.163.com:993", "smtp.163.com:465"},
		"gmail": {"imap.gmail.com:993", "smtp.gmail.com:465"},
	}

	if v, ok := hosts[provider]; ok {
		return v.imap, v.smtp
	}
	return "", ""
}

func (c *Config) ApplyDefaults(account *Account) {
	if account == nil {
		return
	}

	if account.IMAPHost == "" || account.SMTPHost == "" {
		imapHost, smtpHost := c.GetProviderHosts(account.Provider)
		if account.IMAPHost == "" {
			account.IMAPHost = imapHost
		}
		if account.SMTPHost == "" {
			account.SMTPHost = smtpHost
		}
	}
}
