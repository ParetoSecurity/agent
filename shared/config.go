package shared

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/log"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
)

var Config ParetoConfig
var ConfigPath string

type ParetoConfig struct {
	TeamID        string
	AuthToken     string
	SystemUUID    string
	DisableChecks []string
}

// init initializes the configuration path based on the user's operating system
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Warn("failed to get user home directory, using current directory instead")
		homeDir = "."
	}
	ConfigPath = filepath.Join(homeDir, ".config", "pareto.toml")
	if runtime.GOOS == "windows" {
		ConfigPath = filepath.Join(homeDir, "pareto.toml")
	}
	log.Debugf("configPath: %s", ConfigPath)
}

// SaveConfig writes the current configuration to the config file
func SaveConfig() error {

	if err := os.MkdirAll(filepath.Dir(ConfigPath), 0755); err != nil {
		log.WithError(err).Error("failed to create config directory")
	}

	file, err := os.Create(ConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := toml.NewEncoder(file)
	return encoder.Encode(Config)
}

func LoadConfig() error {

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		if err := SaveConfig(); err != nil {
			return err
		}
		return nil
	}
	file, err := os.Open(ConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		return err
	}

	return nil
}

// ResetConfig clears all configuration values to defaults
func ResetConfig() {
	Config = ParetoConfig{
		TeamID:        "",
		AuthToken:     "",
		SystemUUID:    "",
		DisableChecks: []string{},
	}
	SaveConfig()
}

// EnableCheck removes a check from the disabled checks list
func EnableCheck(checkUUID string) error {
	for i, check := range Config.DisableChecks {
		if check == checkUUID {
			Config.DisableChecks = append(Config.DisableChecks[:i], Config.DisableChecks[i+1:]...)
			return SaveConfig()
		}
	}
	return nil
}

// DisableCheck adds a check to the disabled checks list
func DisableCheck(checkUUID string) error {
	for _, check := range Config.DisableChecks {
		if check == checkUUID {
			return nil
		}
	}
	Config.DisableChecks = append(Config.DisableChecks, checkUUID)
	return SaveConfig()
}

// IsCheckDisabled checks if a given check UUID is present in the list of disabled checks
func IsCheckDisabled(checkUUID string) bool {
	if len(Config.DisableChecks) == 0 {
		return false
	}
	for _, check := range Config.DisableChecks {
		if check == checkUUID {
			return true
		}
	}
	return false
}

// GetDeviceUUID returns the system UUID from the configuration
func GetDeviceUUID() string {
	if Config.SystemUUID != "" {
		return Config.SystemUUID
	}

	duid, err := systemUUID()
	if err != nil || duid == "" {
		log.Warn("Failed to get system UUID, using fallback")
		duid, err := uuid.NewRandom()
		if err != nil {
			log.WithError(err).Fatal("Failed to generate fallback system UUID")
		}
		Config.SystemUUID = duid.String()
	} else {
		Config.SystemUUID = duid
	}

	if err := SaveConfig(); err != nil {
		log.WithError(err).Fatal("Failed to save system UUID to config")
	}

	return Config.SystemUUID
}
