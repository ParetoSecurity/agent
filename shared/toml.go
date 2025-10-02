package shared

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

// GetTOMLSectionKeyMock is a mock function for testing
var GetTOMLSectionKeyMock func(filepath, section, key string) (string, bool)

// GetTOMLSectionKey gets a key value under a specific section in a TOML file
// Returns the value and whether it exists
// Example: value, exists := GetTOMLSectionKey("/etc/greetd/config.toml", "initial_session", "user")
func GetTOMLSectionKey(filepath, section, key string) (string, bool) {
	if GetTOMLSectionKeyMock != nil {
		return GetTOMLSectionKeyMock(filepath, section, key)
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return "", false
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", false
	}

	tree, err := toml.Load(string(content))
	if err != nil {
		return "", false
	}

	if !tree.Has(section) {
		return "", false
	}

	sectionTree := tree.Get(section)
	if sectionTree == nil {
		return "", false
	}

	fullPath := fmt.Sprintf("%s.%s", section, key)
	if !tree.Has(fullPath) {
		return "", false
	}

	value := tree.Get(fullPath)
	if value == nil {
		return "", false
	}

	return fmt.Sprintf("%v", value), true
}
