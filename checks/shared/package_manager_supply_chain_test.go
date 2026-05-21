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

func TestPackageManagerSupplyChain_Metadata(t *testing.T) {
	check := PackageManagerSupplyChain{}

	assert.Equal(t, "Package managers delay new releases", check.Name())
	assert.Equal(t, "61bf7ef7-a3ee-4d66-a859-49c3ebeb1e7f", check.UUID())
	assert.Equal(t, "Package manager supply-chain protection is configured", check.PassedMessage())
	assert.Equal(t, "Package managers install new releases immediately", check.FailedMessage())
	assert.False(t, check.RequiresRoot())
}

func TestPackageManagerSupplyChain_IsRunnable(t *testing.T) {
	home := t.TempDir()
	check := testPackageManagerSupplyChain(home, nil, nil)

	assert.False(t, check.IsRunnable())

	check = testPackageManagerSupplyChain(home, nil, map[string]bool{"npm": true})
	assert.True(t, check.IsRunnable())

	check = testPackageManagerSupplyChain(home, nil, map[string]bool{"pnpm": true})
	assert.True(t, check.IsRunnable())

	writeFile(t, filepath.Join(home, ".npmrc"), "")
	check = testPackageManagerSupplyChain(home, nil, nil)
	assert.True(t, check.IsRunnable())
}

func TestPackageManagerSupplyChain_RunPassesWithProtectedConfigs(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".npmrc"), `
min-release-age=7
minimum-release-age=10080
save-exact=true
`)
	writeFile(t, filepath.Join(home, ".yarnrc.yml"), `
npmMinimalAgeGate: 10080
`)
	writeFile(t, filepath.Join(home, ".config", "pnpm", "config.yaml"), `
minimumReleaseAge: 10080
`)
	writeFile(t, filepath.Join(home, ".bunfig.toml"), `
[install]
minimumReleaseAge = 604800
`)
	writeFile(t, filepath.Join(home, ".config", "uv", "uv.toml"), `
[pip]
exclude-newer = "7d"
`)
	writeFile(t, filepath.Join(home, ".pypirc"), `
[pypi]
username = __token__
`)
	check := testPackageManagerSupplyChain(home, nil, nil)

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, stringsJoin(
		"~/.npmrc delays npm-compatible package releases and pins exact versions",
		"~/.yarnrc.yml delays Yarn package releases",
		filepath.Join(home, ".config", "pnpm", "config.yaml")+" delays pnpm package releases",
		"~/.bunfig.toml delays Bun package releases",
		filepath.Join(home, ".config", "uv", "uv.toml")+" excludes Python packages newer than 7 days",
		"~/.pypirc has no plaintext credentials",
	), check.Status())
}

func TestPackageManagerSupplyChain_RunFailsWithUnprotectedConfigs(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".npmrc"), `
min-release-age=6
minimum-release-age=10079
save-exact=false
`)
	writeFile(t, filepath.Join(home, ".yarnrc.yml"), `
npmMinimalAgeGate: 10079
`)
	writeFile(t, filepath.Join(home, ".config", "pnpm", "config.yaml"), `
minimumReleaseAge: 10079
`)
	writeFile(t, filepath.Join(home, ".bunfig.toml"), `
[install]
minimumReleaseAge = 604799
`)
	writeFile(t, filepath.Join(home, ".config", "uv", "uv.toml"), `
[pip]
exclude-newer = "6d"
`)
	writeFile(t, filepath.Join(home, ".pypirc"), `
[pypi]
password = pypi-secret
`)
	check := testPackageManagerSupplyChain(home, nil, nil)

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, stringsJoin(
		"~/.npmrc min-release-age is below 7 days",
		"~/.npmrc minimum-release-age is below 10080 minutes",
		"~/.npmrc save-exact is not enabled",
		"~/.yarnrc.yml npmMinimalAgeGate is below 10080 minutes",
		"pnpm minimumReleaseAge is below 10080 minutes",
		"~/.bunfig.toml minimumReleaseAge is below 604800 seconds",
		"uv exclude-newer is below 7 days",
		"~/.pypirc contains plaintext credentials",
	), check.Status())
}

func TestPackageManagerSupplyChain_RunFailsWhenBinaryConfigIsMissing(t *testing.T) {
	home := t.TempDir()
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{
		"npm":  true,
		"yarn": true,
		"pnpm": true,
		"bun":  true,
		"uv":   true,
	})

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, stringsJoin(
		filepath.Join(home, ".npmrc")+" is missing",
		filepath.Join(home, ".yarnrc.yml")+" is missing",
		filepath.Join(home, ".config", "pnpm", "config.yaml")+" is missing",
		filepath.Join(home, ".bunfig.toml")+" is missing",
		filepath.Join(home, ".config", "uv", "uv.toml")+" is missing",
	), check.Status())
}

func TestPackageManagerSupplyChain_WindowsUvConfigPath(t *testing.T) {
	home := filepath.Join("C:", "Users", "alice")
	appData := filepath.Join(home, "AppData", "Roaming")
	check := testPackageManagerSupplyChain(home, nil, nil)
	check.GOOS = "windows"
	check.AppDataDir = appData

	assert.Equal(t, filepath.Join(appData, "uv", "uv.toml"), check.uvConfigPath())
}

func TestPackageManagerSupplyChain_UvConfigPathUsesUvConfigFile(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, "uv.toml")
	check := testPackageManagerSupplyChain(home, nil, nil)
	check.Getenv = func(name string) string {
		if name == "UV_CONFIG_FILE" {
			return configPath
		}
		return ""
	}

	assert.Equal(t, configPath, check.uvConfigPath())
}

func TestPackageManagerSupplyChain_RunUsesUvXdgConfigHome(t *testing.T) {
	home := t.TempDir()
	configHome := filepath.Join(home, "xdg")
	uvConfig := filepath.Join(configHome, "uv", "uv.toml")
	writeFile(t, uvConfig, `
[pip]
exclude-newer = "7d"
`)
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{"uv": true})
	check.Getenv = func(name string) string {
		if name == "XDG_CONFIG_HOME" {
			return configHome
		}
		return ""
	}

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, uvConfig+" excludes Python packages newer than 7 days", check.Status())
}

func TestPackageManagerSupplyChain_RunAcceptsTopLevelUvExcludeNewer(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "uv", "uv.toml"), `
exclude-newer = "7d"
`)
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{"uv": true})

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, filepath.Join(home, ".config", "uv", "uv.toml")+" excludes Python packages newer than 7 days", check.Status())
}

func TestPackageManagerSupplyChain_RunUsesExplicitTopLevelUvExcludeNewer(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "uv", "uv.toml"), `
exclude-newer = "0d"

[pip]
exclude-newer = "7d"
`)
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{"uv": true})

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, "uv exclude-newer is below 7 days", check.Status())
}

func TestPackageManagerSupplyChain_PnpmConfigPath(t *testing.T) {
	home := t.TempDir()

	check := testPackageManagerSupplyChain(home, nil, nil)
	check.GOOS = "linux"
	assert.Equal(t, filepath.Join(home, ".config", "pnpm", "config.yaml"), check.pnpmConfigPath())

	check = testPackageManagerSupplyChain(home, nil, nil)
	check.GOOS = "darwin"
	assert.Equal(t, filepath.Join(home, "Library", "Preferences", "pnpm", "config.yaml"), check.pnpmConfigPath())

	check = testPackageManagerSupplyChain(home, nil, nil)
	check.GOOS = "windows"
	assert.Equal(t, filepath.Join(home, "AppData", "Local", "pnpm", "config", "config.yaml"), check.pnpmConfigPath())
}

func TestPackageManagerSupplyChain_PnpmConfigPathUsesXdgConfigHome(t *testing.T) {
	home := t.TempDir()
	configHome := filepath.Join(home, "xdg")
	check := testPackageManagerSupplyChain(home, nil, nil)
	check.Getenv = func(name string) string {
		if name == "XDG_CONFIG_HOME" {
			return configHome
		}
		return ""
	}

	assert.Equal(t, filepath.Join(configHome, "pnpm", "config.yaml"), check.pnpmConfigPath())
}

func TestKeyValuePairsPreservesQuotedCommentCharacters(t *testing.T) {
	values := keyValuePairs(`
single = 'abc#123;456' # comment
double = "abc#123;456" ; comment
plain = abc # comment
`)

	assert.Equal(t, "abc#123;456", values["single"])
	assert.Equal(t, "abc#123;456", values["double"])
	assert.Equal(t, "abc", values["plain"])
}

func testPackageManagerSupplyChain(home string, files map[string]string, binaries map[string]bool) *PackageManagerSupplyChain {
	return &PackageManagerSupplyChain{
		HomeDir: home,
		GOOS:    "linux",
		ReadFile: func(path string) ([]byte, error) {
			if files != nil {
				if contents, ok := files[path]; ok {
					return []byte(contents), nil
				}
			}
			return os.ReadFile(path)
		},
		FileExists: func(path string) bool {
			if files != nil {
				_, ok := files[path]
				return ok
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
		Getenv: func(string) string {
			return ""
		},
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
}

func stringsJoin(values ...string) string {
	return strings.Join(values, "; ")
}
