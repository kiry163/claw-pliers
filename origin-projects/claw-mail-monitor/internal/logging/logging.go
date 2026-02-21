package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiry163/claw-mail-monitor/internal/config"
)

func Init(cfg config.Logging) {
	initWithWriter(cfg, os.Stdout)
}

func InitStderr(cfg config.Logging) {
	initWithWriter(cfg, os.Stderr)
}

func initWithWriter(cfg config.Logging, output io.Writer) {
	level := parseLevel(cfg.Level)
	handler := slog.NewTextHandler(output, &slog.HandlerOptions{Level: level})

	if cfg.File != "" {
		path := expandHome(cfg.File)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
			file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err == nil {
				handler = slog.NewTextHandler(io.MultiWriter(output, file), &slog.HandlerOptions{Level: level})
			}
		}
	}

	slog.SetDefault(slog.New(handler))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			rest := strings.TrimPrefix(path, "~")
			rest = strings.TrimPrefix(rest, string(filepath.Separator))
			return filepath.Join(home, rest)
		}
	}
	return path
}
