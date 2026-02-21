package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	Level      string `mapstructure:"level" json:"level"`
	Format     string `mapstructure:"format" json:"format"`
	OutputPath string `mapstructure:"output_path" json:"output_path"`
	EnableFile bool   `mapstructure:"enable_file" json:"enable_file"`
}

func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "console",
		OutputPath: "./logs",
		EnableFile: true,
	}
}

var log zerolog.Logger

func Init(cfg Config) error {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}

	var writers []io.Writer

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}
	writers = append(writers, consoleWriter)

	if cfg.EnableFile {
		logDir := cfg.OutputPath
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile := filepath.Join(logDir, "app.log")
		fileWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		writers = append(writers, fileWriter)
	}

	multiWriter := io.MultiWriter(writers...)

	log = zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Info().
		Str("level", level.String()).
		Str("format", cfg.Format).
		Bool("file", cfg.EnableFile).
		Msg("logger initialized")

	return nil
}

func Get() *zerolog.Logger {
	return &log
}

func Trace() *zerolog.Event {
	return log.Trace()
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Info() *zerolog.Event {
	return log.Info()
}

func Warn() *zerolog.Event {
	return log.Warn()
}

func Error() *zerolog.Event {
	return log.Error()
}

func Fatal() *zerolog.Event {
	return log.Fatal()
}

func Panic() *zerolog.Event {
	return log.Panic()
}

func With() zerolog.Context {
	return log.With()
}
