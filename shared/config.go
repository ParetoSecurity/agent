package shared

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/log"
	"github.com/pelletier/go-toml"
)

var Config ParetoConfig
var configPath string

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
	configPath = filepath.Join(homeDir, ".config", "pareto.toml")
	log.Debugf("configPath: %s", configPath)
}

func SaveConfig() error {

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(Config)
}

func LoadConfig() error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := SaveConfig(); err != nil {
			return err
		}
		return nil
	}
	file, err := os.Open(configPath)
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
