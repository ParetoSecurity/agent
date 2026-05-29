package shared

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageManagerSupplyChain_RealPackageManagerVersions(t *testing.T) {
	if os.Getenv("PARETO_TEST_PACKAGE_MANAGER_VERSIONS") == "" {
		t.Skip("set PARETO_TEST_PACKAGE_MANAGER_VERSIONS=1 to run real npm/pnpm version checks")
	}
	requireCommand(t, "npx")

	t.Run("npm 11.15.0 accepts comments", func(t *testing.T) {
		home := t.TempDir()
		npmrc := filepath.Join(home, ".npmrc")
		writeFile(t, npmrc, `
# full-line comment
min-release-age=7 # trailing comment
save-exact=true ; trailing comment
`)
		emptyNpmrc := filepath.Join(home, "empty.npmrc")
		writeFile(t, emptyNpmrc, "")

		assert.Equal(t, "7", runTool(t, home, emptyNpmrc, "npx", "--yes", "-p", "npm@11.15.0", "npm", "--userconfig", npmrc, "config", "get", "min-release-age"))
		assert.Equal(t, "true", runTool(t, home, emptyNpmrc, "npx", "--yes", "-p", "npm@11.15.0", "npm", "--userconfig", npmrc, "config", "get", "save-exact"))
	})

	t.Run("npm 11.13.0 does not apply min release age", func(t *testing.T) {
		home := t.TempDir()
		npmrc := filepath.Join(home, ".npmrc")
		writeFile(t, npmrc, `
min-release-age=7
save-exact=true
`)
		emptyNpmrc := filepath.Join(home, "empty.npmrc")
		writeFile(t, emptyNpmrc, "")

		assert.Equal(t, "null", runTool(t, home, emptyNpmrc, "npx", "--yes", "-p", "npm@11.13.0", "npm", "--userconfig", npmrc, "config", "get", "min-release-age"))
	})

	t.Run("pnpm 10.24.0 accepts rc comments", func(t *testing.T) {
		home := t.TempDir()
		xdgConfigHome := filepath.Join(home, "xdg")
		pnpmrc := filepath.Join(xdgConfigHome, "pnpm", "rc")
		writeFile(t, pnpmrc, `
; full-line comment
minimum-release-age=10080 # trailing comment
`)
		emptyNpmrc := filepath.Join(home, "empty.npmrc")
		writeFile(t, emptyNpmrc, "")

		assert.Equal(t, "10080", runToolWithEnv(t, home, emptyNpmrc, []string{"XDG_CONFIG_HOME=" + xdgConfigHome}, "npx", "--yes", "-p", "pnpm@10.24.0", "pnpm", "config", "get", "minimumReleaseAge"))
	})

	t.Run("pnpm 11.4.0 accepts config yaml comments", func(t *testing.T) {
		home := t.TempDir()
		xdgConfigHome := filepath.Join(home, "xdg")
		pnpmConfig := filepath.Join(xdgConfigHome, "pnpm", "config.yaml")
		writeFile(t, pnpmConfig, `
# full-line comment
minimumReleaseAge: 10080 # trailing comment
`)
		emptyNpmrc := filepath.Join(home, "empty.npmrc")
		writeFile(t, emptyNpmrc, "")

		assert.Equal(t, "10080", runToolWithEnv(t, home, emptyNpmrc, []string{"XDG_CONFIG_HOME=" + xdgConfigHome}, "npx", "--yes", "-p", "pnpm@11.4.0", "pnpm", "config", "get", "minimumReleaseAge"))
	})
}

func requireCommand(t *testing.T, name string) {
	t.Helper()
	_, err := exec.LookPath(name)
	require.NoError(t, err)
}

func runTool(t *testing.T, home string, npmUserConfig string, name string, args ...string) string {
	t.Helper()
	return runToolWithEnv(t, home, npmUserConfig, nil, name, args...)
}

func runToolWithEnv(t *testing.T, home string, npmUserConfig string, extraEnv []string, name string, args ...string) string {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = home
	globalNpmConfig := filepath.Join(home, "empty-global.npmrc")
	require.NoError(t, os.WriteFile(globalNpmConfig, []byte(""), 0o600))
	command.Env = append(os.Environ(),
		"HOME="+home,
		"NPM_CONFIG_AUDIT=false",
		"NPM_CONFIG_FUND=false",
		"NPM_CONFIG_GLOBALCONFIG="+globalNpmConfig,
		"NPM_CONFIG_LOGLEVEL=error",
		"NPM_CONFIG_UPDATE_NOTIFIER=false",
		"NPM_CONFIG_USERCONFIG="+npmUserConfig,
	)
	command.Env = append(command.Env, extraEnv...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	require.NoError(t, err, stderr.String())
	return strings.TrimSpace(string(output))
}
