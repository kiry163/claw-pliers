package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kiry163/claw-pliers/internal/api"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/file"
	"github.com/kiry163/claw-pliers/internal/image"
	"github.com/kiry163/claw-pliers/internal/logger"
	"github.com/kiry163/claw-pliers/internal/mail"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	configPath := flag.String("config", "config.yaml", "Config file path")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	logCfg := logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		OutputPath: cfg.Logger.OutputPath,
		EnableFile: cfg.Logger.EnableFile,
	}
	if err := logger.Init(logCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	log := logger.Get()
	log.Info().Str("version", version).Str("config", *configPath).Msg("starting claw-pliers")

	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o755); err != nil {
		log.Fatal().Err(err).Msg("failed to create database directory")
	}

	if err := initModules(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize modules")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("shutting down...")
		cancel()
	}()

	router := api.NewRouter(&cfg, file.Database, version)

	address := ":" + fmt.Sprintf("%d", cfg.Server.Port)
	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Info().Str("address", address).Msg("server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("server stopped")
}

func initModules(cfg config.Config) error {
	log := logger.Get()

	log.Info().Msg("initializing file module")
	if err := file.Init(cfg); err != nil {
		return fmt.Errorf("file module init failed: %w", err)
	}
	log.Info().Msg("file module initialized")

	log.Info().Msg("initializing mail module")
	if err := mail.Init(cfg); err != nil {
		return fmt.Errorf("mail module init failed: %w", err)
	}
	log.Info().Msg("mail module initialized")

	log.Info().Msg("initializing image module")
	if err := image.Init(cfg); err != nil {
		return fmt.Errorf("image module init failed: %w", err)
	}
	log.Info().Msg("image module initialized")

	return nil
}
