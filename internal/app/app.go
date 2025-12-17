// Package app provides application initialization and configuration loading.
package app

import (
	"fmt"

	"github.com/neox5/tct/internal/config"
	"github.com/neox5/tct/internal/env"
	"github.com/neox5/tct/internal/logger"
)

// App holds the initialized application state.
type App struct {
	Mode   string
	Config *config.Config
	Logger *logger.Logger
}

// New initializes the application by loading configuration and setting up logging.
// It validates the mode and returns an error if initialization fails.
func New() (*App, error) {
	// Load configuration from environment
	cfg := &config.Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Validate mode
	if cfg.Mode != "sender" && cfg.Mode != "receiver" {
		return nil, fmt.Errorf("invalid mode %q (must be 'sender' or 'receiver')", cfg.Mode)
	}

	// Initialize logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return &App{
		Mode:   cfg.Mode,
		Config: cfg,
		Logger: log,
	}, nil
}
