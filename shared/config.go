package shared

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/caarlos0/log"
	"github.com/pelletier/go-toml"
)

var Config ParetoConfig
var ConfigPath string

type ParetoConfig struct {
	TeamID        string
	AuthToken     string
	DisableChecks []string
}

func init() {
	states = make(map[string]LastState)
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

// SaveConfig writes the current configuration to the config file.
func SaveConfig() error {

	file, err := os.Create(ConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(Config)
}

// LoadConfig reads the configuration file and populates the global Config.
// If the config file doesn't exist, it creates one with default values.
func LoadConfig() error {
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		// Create the config file with default values if it doesn't exist
		err := ResetConfig()
		if err != nil {
			return err
		}
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

// ResetConfig resets the configuration to default values and saves it to disk.
func ResetConfig() error {
	Config = ParetoConfig{
		TeamID:        "",
		AuthToken:     "",
		DisableChecks: []string{},
	}
	if err := SaveConfig(); err != nil {
		return err
	}
	return nil
}

// EnableCheck removes a check from the disabled checks list and saves the configuration.
func EnableCheck(checkUUID string) error {
	for i, check := range Config.DisableChecks {
		if check == checkUUID {
			Config.DisableChecks = append(Config.DisableChecks[:i], Config.DisableChecks[i+1:]...)
			return SaveConfig()
		}
	}
	return nil
}

// DisableCheck adds a check to the disabled checks list and saves the configuration.
func DisableCheck(checkUUID string) error {
	for _, check := range Config.DisableChecks {
		if check == checkUUID {
			return nil
		}
	}
	Config.DisableChecks = append(Config.DisableChecks, checkUUID)
	return SaveConfig()
}

// IsCheckDisabled checks if a given check UUID is present in the list of disabled checks.
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
