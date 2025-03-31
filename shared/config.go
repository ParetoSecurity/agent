package shared

import (
	"os"
	"path/filepath"

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
	log.Debugf("configPath: %s", ConfigPath)
}

func SaveConfig() error {

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

func ResetConfig() {
	Config = ParetoConfig{
		TeamID:        "",
		AuthToken:     "",
		DisableChecks: []string{},
	}
	SaveConfig()
}

func EnableCheck(checkUUID string) error {
	for i, check := range Config.DisableChecks {
		if check == checkUUID {
			Config.DisableChecks = append(Config.DisableChecks[:i], Config.DisableChecks[i+1:]...)
			return SaveConfig()
		}
	}
	return nil
}

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
