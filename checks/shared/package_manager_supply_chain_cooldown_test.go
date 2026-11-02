package shared

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageManagerSupplyChain_Cooldowns(t *testing.T) {
	tests := []struct {
		name     string
		file     []string
		contents string
		binaries map[string]bool
		versions map[string]string
		passed   bool
		status   string
	}{
		{
			name:     "yarn duration",
			file:     []string{".yarnrc.yml"},
			contents: `npmMinimalAgeGate: "7d"`,
			binaries: map[string]bool{"yarn": true},
			versions: map[string]string{"yarn": "4.11.0"},
			passed:   true,
			status:   "~/.yarnrc.yml delays Yarn package releases",
		},
		{
			name:     "yarn low duration",
			file:     []string{".yarnrc.yml"},
			contents: `npmMinimalAgeGate: "6d"`,
			status:   "~/.yarnrc.yml npmMinimalAgeGate is below 7 days; set it to at least 7d or 10080",
		},
		{
			name:     "legacy yarn",
			file:     []string{".yarnrc.yml"},
			contents: `npmMinimalAgeGate: 10080`,
			binaries: map[string]bool{"yarn": true},
			versions: map[string]string{"yarn": "1.22.22"},
			status:   "yarn is older than 4.11.0 and does not enforce ~/.yarnrc.yml npmMinimalAgeGate",
		},
		{
			name:     "bun duration",
			file:     []string{".bunfig.toml"},
			contents: "[install]\nminimumReleaseAge = \"7d\"",
			binaries: map[string]bool{"bun": true},
			passed:   true,
			status:   "~/.bunfig.toml delays Bun package releases",
		},
		{
			name:     "uv iso duration",
			file:     []string{".config", "uv", "uv.toml"},
			contents: `exclude-newer = "P7D"`,
			binaries: map[string]bool{"uv": true},
			passed:   true,
			status:   "{path} excludes Python packages newer than 7 days",
		},
		{
			name:     "uv stale timestamp",
			file:     []string{".config", "uv", "uv.toml"},
			contents: `exclude-newer = "2026-03-27T00:00:00Z"`,
			status:   "{path} exclude-newer is an absolute timestamp that goes stale; set it to a relative duration like 7 days",
		},
		{
			name:     "pip",
			file:     []string{".config", "pip", "pip.conf"},
			contents: "[install]\nuploaded-prior-to = P7D",
			binaries: map[string]bool{"pip3": true},
			versions: map[string]string{"pip3": "pip 26.1.0 from /usr/lib/pip (python 3.13)"},
			passed:   true,
			status:   "{path} delays pip package releases",
		},
		{
			name:     "pip low",
			file:     []string{".config", "pip", "pip.conf"},
			contents: "[install]\nuploaded-prior-to = P3D",
			status:   "{path} uploaded-prior-to is below 7 days; set it to at least P7D",
		},
		{
			name:     "pip stale timestamp",
			file:     []string{".config", "pip", "pip.conf"},
			contents: "[install]\nuploaded-prior-to = 2026-01-01T00:00:00Z",
			status:   "{path} uploaded-prior-to is an absolute timestamp that goes stale; set it to a duration like P7D",
		},
		{
			name:     "old pip",
			file:     []string{".config", "pip", "pip.conf"},
			contents: "[global]\nuploaded-prior-to = P7D",
			binaries: map[string]bool{"pip": true},
			versions: map[string]string{"pip": "pip 26.0.0 from /usr/lib/pip (python 3.13)"},
			status:   "pip is older than 26.1.0 and does not enforce ~/.config/pip/pip.conf uploaded-prior-to",
		},
		{
			name:     "poetry",
			file:     []string{".config", "pypoetry", "config.toml"},
			contents: "[solver]\nmin-release-age = 7",
			binaries: map[string]bool{"poetry": true},
			versions: map[string]string{"poetry": "Poetry (version 2.4.0)"},
			passed:   true,
			status:   "{path} delays Poetry package releases",
		},
		{
			name:     "poetry low",
			file:     []string{".config", "pypoetry", "config.toml"},
			contents: "[solver]\nmin-release-age = 3",
			status:   "{path} solver.min-release-age is below 7 days; set it to at least 7",
		},
		{
			name:     "pdm",
			file:     []string{".config", "pdm", "config.toml"},
			contents: "[strategy]\nexclude-newer = \"7d\"",
			binaries: map[string]bool{"pdm": true},
			versions: map[string]string{"pdm": "PDM, version 2.27.0"},
			passed:   true,
			status:   "{path} delays PDM package releases",
		},
		{
			name:     "pdm low",
			file:     []string{".config", "pdm", "config.toml"},
			contents: "[strategy]\nexclude-newer = \"6d\"",
			status:   "{path} strategy.exclude-newer is below 7 days; set it to at least 7d",
		},
		{
			name:     "conda",
			file:     []string{".condarc"},
			contents: "solver: classic\nexclude_newer: 7d",
			binaries: map[string]bool{"conda": true},
			versions: map[string]string{"conda": "conda 26.9.0"},
			passed:   true,
			status:   "{path} delays conda package releases",
		},
		{
			name:     "conda low",
			file:     []string{".condarc"},
			contents: "solver: classic\nexclude_newer: 1d",
			status:   "{path} exclude_newer is below 7 days; set it to at least 7d",
		},
		{
			name:     "conda default solver",
			file:     []string{".condarc"},
			contents: "exclude_newer: 7d",
			status:   "{path} solver must be classic; the default solver does not support exclude_newer",
		},
		{
			name:     "bundler",
			file:     []string{".bundle", "config"},
			contents: "---\nBUNDLE_COOLDOWN: \"7\"",
			binaries: map[string]bool{"bundle": true},
			versions: map[string]string{"bundle": "Bundler version 4.0.13"},
			passed:   true,
			status:   "{path} delays Bundler package releases",
		},
		{
			name:     "old bundler",
			file:     []string{".bundle", "config"},
			contents: "---\nBUNDLE_COOLDOWN: \"7\"",
			binaries: map[string]bool{"bundle": true},
			versions: map[string]string{"bundle": "Bundler version 2.6.7"},
			status:   "bundle is older than 4.0.13 and does not enforce ~/.bundle/config BUNDLE_COOLDOWN",
		},
		{
			name:     "bundler low",
			file:     []string{".bundle", "config"},
			contents: "---\nBUNDLE_COOLDOWN: \"3\"",
			status:   "{path} BUNDLE_COOLDOWN is below 7 days; set it to at least 7 with bundle config set --global cooldown 7",
		},
		{
			name:     "hex duration",
			file:     []string{".hex", "hex.config"},
			contents: `{cooldown,<<"7d">>}.`,
			binaries: map[string]bool{"mix": true},
			passed:   true,
			status:   "{path} delays Hex package releases",
		},
		{
			name:     "hex days",
			file:     []string{".hex", "hex.config"},
			contents: `{cooldown,7}.`,
			passed:   true,
			status:   "{path} delays Hex package releases",
		},
		{
			name:     "hex low",
			file:     []string{".hex", "hex.config"},
			contents: `{cooldown,<<"3d">>}.`,
			status:   "{path} cooldown is below 7 days; set it to at least 7d with mix hex.config cooldown 7d",
		},
		{
			name:     "mise",
			file:     []string{".config", "mise", "config.toml"},
			contents: "[settings]\nminimum_release_age = \"7d\"",
			binaries: map[string]bool{"mise": true},
			passed:   true,
			status:   "{path} delays mise package releases",
		},
		{
			name:     "mise unset",
			file:     []string{".config", "mise", "config.toml"},
			contents: "[tools]\nnode = \"22\"",
			status:   "{path} settings.minimum_release_age is not set; set it to at least 7d",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(append([]string{home}, test.file...)...)
			writeFile(t, path, test.contents)
			check := testPackageManagerSupplyChain(home, nil, test.binaries)
			check.Versions = test.versions

			require.NoError(t, check.Run())

			assert.Equal(t, test.passed, check.Passed())
			assert.Equal(t, replacePath(test.status, path), check.Status())
		})
	}
}

func TestPackageManagerSupplyChain_RunFailsWhenNewManagerConfigIsMissing(t *testing.T) {
	home := t.TempDir()
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{
		"pip3":    true,
		"poetry":  true,
		"pdm":     true,
		"conda":   true,
		"bundler": true,
		"mix":     true,
		"mise":    true,
	})

	assert.True(t, check.IsRunnable())
	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, stringsJoin(
		filepath.Join(home, ".config", "pip", "pip.conf")+" is missing",
		filepath.Join(home, ".config", "pypoetry", "config.toml")+" is missing",
		filepath.Join(home, ".config", "pdm", "config.toml")+" is missing",
		filepath.Join(home, ".condarc")+" is missing",
		filepath.Join(home, ".bundle", "config")+" is missing",
		filepath.Join(home, ".hex", "hex.config")+" is missing",
		filepath.Join(home, ".config", "mise", "config.toml")+" is missing",
	), check.Status())
}

func TestPackageManagerSupplyChain_DarwinPoetryAndPdmUseApplicationSupport(t *testing.T) {
	home := t.TempDir()
	poetryConfig := filepath.Join(home, "Library", "Application Support", "pypoetry", "config.toml")
	pdmConfig := filepath.Join(home, "Library", "Application Support", "pdm", "config.toml")
	writeFile(t, poetryConfig, "[solver]\nmin-release-age = 7")
	writeFile(t, pdmConfig, "[strategy]\nexclude-newer = \"7d\"")
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{"poetry": true, "pdm": true})
	check.GOOS = "darwin"
	check.Versions = map[string]string{"poetry": "Poetry (version 2.4.0)", "pdm": "PDM, version 2.27.0"}

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, stringsJoin(
		poetryConfig+" delays Poetry package releases",
		pdmConfig+" delays PDM package releases",
	), check.Status())
}

func TestPackageManagerSupplyChain_DarwinPoetryAndPdmMissingNamesApplicationSupport(t *testing.T) {
	home := t.TempDir()
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{"poetry": true, "pdm": true})
	check.GOOS = "darwin"

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, stringsJoin(
		filepath.Join(home, "Library", "Application Support", "pypoetry", "config.toml")+" is missing",
		filepath.Join(home, "Library", "Application Support", "pdm", "config.toml")+" is missing",
	), check.Status())
}

func TestPackageManagerSupplyChain_DarwinPoetryFallsBackToXdgConfig(t *testing.T) {
	home := t.TempDir()
	poetryConfig := filepath.Join(home, ".config", "pypoetry", "config.toml")
	writeFile(t, poetryConfig, "[solver]\nmin-release-age = 7")
	check := testPackageManagerSupplyChain(home, nil, nil)
	check.GOOS = "darwin"

	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, poetryConfig+" delays Poetry package releases", check.Status())
}

func TestPackageManagerSupplyChain_DarwinOldPoetryNamesApplicationSupport(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "Library", "Application Support", "pypoetry", "config.toml"), "[solver]\nmin-release-age = 7")
	check := testPackageManagerSupplyChain(home, nil, map[string]bool{"poetry": true})
	check.GOOS = "darwin"
	check.Versions = map[string]string{"poetry": "Poetry (version 2.3.0)"}

	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, "poetry is older than 2.4.0 and does not enforce ~/Library/Application Support/pypoetry/config.toml solver.min-release-age", check.Status())
}

func TestPackageManagerSupplyChain_NewManagerConfigPaths(t *testing.T) {
	home := "/home/user"
	check := testPackageManagerSupplyChain(home, nil, nil)

	check.GOOS = "darwin"
	assert.Equal(t, []string{
		filepath.Join(home, ".config", "pip", "pip.conf"),
		filepath.Join(home, "Library", "Application Support", "pip", "pip.conf"),
		filepath.Join(home, ".pip", "pip.conf"),
	}, check.pipConfigPaths())
	assert.Equal(t, filepath.Join(home, "Library", "Application Support", "pypoetry", "config.toml"), check.poetryConfigPaths()[0])
	assert.Equal(t, filepath.Join(home, "Library", "Application Support", "pdm", "config.toml"), check.pdmConfigPaths()[0])

	check.GOOS = "windows"
	check.AppDataDir = filepath.Join(home, "Roaming")
	assert.Equal(t, filepath.Join(home, "Roaming", "pip", "pip.ini"), check.pipConfigPaths()[0])
	assert.Equal(t, []string{filepath.Join(home, "Roaming", "pypoetry", "config.toml")}, check.poetryConfigPaths())
	assert.Equal(t, []string{filepath.Join(home, "AppData", "Local", "pdm", "pdm", "config.toml")}, check.pdmConfigPaths())

	env := map[string]string{
		"PIP_CONFIG_FILE":         "/etc/custom/pip.conf",
		"POETRY_CONFIG_DIR":       "/etc/poetry",
		"CONDARC":                 "/etc/condarc",
		"BUNDLE_USER_CONFIG":      "/etc/bundle",
		"HEX_HOME":                "/opt/hex",
		"MISE_GLOBAL_CONFIG_FILE": "/etc/mise.toml",
	}
	check.Getenv = func(name string) string { return env[name] }
	assert.Equal(t, []string{"/etc/custom/pip.conf"}, check.pipConfigPaths())
	assert.Equal(t, []string{filepath.Join("/etc/poetry", "config.toml")}, check.poetryConfigPaths())
	assert.Equal(t, []string{"/etc/condarc"}, check.condaConfigPaths())
	assert.Equal(t, "/etc/bundle", check.bundlerConfigPath())
	assert.Equal(t, filepath.Join("/opt/hex", "hex.config"), check.hexConfigPath())
	assert.Equal(t, "/etc/mise.toml", check.miseConfigPath())
}

func TestPackageManagerSupplyChain_DurationParsing(t *testing.T) {
	assert.Equal(t, 604800, durationSeconds("604800"))
	assert.Equal(t, 604800, durationSeconds("7 days"))
	assert.Equal(t, 302400, durationSeconds("3d12h"))
	assert.Equal(t, 5400, durationSeconds("1h 30m"))
	assert.Equal(t, 2592000, durationSeconds("1mo"))
	assert.Equal(t, 0, durationSeconds("7 fortnights"))
	assert.Equal(t, 604800, iso8601DurationSeconds("P7D"))
	assert.Equal(t, 604800, iso8601DurationSeconds("PT168H"))
	assert.Equal(t, 604800, iso8601DurationSeconds("P1W"))
	assert.Equal(t, 0, iso8601DurationSeconds("P"))
	assert.Equal(t, 604800, yarnAgeGateSeconds("10080"))
	assert.Equal(t, 604800, yarnAgeGateSeconds("7d"))
	assert.True(t, isAbsoluteTimestamp("2026-03-27"))
	assert.False(t, isAbsoluteTimestamp("7d"))
}

func TestPackageManagerSupplyChain_VersionToken(t *testing.T) {
	assert.Equal(t, "11.10.0", versionToken("11.10.0\n"))
	assert.Equal(t, "26.1.0", versionToken("pip 26.1.0 from /usr/lib/pip (python 3.13)"))
	assert.Equal(t, "2.6.7", versionToken("Bundler version 2.6.7"))
	assert.Equal(t, "26.9.0", versionToken("conda 26.9.0"))
	assert.Equal(t, "", versionToken("  \n"))
}

func replacePath(status string, path string) string {
	if len(status) >= 6 && status[:6] == "{path}" {
		return path + status[6:]
	}
	return status
}
