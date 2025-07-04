//go:build windows
// +build windows

package main

import (
	"math/rand"
	"os"
	"time"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/trayapp"
	"github.com/caarlos0/log"
)

// TrayAppConfig holds the configuration for the tray application
type TrayAppConfig struct {
	LoadConfig      func() error
	GetModifiedTime func() time.Time
	RunCommand      func(command string, args ...string) (string, error)
	SelfExe         func() string
	SystrayRun      func(onReady func(), onExit func())
	Exit            func(code int)
	Sleep           func(duration time.Duration)
	Rand            func(n int) int
}

// DefaultTrayAppConfig returns the default configuration
func DefaultTrayAppConfig() *TrayAppConfig {
	return &TrayAppConfig{
		LoadConfig:      shared.LoadConfig,
		GetModifiedTime: shared.GetModifiedTime,
		RunCommand:      shared.RunCommand,
		SelfExe:         shared.SelfExe,
		SystrayRun:      systray.Run,
		Exit:            os.Exit,
		Sleep:           time.Sleep,
		Rand:            rand.Intn,
	}
}

// TrayApp represents the tray application
type TrayApp struct {
	config *TrayAppConfig
}

// NewTrayApp creates a new tray application instance
func NewTrayApp(config *TrayAppConfig) *TrayApp {
	if config == nil {
		config = DefaultTrayAppConfig()
	}
	return &TrayApp{config: config}
}

// Run executes the tray application logic
func (t *TrayApp) Run() {
	if err := t.config.LoadConfig(); err != nil {
		log.WithError(err).Warn("failed to load config")
	}

	// Scheduled update command
	go t.updateScheduler()

	// Scheduled check command
	go t.checkScheduler()

	// Initialize the state file
	if t.config.GetModifiedTime().IsZero() {
		log.Info("Initializing state file...")
		go t.runInitialCheck()
	}

	onExit := func() {
		log.Info("Exiting...")
		t.config.Exit(0)
	}

	log.Info("Starting system tray application...")
	t.config.SystrayRun(trayapp.OnReady, onExit)
}

func (t *TrayApp) updateScheduler() {
	log.Info("Starting update command scheduler...")
	for {
		if time.Since(t.config.GetModifiedTime()) > time.Hour && !t.config.GetModifiedTime().IsZero() {
			_, err := t.config.RunCommand(t.config.SelfExe(), "update")
			if err != nil {
				log.WithError(err).Error("Failed to run update command")
			}
		}
		t.config.Sleep(time.Duration(50+t.config.Rand(15)) * time.Minute)
	}
}

func (t *TrayApp) checkScheduler() {
	log.Info("Starting check command scheduler...")
	for {
		t.config.Sleep(time.Duration(45+t.config.Rand(15)) * time.Minute)
		_, err := t.config.RunCommand(t.config.SelfExe(), "check")
		if err != nil {
			log.WithError(err).Error("Failed to run check command")
		}
	}
}

func (t *TrayApp) runInitialCheck() {
	_, err := t.config.RunCommand(t.config.SelfExe(), "check")
	if err != nil {
		log.WithError(err).Error("Failed to run check command")
	}
}
