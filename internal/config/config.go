package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig    `mapstructure:"server" json:"server"`
	Database DatabaseConfig  `mapstructure:"database" json:"database"`
	Auth     AuthConfig      `mapstructure:"auth" json:"auth"`
	Upload   UploadConfig    `mapstructure:"upload" json:"upload"`
	Minio    MinioConfig     `mapstructure:"minio" json:"minio"`
	Includes []IncludeConfig `mapstructure:"includes" json:"includes"`
	Mail     MailConfig      `mapstructure:"mail" json:"mail"`
	Image    ImageConfig     `mapstructure:"image" json:"image"`
	Logger   LoggerConfig    `mapstructure:"logger" json:"logger"`
}

type ServerConfig struct {
	Port           int    `mapstructure:"port" json:"port"`
	LogLevel       string `mapstructure:"log_level" json:"log_level"`
	PublicEndpoint string `mapstructure:"public_endpoint" json:"public_endpoint"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path" json:"path"`
}

type AuthConfig struct {
	JWTSecret         string `mapstructure:"jwt_secret" json:"jwt_secret"`
	LocalKey          string `mapstructure:"local_key" json:"local_key"`
	JWTExpireHours    int64  `mapstructure:"jwt_expire_hours" json:"jwt_expire_hours"`
	RefreshExpireDays int64  `mapstructure:"refresh_expire_days" json:"refresh_expire_days"`
	AdminUsername     string `mapstructure:"admin_username" json:"admin_username"`
	AdminPassword     string `mapstructure:"admin_password" json:"admin_password"`
}

type UploadConfig struct {
	MaxSizeMB int64 `mapstructure:"max_size_mb" json:"max_size_mb"`
}

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint" json:"endpoint"`
	AccessKey string `mapstructure:"access_key" json:"access_key"`
	SecretKey string `mapstructure:"secret_key" json:"secret_key"`
	Bucket    string `mapstructure:"bucket" json:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl" json:"use_ssl"`
	Region    string `mapstructure:"region" json:"region"`
}

type IncludeConfig struct {
	Name string `mapstructure:"name" json:"name"`
	Path string `mapstructure:"path" json:"path"`
}

type MailConfig struct {
	Accounts   []AccountConfig  `mapstructure:"accounts" json:"accounts"`
	Webhook    WebhookConfig    `mapstructure:"webhook" json:"webhook"`
	Monitoring MonitoringConfig `mapstructure:"monitoring" json:"monitoring"`
}

type AccountConfig struct {
	Provider  string `mapstructure:"provider" json:"provider"`
	Email     string `mapstructure:"email" json:"email"`
	AuthToken string `mapstructure:"auth_token" json:"auth_token"`
	Enabled   bool   `mapstructure:"enabled" json:"enabled"`
}

type WebhookConfig struct {
	URL           string `mapstructure:"url" json:"url"`
	Token         string `mapstructure:"token" json:"token"`
	To            string `mapstructure:"to" json:"to"`
	SessionKey    string `mapstructure:"session_key" json:"session_key"`
	CustomPayload string `mapstructure:"custom_payload" json:"custom_payload"`
	Enable        bool   `mapstructure:"enable" json:"enable"`
}

type MonitoringConfig struct {
	PollInterval string `mapstructure:"poll_interval" json:"poll_interval"`
}

type ImageConfig struct {
	Libvips         LibvipsConfig         `mapstructure:"libvips" json:"libvips"`
	OCR             OCRConfig             `mapstructure:"ocr" json:"ocr"`
	Vision          VisionConfig          `mapstructure:"vision" json:"vision"`
	ImageGeneration ImageGenerationConfig `mapstructure:"image_generation" json:"image_generation"`
}

type LibvipsConfig struct {
	Path string `mapstructure:"path" json:"path"`
}

type OCRConfig struct {
	APIKey string `mapstructure:"api_key" json:"api_key"`
}

type VisionConfig struct {
	APIKey string `mapstructure:"api_key" json:"api_key"`
}

type ImageGenerationConfig struct {
	APIKey string `mapstructure:"api_key" json:"api_key"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	Format     string `mapstructure:"format" json:"format"`
	OutputPath string `mapstructure:"output_path" json:"output_path"`
	EnableFile bool   `mapstructure:"enable_file" json:"enable_file"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:     8080,
			LogLevel: "info",
		},
		Database: DatabaseConfig{
			Path: "./data/claw-pliers.db",
		},
		Auth: AuthConfig{
			JWTSecret:         "",
			LocalKey:          "",
			AdminUsername:     "admin",
			JWTExpireHours:    24,
			RefreshExpireDays: 7,
		},
		Upload: UploadConfig{
			MaxSizeMB: 1024,
		},
		Minio: MinioConfig{
			Bucket: "claw-pliers",
			UseSSL: false,
		},
		Mail: MailConfig{
			Monitoring: MonitoringConfig{
				PollInterval: "30s",
			},
		},
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "console",
			OutputPath: "./logs",
			EnableFile: true,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return Config{}, fmt.Errorf("invalid config path: %w", err)
	}

	dir := filepath.Dir(absPath)
	filename := filepath.Base(absPath)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/claw-pliers")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := loadIncludes(&cfg, dir); err != nil {
		return Config{}, err
	}

	overrideWithEnv(&cfg)

	if cfg.Auth.LocalKey == "" {
		return Config{}, fmt.Errorf("missing auth.local_key in config")
	}

	return cfg, nil
}

func loadIncludes(cfg *Config, baseDir string) error {
	for _, inc := range cfg.Includes {
		if inc.Path == "" || inc.Name == "" {
			continue
		}

		incAbsPath := inc.Path
		if !filepath.IsAbs(inc.Path) {
			incAbsPath = filepath.Join(baseDir, inc.Path)
		}

		v := viper.New()
		v.SetConfigFile(incAbsPath)
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			continue
		}

		switch inc.Name {
		case "file":
			if v.IsSet("database.path") {
				cfg.Database.Path = v.GetString("database.path")
			}
			if v.IsSet("upload.max_size_mb") {
				cfg.Upload.MaxSizeMB = v.GetInt64("upload.max_size_mb")
			}
			if v.IsSet("minio.endpoint") {
				cfg.Minio.Endpoint = v.GetString("minio.endpoint")
			}
			if v.IsSet("minio.access_key") {
				cfg.Minio.AccessKey = v.GetString("minio.access_key")
			}
			if v.IsSet("minio.secret_key") {
				cfg.Minio.SecretKey = v.GetString("minio.secret_key")
			}
			if v.IsSet("minio.bucket") {
				cfg.Minio.Bucket = v.GetString("minio.bucket")
			}
			if v.IsSet("minio.use_ssl") {
				cfg.Minio.UseSSL = v.GetBool("minio.use_ssl")
			}
			if v.IsSet("minio.region") {
				cfg.Minio.Region = v.GetString("minio.region")
			}
			if v.IsSet("auth.local_key") && cfg.Auth.LocalKey == "" {
				cfg.Auth.LocalKey = v.GetString("auth.local_key")
			}
			if v.IsSet("auth.jwt_secret") {
				cfg.Auth.JWTSecret = v.GetString("auth.jwt_secret")
			}
			if v.IsSet("auth.admin_username") {
				cfg.Auth.AdminUsername = v.GetString("auth.admin_username")
			}
			if v.IsSet("auth.admin_password") {
				cfg.Auth.AdminPassword = v.GetString("auth.admin_password")
			}

		case "mail":
			if v.IsSet("accounts") {
				accounts := v.Get("accounts").([]interface{})
				cfg.Mail.Accounts = parseAccounts(accounts)
			}
			if v.IsSet("webhook.url") {
				cfg.Mail.Webhook.URL = v.GetString("webhook.url")
			}
			if v.IsSet("webhook.token") {
				cfg.Mail.Webhook.Token = v.GetString("webhook.token")
			}
			if v.IsSet("webhook.to") {
				cfg.Mail.Webhook.To = v.GetString("webhook.to")
			}
			if v.IsSet("webhook.session_key") {
				cfg.Mail.Webhook.SessionKey = v.GetString("webhook.session_key")
			}
			if v.IsSet("webhook.custom_payload") {
				cfg.Mail.Webhook.CustomPayload = v.GetString("webhook.custom_payload")
			}
			if v.IsSet("webhook.enable") {
				cfg.Mail.Webhook.Enable = v.GetBool("webhook.enable")
			}
			if v.IsSet("monitoring.poll_interval") {
				cfg.Mail.Monitoring.PollInterval = v.GetString("monitoring.poll_interval")
			}

		case "image":
			if v.IsSet("libvips.path") {
				cfg.Image.Libvips.Path = v.GetString("libvips.path")
			}
			if v.IsSet("ocr.api_key") {
				cfg.Image.OCR.APIKey = v.GetString("ocr.api_key")
			}
			if v.IsSet("vision.api_key") {
				cfg.Image.Vision.APIKey = v.GetString("vision.api_key")
			}
			if v.IsSet("image_generation.api_key") {
				cfg.Image.ImageGeneration.APIKey = v.GetString("image_generation.api_key")
			}
		}
	}

	return nil
}

func parseAccounts(list []interface{}) []AccountConfig {
	accounts := make([]AccountConfig, 0, len(list))
	for _, item := range list {
		if acc, ok := item.(map[string]interface{}); ok {
			account := AccountConfig{
				Enabled: true,
			}
			if p, ok := acc["provider"].(string); ok {
				account.Provider = p
			}
			if e, ok := acc["email"].(string); ok {
				account.Email = e
			}
			if t, ok := acc["auth_token"].(string); ok {
				account.AuthToken = t
			}
			if en, ok := acc["enabled"].(bool); ok {
				account.Enabled = en
			}
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func overrideWithEnv(cfg *Config) {
	if value := os.Getenv("CLAWPLIERS_SERVER_PORT"); value != "" {
		cfg.Server.Port = parseIntValue(value, cfg.Server.Port)
	}
	if value := os.Getenv("CLAWPLIERS_SERVER_LOG_LEVEL"); value != "" {
		cfg.Server.LogLevel = value
	}
	if value := os.Getenv("CLAWPLIERS_SERVER_PUBLIC_ENDPOINT"); value != "" {
		cfg.Server.PublicEndpoint = value
	}
	if value := os.Getenv("CLAWPLIERS_DATABASE_PATH"); value != "" {
		cfg.Database.Path = value
	}
	if value := os.Getenv("CLAWPLIERS_AUTH_LOCAL_KEY"); value != "" {
		cfg.Auth.LocalKey = value
	}
	if value := os.Getenv("CLAWPLIERS_AUTH_JWT_SECRET"); value != "" {
		cfg.Auth.JWTSecret = value
	}
	if value := os.Getenv("CLAWPLIERS_AUTH_ADMIN_USERNAME"); value != "" {
		cfg.Auth.AdminUsername = value
	}
	if value := os.Getenv("CLAWPLIERS_AUTH_ADMIN_PASSWORD"); value != "" {
		cfg.Auth.AdminPassword = value
	}
	if value := os.Getenv("CLAWPLIERS_MINIO_ENDPOINT"); value != "" {
		cfg.Minio.Endpoint = value
	}
	if value := os.Getenv("CLAWPLIERS_MINIO_ACCESS_KEY"); value != "" {
		cfg.Minio.AccessKey = value
	}
	if value := os.Getenv("CLAWPLIERS_MINIO_SECRET_KEY"); value != "" {
		cfg.Minio.SecretKey = value
	}
	if value := os.Getenv("CLAWPLIERS_MINIO_BUCKET"); value != "" {
		cfg.Minio.Bucket = value
	}
	if value := os.Getenv("CLAWPLIERS_MINIO_USE_SSL"); value != "" {
		cfg.Minio.UseSSL = parseBoolValue(value, cfg.Minio.UseSSL)
	}
	if value := os.Getenv("CLAWPLIERS_MAIL_WEBHOOK_URL"); value != "" {
		cfg.Mail.Webhook.URL = value
	}
	if value := os.Getenv("CLAWPLIERS_MAIL_WEBHOOK_TOKEN"); value != "" {
		cfg.Mail.Webhook.Token = value
	}
	if value := os.Getenv("CLAWPLIERS_OCR_API_KEY"); value != "" {
		cfg.Image.OCR.APIKey = value
	}
	if value := os.Getenv("CLAWPLIERS_VISION_API_KEY"); value != "" {
		cfg.Image.Vision.APIKey = value
	}
	if value := os.Getenv("CLAWPLIERS_IMAGE_GENERATION_API_KEY"); value != "" {
		cfg.Image.ImageGeneration.APIKey = value
	}
	if value := os.Getenv("CLAWPLIERS_LOG_LEVEL"); value != "" {
		cfg.Logger.Level = value
	}
}

func parseIntValue(value string, fallback int) int {
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return fallback
	}
	return n
}

func parseBoolValue(value string, fallback bool) bool {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	}
	return fallback
}
