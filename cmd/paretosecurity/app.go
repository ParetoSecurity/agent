package main

import (
	"github.com/ParetoSecurity/agent/cmd"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

// AppConfig holds the configuration for the application
type AppConfig struct {
	LoadConfig func() error
	IsRoot     func() bool
	Execute    func()
}

// DefaultAppConfig returns the default configuration
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		LoadConfig: shared.LoadConfig,
		IsRoot:     shared.IsRoot,
		Execute:    cmd.Execute,
	}
}

// App represents the main application
type App struct {
	config *AppConfig
}

// NewApp creates a new application instance
func NewApp(config *AppConfig) *App {
	if config == nil {
		config = DefaultAppConfig()
	}
	return &App{config: config}
}

// Run executes the application logic
func (a *App) Run() error {
	if err := a.config.LoadConfig(); err != nil {
		if !a.config.IsRoot() {
			log.WithError(err).Warn("failed to load config")
		}
	}
	a.config.Execute()
	return nil
}
