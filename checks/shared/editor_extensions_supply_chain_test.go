package shared

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditorExtensionsSupplyChain_Metadata(t *testing.T) {
	check := EditorExtensionsSupplyChain{}

	assert.Equal(t, "Editor extensions delay new releases", check.Name())
	assert.Equal(t, "d53a7b3b-2484-4d2f-966e-0ef1e9b64e48", check.UUID())
	assert.Equal(t, "Editor extension supply-chain protection is configured", check.PassedMessage())
	assert.Equal(t, "Editor extensions install new releases immediately", check.FailedMessage())
	assert.False(t, check.RequiresRoot())
}

func TestEditorExtensionsSupplyChain_IsRunnable(t *testing.T) {
	home := t.TempDir()

	assert.False(t, testEditorExtensionsSupplyChain(home, "linux", nil).IsRunnable())
	assert.True(t, testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"code": true}).IsRunnable())
	assert.True(t, testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"cursor": true}).IsRunnable())

	writeFile(t, filepath.Join(home, ".config", "VSCodium", "User", "settings.json"), "{}")
	assert.True(t, testEditorExtensionsSupplyChain(home, "linux", nil).IsRunnable())
}

func TestEditorExtensionsSupplyChain_IsRunnableWithFlatpak(t *testing.T) {
	home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".var", "app", "com.visualstudio.code"), 0o700))
	check := testEditorExtensionsSupplyChain(home, "linux", nil)

	assert.True(t, check.IsRunnable())
	require.NoError(t, check.Run())
	assert.False(t, check.Passed())
	assert.Equal(t, "~/.config/Code/User/settings.json is missing", check.Status())
}

func TestEditorExtensionsSupplyChain_RunPassesWithCooldowns(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "Code", "User", "settings.json"), `{
	// Wait a week before updating extensions.
	"extensions.autoUpdateDelay": 168,
}`)
	writeFile(t, filepath.Join(home, ".config", "Cursor", "User", "settings.json"), `{
	"extensions.installCooldownHours": 200
}`)
	check := testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"code": true, "cursor": true})
	check.Versions = map[string]string{"code": "1.125.0\nabcdef\nx64"}

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, stringsJoin(
		"~/.config/Code/User/settings.json delays Visual Studio Code extension updates",
		"~/.config/Cursor/User/settings.json delays Cursor extension updates",
	), check.Status())
}

func TestEditorExtensionsSupplyChain_RunPassesWithAutoUpdateDisabled(t *testing.T) {
	for _, value := range []string{"false", `"off"`} {
		home := t.TempDir()
		writeFile(t, filepath.Join(home, ".config", "Code", "User", "settings.json"), `{"extensions.autoUpdate": `+value+`}`)
		check := testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"code": true})

		require.NoError(t, check.Run())

		assert.True(t, check.Passed(), value)
		assert.Equal(t, "~/.config/Code/User/settings.json has automatic extension updates disabled", check.Status())
	}
}

func TestEditorExtensionsSupplyChain_RunFailsWithShortOrMissingCooldown(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "Code", "User", "settings.json"), `{"extensions.autoUpdateDelay": 24}`)
	writeFile(t, filepath.Join(home, ".config", "Cursor", "User", "settings.json"), `{
	// "extensions.installCooldownHours": 168,
	"editor.fontSize": 14
}`)
	check := testEditorExtensionsSupplyChain(home, "linux", nil)

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, stringsJoin(
		"~/.config/Code/User/settings.json extensions.autoUpdateDelay is below 168 hours; set it to at least 168",
		"~/.config/Cursor/User/settings.json extensions.installCooldownHours is not set; set it to at least 168 hours",
	), check.Status())
}

func TestEditorExtensionsSupplyChain_RunFailsWhenInstalledWithoutSettings(t *testing.T) {
	home := t.TempDir()
	check := testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"code-insiders": true, "cursor": true})

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, stringsJoin(
		"~/.config/Code - Insiders/User/settings.json is missing",
		"~/.config/Cursor/User/settings.json is missing",
	), check.Status())
}

func TestEditorExtensionsSupplyChain_VersionGate(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "Code", "User", "settings.json"), `{"extensions.autoUpdateDelay": 168}`)

	check := testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"code": true})
	check.Versions = map[string]string{"code": "1.124.2"}
	require.NoError(t, check.Run())
	assert.False(t, check.Passed())
	assert.Equal(t, "Visual Studio Code is older than 1.125.0 and does not enforce extensions.autoUpdateDelay", check.Status())

	check = testEditorExtensionsSupplyChain(home, "linux", map[string]bool{"code": true})
	check.RunCommand = func(string, ...string) ([]byte, error) { return nil, errors.New("failed") }
	require.NoError(t, check.Run())
	assert.False(t, check.Passed())
	assert.Equal(t, "Visual Studio Code version could not be determined, so extensions.autoUpdateDelay could not be verified", check.Status())

	check = testEditorExtensionsSupplyChain(home, "linux", nil)
	require.NoError(t, check.Run())
	assert.True(t, check.Passed(), "without the CLI on PATH the version is not gated")
}

func TestEditorExtensionsSupplyChain_FlatpakSettings(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".var", "app", "com.vscodium.codium", "config", "VSCodium", "User", "settings.json"), `{"extensions.autoUpdateDelay": 168}`)
	check := testEditorExtensionsSupplyChain(home, "linux", nil)

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, "~/.var/app/com.vscodium.codium/config/VSCodium/User/settings.json delays VSCodium extension updates", check.Status())
}

func TestEditorExtensionsSupplyChain_PlatformSettingsPaths(t *testing.T) {
	home := t.TempDir()
	appData := filepath.Join(home, "AppData", "Roaming")
	writeFile(t, filepath.Join(appData, "Cursor", "User", "settings.json"), `{"extensions.installCooldownHours": 168}`)
	check := testEditorExtensionsSupplyChain(home, "windows", nil)
	check.AppDataDir = appData

	require.NoError(t, check.Run())
	assert.True(t, check.Passed())
	assert.Equal(t, "~/AppData/Roaming/Cursor/User/settings.json delays Cursor extension updates", check.Status())

	home = t.TempDir()
	writeFile(t, filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"), `{"extensions.autoUpdateDelay": 168}`)
	check = testEditorExtensionsSupplyChain(home, "darwin", nil)

	require.NoError(t, check.Run())
	assert.True(t, check.Passed())
	assert.Equal(t, "~/Library/Application Support/Code/User/settings.json delays Visual Studio Code extension updates", check.Status())
}

func TestEditorExtensionsSupplyChain_InstallPathWithoutCLI(t *testing.T) {
	home := t.TempDir()
	check := testEditorExtensionsSupplyChain(home, "windows", nil)
	check.Getenv = func(name string) string {
		if name == "LOCALAPPDATA" {
			return filepath.Join(home, "AppData", "Local")
		}
		return ""
	}
	require.NoError(t, os.MkdirAll(filepath.Join(home, "AppData", "Local", "Programs", "Microsoft VS Code"), 0o700))

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, "~/AppData/Roaming/Code/User/settings.json is missing", check.Status())
}

func TestStripJSONCComments(t *testing.T) {
	contents := `{
	"url": "https://example.com", // trailing comment
	/* "extensions.autoUpdate": false, */
	"path": "C:\\\"quoted\"//not a comment"
}`

	assert.Equal(t, "{\n\t\"url\": \"https://example.com\", \n\t\n\t\"path\": \"C:\\\\\\\"quoted\\\"//not a comment\"\n}", stripJSONCComments(contents))
	assert.Equal(t, "", jsoncValue(contents, "extensions.autoUpdate"))
	assert.Equal(t, "https://example.com", jsoncValue(contents, "url"))
}

func testEditorExtensionsSupplyChain(home string, goos string, binaries map[string]bool) *EditorExtensionsSupplyChain {
	return &EditorExtensionsSupplyChain{
		HomeDir: home,
		GOOS:    goos,
		Getenv:  func(string) string { return "" },
		// Only paths under the temporary home count. Apps on the test
		// machine must not change the result.
		FileExists: func(path string) bool {
			if relative, err := filepath.Rel(home, path); err != nil || strings.HasPrefix(relative, "..") {
				return false
			}
			_, err := os.Stat(path)
			return err == nil
		},
		LookPath: func(name string) (string, error) {
			if binaries[name] {
				return "/usr/bin/" + name, nil
			}
			return "", errors.New("not found")
		},
		RunCommand: func(string, ...string) ([]byte, error) {
			return []byte("1.125.0\nabcdef\nx64\n"), nil
		},
	}
}
