package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTOMLSectionKey(t *testing.T) {
	tests := []struct {
		name        string
		tomlContent string
		section     string
		key         string
		wantValue   string
		wantExists  bool
	}{
		{
			name: "key exists in section",
			tomlContent: `[database]
host = "localhost"
port = 5432

[cache]
redis_host = "127.0.0.1"`,
			section:    "database",
			key:        "host",
			wantValue:  "localhost",
			wantExists: true,
		},
		{
			name: "greetd initial_session example",
			tomlContent: `[initial_session]
command = "sway"
user = "jane"

[default_session]
command = "tuigreet --cmd sway"`,
			section:    "initial_session",
			key:        "user",
			wantValue:  "jane",
			wantExists: true,
		},
		{
			name: "greetd default_session command",
			tomlContent: `[initial_session]
command = "sway"
user = "jane"

[default_session]
command = "tuigreet --cmd sway"`,
			section:    "default_session",
			key:        "command",
			wantValue:  "tuigreet --cmd sway",
			wantExists: true,
		},
		{
			name: "key with spaces in value",
			tomlContent: `[settings]
app_name = "My Application"
debug = true`,
			section:    "settings",
			key:        "app_name",
			wantValue:  "My Application",
			wantExists: true,
		},
		{
			name: "boolean value",
			tomlContent: `[settings]
app_name = "My Application"
debug = true`,
			section:    "settings",
			key:        "debug",
			wantValue:  "true",
			wantExists: true,
		},
		{
			name: "integer value",
			tomlContent: `[database]
host = "localhost"
port = 5432`,
			section:    "database",
			key:        "port",
			wantValue:  "5432",
			wantExists: true,
		},
		{
			name: "section does not exist",
			tomlContent: `[database]
host = "localhost"`,
			section:    "NonExistent",
			key:        "host",
			wantValue:  "",
			wantExists: false,
		},
		{
			name: "key does not exist in section",
			tomlContent: `[database]
host = "localhost"`,
			section:    "database",
			key:        "port",
			wantValue:  "",
			wantExists: false,
		},
		{
			name:        "empty toml file",
			tomlContent: ``,
			section:     "database",
			key:         "host",
			wantValue:   "",
			wantExists:  false,
		},
		{
			name: "nested tables",
			tomlContent: `[terminal]
font = "MesloLGS NF"

[terminal.colors]
background = "#1e1e2e"
foreground = "#cdd6f4"`,
			section:    "terminal",
			key:        "font",
			wantValue:  "MesloLGS NF",
			wantExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary TOML file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.toml")
			err := os.WriteFile(tmpFile, []byte(tt.tomlContent), 0644)
			assert.NoError(t, err)

			// Test the function
			value, exists := GetTOMLSectionKey(tmpFile, tt.section, tt.key)
			assert.Equal(t, tt.wantExists, exists)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestGetTOMLSectionKey_FileNotExists(t *testing.T) {
	value, exists := GetTOMLSectionKey("/non/existent/file.toml", "Section", "key")
	assert.False(t, exists)
	assert.Empty(t, value)
}

func TestGetTOMLSectionKey_InvalidTOML(t *testing.T) {
	// Create temporary invalid TOML file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.toml")
	err := os.WriteFile(tmpFile, []byte("[Invalid\nno closing bracket"), 0644)
	assert.NoError(t, err)

	value, exists := GetTOMLSectionKey(tmpFile, "Invalid", "key")
	assert.False(t, exists)
	assert.Empty(t, value)
}
