package shared

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const minReleaseAgeSeconds = 7 * 24 * 60 * 60
const minReleaseAgeMinutes = 7 * 24 * 60
const minReleaseAgeDays = 7

type packageManagerConfig struct {
	paths      []string
	binaries   []string
	validate   func(string, string) []string
	missing    func([]string) string
	passDetail func(string) string
	// minimumVersion is set when only newer tool versions enforce the setting.
	minimumVersion string
	// setting names the cooldown key in version gate messages.
	setting string
	// usesSetting reports whether the config sets the version-gated key.
	usesSetting func(string) bool
}

// PackageManagerSupplyChain verifies package managers delay newly published packages.
type PackageManagerSupplyChain struct {
	passed     bool
	status     string
	HomeDir    string
	AppDataDir string
	GOOS       string
	ReadFile   func(string) ([]byte, error)
	FileExists func(string) bool
	LookPath   func(string) (string, error)
	Getenv     func(string) string
	RunCommand func(string, ...string) ([]byte, error)
	Versions   map[string]string
}

// Name returns the name of the check.
func (p *PackageManagerSupplyChain) Name() string {
	return "Package managers delay new releases"
}

// Run executes the check.
func (p *PackageManagerSupplyChain) Run() error {
	failures := p.validationFailures()
	if len(failures) > 0 {
		p.passed = false
		p.status = strings.Join(failures, "; ")
		return nil
	}

	p.passed = true
	p.status = strings.Join(p.passingDetails(), "; ")
	return nil
}

// Passed returns whether the check passed.
func (p *PackageManagerSupplyChain) Passed() bool {
	return p.passed
}

// IsRunnable returns whether any relevant package manager config or binary exists.
func (p *PackageManagerSupplyChain) IsRunnable() bool {
	if p.hasConfigFile() {
		return true
	}
	return p.hasPackageManagerBinary()
}

// UUID returns the UUID of the check.
func (p *PackageManagerSupplyChain) UUID() string {
	return "61bf7ef7-a3ee-4d66-a859-49c3ebeb1e7f"
}

// PassedMessage returns the message to return if the check passed.
func (p *PackageManagerSupplyChain) PassedMessage() string {
	return "Package manager supply-chain protection is configured"
}

// FailedMessage returns the message to return if the check failed.
func (p *PackageManagerSupplyChain) FailedMessage() string {
	return "Package managers install new releases immediately"
}

// RequiresRoot returns whether the check requires root access.
func (p *PackageManagerSupplyChain) RequiresRoot() bool {
	return false
}

// Status returns details for the current check state.
func (p *PackageManagerSupplyChain) Status() string {
	if p.status != "" {
		return p.status
	}
	if !p.IsRunnable() {
		return "No package manager config or binary found"
	}
	if p.Passed() {
		return p.PassedMessage()
	}
	return p.FailedMessage()
}

func (p *PackageManagerSupplyChain) validationFailures() []string {
	var failures []string

	for _, config := range p.configs() {
		if contents, path, ok := p.activeConfig(config.paths); ok {
			configFailures := config.validate(contents, path)
			if len(configFailures) == 0 && config.minimumVersion != "" && config.usesSetting(contents) {
				configFailures = p.versionGateFailures(config, path)
			}
			failures = append(failures, configFailures...)
			continue
		}
		if p.anyBinaryInstalled(config.binaries...) {
			failures = append(failures, config.missing(config.paths))
		}
	}

	return failures
}

func (p *PackageManagerSupplyChain) passingDetails() []string {
	var details []string

	for _, config := range p.configs() {
		if contents, path, ok := p.activeConfig(config.paths); ok && len(config.validate(contents, path)) == 0 {
			details = append(details, config.passDetail(path))
		}
	}
	if len(details) == 0 {
		return []string{p.PassedMessage()}
	}
	return details
}

func (p *PackageManagerSupplyChain) hasConfigFile() bool {
	for _, config := range p.configs() {
		for _, path := range config.paths {
			if p.fileExists(path) {
				return true
			}
		}
	}
	return false
}

func (p *PackageManagerSupplyChain) hasPackageManagerBinary() bool {
	for _, config := range p.configs() {
		if p.anyBinaryInstalled(config.binaries...) {
			return true
		}
	}
	return false
}

func (p *PackageManagerSupplyChain) firstInstalledBinary(names []string) string {
	for _, name := range names {
		if _, err := p.lookPath(name); err == nil {
			return name
		}
	}
	return ""
}

func (p *PackageManagerSupplyChain) versionGateFailures(config packageManagerConfig, path string) []string {
	binary := p.firstInstalledBinary(config.binaries)
	if binary == "" {
		return nil
	}
	description := p.displayPath(path) + " " + config.setting
	version := p.packageManagerVersion(binary)
	if version == "" {
		return []string{binary + " version could not be determined, so " + description + " could not be verified"}
	}
	if !versionAtLeast(version, config.minimumVersion) {
		return []string{binary + " is older than " + config.minimumVersion + " and does not enforce " + description}
	}
	return nil
}

func (p *PackageManagerSupplyChain) anyBinaryInstalled(names ...string) bool {
	for _, name := range names {
		if _, err := p.lookPath(name); err == nil {
			return true
		}
	}
	return false
}

func (p *PackageManagerSupplyChain) configs() []packageManagerConfig {
	home := p.homeDir()
	return []packageManagerConfig{
		// JavaScript / Node.js
		{
			paths:          []string{filepath.Join(home, ".npmrc")},
			binaries:       []string{"npm"},
			validate:       func(contents string, _ string) []string { return validateNpmrc(contents) },
			missing:        firstConfigPathMissing,
			passDetail:     func(string) string { return "~/.npmrc delays npm-compatible package releases and pins exact versions" },
			minimumVersion: "11.14.0",
			setting:        "min-release-age",
			usesSetting:    hasKey("min-release-age"),
		},
		{
			paths:          []string{filepath.Join(home, ".yarnrc.yml")},
			binaries:       []string{"yarn"},
			validate:       func(contents string, _ string) []string { return validateYarnrc(contents) },
			missing:        firstConfigPathMissing,
			passDetail:     func(string) string { return "~/.yarnrc.yml delays Yarn package releases" },
			minimumVersion: "4.11.0",
			setting:        "npmMinimalAgeGate",
			usesSetting:    hasKey("npmminimalagegate"),
		},
		{
			paths:      p.pnpmConfigPaths(),
			binaries:   []string{"pnpm"},
			validate:   validatePnpmConfig,
			missing:    pnpmConfigMissing,
			passDetail: func(path string) string { return path + " delays pnpm package releases" },
		},
		{
			paths:      []string{filepath.Join(home, ".bunfig.toml")},
			binaries:   []string{"bun"},
			validate:   func(contents string, _ string) []string { return validateBunfig(contents) },
			missing:    firstConfigPathMissing,
			passDetail: func(string) string { return "~/.bunfig.toml delays Bun package releases" },
		},

		// Python
		{
			paths:      []string{p.uvConfigPath()},
			binaries:   []string{"uv"},
			validate:   validateUv,
			missing:    firstConfigPathMissing,
			passDetail: func(path string) string { return path + " excludes Python packages newer than 7 days" },
		},
		{
			paths:      []string{filepath.Join(home, ".pypirc")},
			validate:   func(contents string, _ string) []string { return validatePypirc(contents) },
			missing:    firstConfigPathMissing,
			passDetail: func(string) string { return "~/.pypirc has no plaintext credentials" },
		},
		{
			paths:          p.pipConfigPaths(),
			binaries:       []string{"pip", "pip3"},
			validate:       validatePipConf,
			missing:        firstConfigPathMissing,
			passDetail:     func(path string) string { return path + " delays pip package releases" },
			minimumVersion: "26.1.0",
			setting:        "uploaded-prior-to",
			usesSetting:    hasKey("uploaded-prior-to"),
		},
		{
			paths:          p.poetryConfigPaths(),
			binaries:       []string{"poetry"},
			validate:       validatePoetryConfig,
			missing:        firstConfigPathMissing,
			passDetail:     func(path string) string { return path + " delays Poetry package releases" },
			minimumVersion: "2.4.0",
			setting:        "solver.min-release-age",
			usesSetting:    hasTomlKey("solver.min-release-age"),
		},
		{
			paths:          p.pdmConfigPaths(),
			binaries:       []string{"pdm"},
			validate:       validatePdmConfig,
			missing:        firstConfigPathMissing,
			passDetail:     func(path string) string { return path + " delays PDM package releases" },
			minimumVersion: "2.27.0",
			setting:        "strategy.exclude-newer",
			usesSetting:    hasTomlKey("strategy.exclude-newer", "strategy.exclude_newer"),
		},
		{
			paths:          p.condaConfigPaths(),
			binaries:       []string{"conda"},
			validate:       validateCondarc,
			missing:        firstConfigPathMissing,
			passDetail:     func(path string) string { return path + " delays conda package releases" },
			minimumVersion: "26.9.0",
			setting:        "exclude_newer",
			usesSetting:    hasKey("exclude_newer", "exclude-newer"),
		},

		// Ruby
		{
			paths:          []string{p.bundlerConfigPath()},
			binaries:       []string{"bundle", "bundler"},
			validate:       validateBundleConfig,
			missing:        firstConfigPathMissing,
			passDetail:     func(path string) string { return path + " delays Bundler package releases" },
			minimumVersion: "4.0.13",
			setting:        "BUNDLE_COOLDOWN",
			usesSetting:    hasKey("bundle_cooldown"),
		},

		// Elixir
		{
			paths:      []string{p.hexConfigPath()},
			binaries:   []string{"mix"},
			validate:   validateHexConfig,
			missing:    firstConfigPathMissing,
			passDetail: func(path string) string { return path + " delays Hex package releases" },
		},

		// Tool version managers
		{
			paths:      []string{p.miseConfigPath()},
			binaries:   []string{"mise"},
			validate:   validateMiseConfig,
			missing:    firstConfigPathMissing,
			passDetail: func(path string) string { return path + " delays mise package releases" },
		},
	}
}

func hasKey(keys ...string) func(string) bool {
	return func(contents string) bool {
		values := keyValuePairs(contents)
		for _, key := range keys {
			if _, ok := values[key]; ok {
				return true
			}
		}
		return false
	}
}

func hasTomlKey(keys ...string) func(string) bool {
	return func(contents string) bool {
		values := scopedKeyValuePairs(contents)
		for _, key := range keys {
			if _, ok := values[key]; ok {
				return true
			}
		}
		return false
	}
}

func firstConfigPathMissing(paths []string) string {
	return paths[0] + " is missing"
}

func pnpmConfigMissing(paths []string) string {
	return "pnpm config is missing (checked " + strings.Join(paths, ", ") + ")"
}

func (p *PackageManagerSupplyChain) readConfig(path string) (string, bool) {
	contents, err := p.readFile(path)
	if err != nil {
		return "", false
	}
	return string(contents), true
}

func (p *PackageManagerSupplyChain) activeConfig(paths []string) (string, string, bool) {
	for _, path := range paths {
		if contents, ok := p.readConfig(path); ok {
			return contents, path, true
		}
	}
	return "", "", false
}

func (p *PackageManagerSupplyChain) readFile(path string) ([]byte, error) {
	if p.ReadFile != nil {
		return p.ReadFile(path)
	}
	return os.ReadFile(path)
}

func (p *PackageManagerSupplyChain) fileExists(path string) bool {
	if p.FileExists != nil {
		return p.FileExists(path)
	}
	_, err := os.Stat(path)
	return err == nil
}

func (p *PackageManagerSupplyChain) lookPath(name string) (string, error) {
	if p.LookPath != nil {
		return p.LookPath(name)
	}
	return exec.LookPath(name)
}

func (p *PackageManagerSupplyChain) packageManagerVersion(name string) string {
	if version, ok := p.Versions[name]; ok {
		return versionToken(version)
	}
	path, err := p.lookPath(name)
	if err != nil {
		return ""
	}
	output, err := p.runCommand(path, "--version")
	if err != nil {
		return ""
	}
	return versionToken(string(output))
}

// versionToken extracts the version number from `--version` output such as
// "11.14.0", "pip 26.1.0 from ..." or "Bundler version 4.0.13".
func versionToken(output string) string {
	for _, field := range strings.Fields(output) {
		if field[0] >= '0' && field[0] <= '9' {
			return field
		}
	}
	return strings.TrimSpace(output)
}

func (p *PackageManagerSupplyChain) runCommand(name string, args ...string) ([]byte, error) {
	if p.RunCommand != nil {
		return p.RunCommand(name, args...)
	}
	return exec.Command(name, args...).Output()
}

func (p *PackageManagerSupplyChain) homeDir() string {
	if p.HomeDir != "" {
		return p.HomeDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func (p *PackageManagerSupplyChain) uvConfigPath() string {
	if configFile := p.env("UV_CONFIG_FILE"); configFile != "" {
		return configFile
	}
	if p.goos() == "windows" {
		return filepath.Join(p.appDataDir(), "uv", "uv.toml")
	}
	return filepath.Join(p.configHome(), "uv", "uv.toml")
}

// displayPath abbreviates paths under the home directory to ~/...
func (p *PackageManagerSupplyChain) displayPath(path string) string {
	home := p.homeDir()
	if relative, err := filepath.Rel(home, path); home != "" && err == nil && !strings.HasPrefix(relative, "..") {
		return "~/" + filepath.ToSlash(relative)
	}
	return path
}

func (p *PackageManagerSupplyChain) configHome() string {
	if configHome := p.env("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	return filepath.Join(p.homeDir(), ".config")
}

func (p *PackageManagerSupplyChain) appDataDir() string {
	if p.AppDataDir != "" {
		return p.AppDataDir
	}
	if appData := p.env("APPDATA"); appData != "" {
		return appData
	}
	return filepath.Join(p.homeDir(), "AppData", "Roaming")
}

func (p *PackageManagerSupplyChain) localAppDataDir() string {
	if localAppData := p.env("LOCALAPPDATA"); localAppData != "" {
		return localAppData
	}
	return filepath.Join(p.homeDir(), "AppData", "Local")
}

func (p *PackageManagerSupplyChain) applicationSupportDir(name string) string {
	return filepath.Join(p.homeDir(), "Library", "Application Support", name)
}

// pipConfigPaths lists pip's user config files. On macOS pip reads
// ~/Library/Application Support/pip when that directory exists and
// ~/.config/pip otherwise; ~/.pip is the legacy location.
// https://pip.pypa.io/en/stable/topics/configuration/#location
func (p *PackageManagerSupplyChain) pipConfigPaths() []string {
	if configFile := p.env("PIP_CONFIG_FILE"); configFile != "" {
		return []string{configFile}
	}
	switch p.goos() {
	case "windows":
		return []string{
			filepath.Join(p.appDataDir(), "pip", "pip.ini"),
			filepath.Join(p.homeDir(), "pip", "pip.ini"),
		}
	case "darwin":
		return []string{
			filepath.Join(p.configHome(), "pip", "pip.conf"),
			filepath.Join(p.applicationSupportDir("pip"), "pip.conf"),
			filepath.Join(p.homeDir(), ".pip", "pip.conf"),
		}
	default:
		return []string{
			filepath.Join(p.configHome(), "pip", "pip.conf"),
			filepath.Join(p.homeDir(), ".pip", "pip.conf"),
		}
	}
}

// poetryConfigPaths follows platformdirs, so macOS uses
// ~/Library/Application Support/pypoetry rather than ~/.config/pypoetry.
// ~/.config is kept for users who already configured it there.
// https://python-poetry.org/docs/configuration/
func (p *PackageManagerSupplyChain) poetryConfigPaths() []string {
	if configDir := p.env("POETRY_CONFIG_DIR"); configDir != "" {
		return []string{filepath.Join(configDir, "config.toml")}
	}
	switch p.goos() {
	case "windows":
		return []string{filepath.Join(p.appDataDir(), "pypoetry", "config.toml")}
	case "darwin":
		return []string{
			filepath.Join(p.applicationSupportDir("pypoetry"), "config.toml"),
			filepath.Join(p.configHome(), "pypoetry", "config.toml"),
		}
	default:
		return []string{filepath.Join(p.configHome(), "pypoetry", "config.toml")}
	}
}

// pdmConfigPaths follows platformdirs, so macOS uses
// ~/Library/Application Support/pdm and Windows %LOCALAPPDATA%\pdm\pdm.
// https://pdm-project.org/latest/usage/config/
func (p *PackageManagerSupplyChain) pdmConfigPaths() []string {
	switch p.goos() {
	case "windows":
		return []string{filepath.Join(p.localAppDataDir(), "pdm", "pdm", "config.toml")}
	case "darwin":
		return []string{
			filepath.Join(p.applicationSupportDir("pdm"), "config.toml"),
			filepath.Join(p.configHome(), "pdm", "config.toml"),
		}
	default:
		return []string{filepath.Join(p.configHome(), "pdm", "config.toml")}
	}
}

// condaConfigPaths lists the user-level .condarc locations conda searches.
// https://docs.conda.io/projects/conda/en/latest/user-guide/configuration/use-condarc.html
func (p *PackageManagerSupplyChain) condaConfigPaths() []string {
	if configFile := p.env("CONDARC"); configFile != "" {
		return []string{configFile}
	}
	return []string{
		filepath.Join(p.homeDir(), ".condarc"),
		filepath.Join(p.configHome(), "conda", ".condarc"),
		filepath.Join(p.homeDir(), ".conda", ".condarc"),
	}
}

// bundlerConfigPath is where `bundle config set --global` writes.
// https://bundler.io/man/bundle-config.1.html
func (p *PackageManagerSupplyChain) bundlerConfigPath() string {
	if configFile := p.env("BUNDLE_USER_CONFIG"); configFile != "" {
		return configFile
	}
	if userHome := p.env("BUNDLE_USER_HOME"); userHome != "" {
		return filepath.Join(userHome, "config")
	}
	return filepath.Join(p.homeDir(), ".bundle", "config")
}

// hexConfigPath is where `mix hex.config` writes.
// https://hexdocs.pm/hex/Mix.Tasks.Hex.Config.html
func (p *PackageManagerSupplyChain) hexConfigPath() string {
	if hexHome := p.env("HEX_HOME"); hexHome != "" {
		return filepath.Join(hexHome, "hex.config")
	}
	return filepath.Join(p.homeDir(), ".hex", "hex.config")
}

// miseConfigPath is mise's global config file.
// https://mise.jdx.dev/configuration.html
func (p *PackageManagerSupplyChain) miseConfigPath() string {
	if configFile := p.env("MISE_GLOBAL_CONFIG_FILE"); configFile != "" {
		return configFile
	}
	if configDir := p.env("MISE_CONFIG_DIR"); configDir != "" {
		return filepath.Join(configDir, "config.toml")
	}
	return filepath.Join(p.configHome(), "mise", "config.toml")
}

func (p *PackageManagerSupplyChain) pnpmConfigPaths() []string {
	if configHome := p.env("XDG_CONFIG_HOME"); configHome != "" {
		return []string{
			filepath.Join(configHome, "pnpm", "rc"),
			filepath.Join(configHome, "pnpm", "config.yaml"),
		}
	}

	switch p.goos() {
	case "windows":
		localAppData := p.env("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(p.homeDir(), "AppData", "Local")
		}
		return []string{
			filepath.Join(localAppData, "pnpm", "config", "rc"),
			filepath.Join(localAppData, "pnpm", "config", "config.yaml"),
		}
	case "darwin":
		return []string{
			filepath.Join(p.homeDir(), "Library", "Preferences", "pnpm", "rc"),
			filepath.Join(p.homeDir(), "Library", "Preferences", "pnpm", "config.yaml"),
			filepath.Join(p.homeDir(), ".config", "pnpm", "rc"),
			filepath.Join(p.homeDir(), ".config", "pnpm", "config.yaml"),
		}
	default:
		return []string{
			filepath.Join(p.homeDir(), ".config", "pnpm", "rc"),
			filepath.Join(p.homeDir(), ".config", "pnpm", "config.yaml"),
		}
	}
}

func (p *PackageManagerSupplyChain) pnpmConfigPath() string {
	return p.pnpmConfigPaths()[0]
}

func (p *PackageManagerSupplyChain) env(name string) string {
	if p.Getenv != nil {
		return p.Getenv(name)
	}
	return os.Getenv(name)
}

func (p *PackageManagerSupplyChain) goos() string {
	if p.GOOS != "" {
		return p.GOOS
	}
	return runtime.GOOS
}

func validateNpmrc(contents string) []string {
	values := keyValuePairs(contents)
	var failures []string

	if integerValue(values["min-release-age"]) < 7 && integerValue(values["minimum-release-age"]) < minReleaseAgeMinutes {
		failures = append(failures, "~/.npmrc release age is below 7 days; set either min-release-age >= 7 or minimum-release-age >= 10080")
	}
	if strings.ToLower(values["save-exact"]) != "true" {
		failures = append(failures, "~/.npmrc save-exact is not enabled")
	}

	return failures
}

func versionAtLeast(version string, minimumVersion string) bool {
	if version == "" || strings.Contains(version, "-") {
		return false
	}
	current := versionComponents(version)
	minimum := versionComponents(minimumVersion)
	length := len(current)
	if len(minimum) > length {
		length = len(minimum)
	}
	for index := 0; index < length; index++ {
		currentPart := 0
		if index < len(current) {
			currentPart = current[index]
		}
		minimumPart := 0
		if index < len(minimum) {
			minimumPart = minimum[index]
		}
		if currentPart != minimumPart {
			return currentPart > minimumPart
		}
	}
	return true
}

func versionComponents(version string) []int {
	parts := strings.Split(strings.TrimLeft(version, "vV"), ".")
	components := make([]int, 0, len(parts))
	for _, part := range parts {
		number := ""
		for _, character := range part {
			if character < '0' || character > '9' {
				break
			}
			number += string(character)
		}
		component, err := strconv.Atoi(number)
		if err != nil {
			component = 0
		}
		components = append(components, component)
	}
	return components
}

func validateYarnrc(contents string) []string {
	values := keyValuePairs(contents)
	if yarnAgeGateSeconds(values["npmminimalagegate"]) < minReleaseAgeSeconds {
		return []string{"~/.yarnrc.yml npmMinimalAgeGate is below 7 days; set it to at least 7d or 10080"}
	}
	return nil
}

func validatePnpmConfig(contents string, path string) []string {
	values := keyValuePairs(contents)
	releaseAge := integerValue(values["minimumreleaseage"])
	if releaseAge == 0 {
		releaseAge = integerValue(values["minimum-release-age"])
	}
	if releaseAge < minReleaseAgeMinutes {
		return []string{path + " minimumReleaseAge is below 10080 minutes"}
	}
	return nil
}

func validateBunfig(contents string) []string {
	values := scopedKeyValuePairs(contents)
	if cooldownSeconds(values["install.minimumReleaseAge"]) < minReleaseAgeSeconds {
		return []string{"~/.bunfig.toml minimumReleaseAge is below 604800 seconds"}
	}
	return nil
}

func validateUv(contents string, path string) []string {
	values := scopedKeyValuePairs(contents)
	value, ok := values["exclude-newer"]
	if !ok {
		value = values["pip.exclude-newer"]
	}
	return durationCooldownFailures(value, path+" exclude-newer", "7 days")
}

func validatePipConf(contents string, path string) []string {
	value := keyValuePairs(contents)["uploaded-prior-to"]
	switch {
	case value == "":
		return []string{path + " uploaded-prior-to is not set; set it to a duration like P7D"}
	case isAbsoluteTimestamp(value):
		return []string{path + " uploaded-prior-to is an absolute timestamp that goes stale; set it to a duration like P7D"}
	case iso8601DurationSeconds(value) < minReleaseAgeSeconds:
		return []string{path + " uploaded-prior-to is below 7 days; set it to at least P7D"}
	}
	return nil
}

func validatePoetryConfig(contents string, path string) []string {
	values := scopedKeyValuePairs(contents)
	if integerValue(values["solver.min-release-age"]) < minReleaseAgeDays {
		return []string{path + " solver.min-release-age is below 7 days; set it to at least 7"}
	}
	return nil
}

func validatePdmConfig(contents string, path string) []string {
	values := scopedKeyValuePairs(contents)
	value, ok := values["strategy.exclude-newer"]
	if !ok {
		value = values["strategy.exclude_newer"]
	}
	return durationCooldownFailures(value, path+" strategy.exclude-newer", "7d")
}

func validateCondarc(contents string, path string) []string {
	values := keyValuePairs(contents)
	value, ok := values["exclude_newer"]
	if !ok {
		value, ok = values["exclude-newer"]
	}
	if !ok {
		return []string{path + " exclude_newer is not set; set it to at least 7d"}
	}
	failures := durationCooldownFailures(value, path+" exclude_newer", "7d")
	if len(failures) == 0 && strings.ToLower(values["solver"]) != "classic" {
		failures = append(failures, path+" solver must be classic; the default solver does not support exclude_newer")
	}
	return failures
}

func validateBundleConfig(contents string, path string) []string {
	// Bundler rejects anything that is not a whole number of days.
	if integerValue(keyValuePairs(contents)["bundle_cooldown"]) < minReleaseAgeDays {
		return []string{path + " BUNDLE_COOLDOWN is below 7 days; set it to at least 7 with bundle config set --global cooldown 7"}
	}
	return nil
}

func validateHexConfig(contents string, path string) []string {
	if hexCooldownSeconds(contents) < minReleaseAgeSeconds {
		return []string{path + " cooldown is below 7 days; set it to at least 7d with mix hex.config cooldown 7d"}
	}
	return nil
}

func validateMiseConfig(contents string, path string) []string {
	values := scopedKeyValuePairs(contents)
	return durationCooldownFailures(values["settings.minimum_release_age"], path+" settings.minimum_release_age", "7d")
}

// durationCooldownFailures validates settings that accept relative durations
// or absolute timestamps, where only durations keep working as time passes.
func durationCooldownFailures(value string, setting string, hint string) []string {
	switch {
	case value == "":
		return []string{setting + " is not set; set it to at least " + hint}
	case isAbsoluteTimestamp(value):
		return []string{setting + " is an absolute timestamp that goes stale; set it to a relative duration like " + hint}
	case cooldownSeconds(value) < minReleaseAgeSeconds:
		return []string{setting + " is below 7 days; set it to at least " + hint}
	}
	return nil
}

func validatePypirc(contents string) []string {
	values := keyValuePairs(contents)
	for _, key := range []string{"password", "token", "api-token"} {
		if values[key] != "" {
			return []string{"~/.pypirc contains plaintext credentials"}
		}
	}
	return nil
}

func keyValuePairs(contents string) map[string]string {
	values := map[string]string{}

	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(stripComment(line))
		if trimmed == "" || strings.HasPrefix(trimmed, "[") {
			continue
		}
		index := strings.IndexAny(trimmed, "=:")
		if index < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(trimmed[:index]))
		value := unquote(strings.TrimSpace(trimmed[index+1:]))
		values[key] = value
	}

	return values
}

func scopedKeyValuePairs(contents string) map[string]string {
	values := map[string]string{}
	section := ""

	for _, line := range strings.Split(contents, "\n") {
		trimmed := strings.TrimSpace(stripComment(line))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.TrimSpace(strings.Trim(trimmed, "[]"))
			continue
		}
		index := strings.Index(trimmed, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:index])
		value := unquote(strings.TrimSpace(trimmed[index+1:]))
		if section != "" {
			key = section + "." + key
		}
		values[key] = value
	}

	return values
}

func integerValue(value string) int {
	parsed, err := strconv.Atoi(unquote(value))
	if err != nil {
		return 0
	}
	return parsed
}

// yarnAgeGateSeconds treats a bare npmMinimalAgeGate number as minutes; only
// values with a unit suffix are durations.
func yarnAgeGateSeconds(value string) int {
	normalized := strings.TrimSpace(unquote(value))
	if minutes, err := strconv.Atoi(normalized); err == nil {
		return minutes * 60
	}
	return cooldownSeconds(normalized)
}

func cooldownSeconds(value string) int {
	if seconds := iso8601DurationSeconds(value); seconds > 0 {
		return seconds
	}
	return durationSeconds(value)
}

// durationSeconds parses plain seconds or compound relative durations such
// as "7 days", "3d12h" or "1h30m".
func durationSeconds(value string) int {
	normalized := strings.ToLower(strings.TrimSpace(unquote(value)))
	if seconds, err := strconv.Atoi(normalized); err == nil {
		return seconds
	}

	total := 0
	rest := normalized
	for rest != "" {
		digits := leadingRun(rest, func(character rune) bool { return character >= '0' && character <= '9' })
		amount, err := strconv.Atoi(digits)
		if err != nil {
			return 0
		}
		rest = strings.TrimLeft(rest[len(digits):], " \t")
		unit := leadingRun(rest, func(character rune) bool { return character >= 'a' && character <= 'z' })
		seconds, ok := durationUnitSeconds[unit]
		if !ok {
			return 0
		}
		rest = strings.TrimLeft(rest[len(unit):], " \t")
		total += amount * seconds
	}
	return total
}

var durationUnitSeconds = map[string]int{
	"ms": 0, "millisecond": 0, "milliseconds": 0,
	"s": 1, "second": 1, "seconds": 1,
	"m": 60, "minute": 60, "minutes": 60,
	"h": 60 * 60, "hour": 60 * 60, "hours": 60 * 60,
	"d": 24 * 60 * 60, "day": 24 * 60 * 60, "days": 24 * 60 * 60,
	"w": minReleaseAgeSeconds, "week": minReleaseAgeSeconds, "weeks": minReleaseAgeSeconds,
	"mo": 30 * 24 * 60 * 60, "month": 30 * 24 * 60 * 60, "months": 30 * 24 * 60 * 60,
}

func leadingRun(value string, matches func(rune) bool) string {
	for index, character := range value {
		if !matches(character) {
			return value[:index]
		}
	}
	return value
}

var iso8601DurationPattern = regexp.MustCompile(`^p(?:(\d+)w)?(?:(\d+)d)?(?:t(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?)?$`)

// iso8601DurationSeconds parses durations such as "P7D" or "PT168H".
func iso8601DurationSeconds(value string) int {
	normalized := strings.ToLower(strings.TrimSpace(unquote(value)))
	match := iso8601DurationPattern.FindStringSubmatch(normalized)
	if len(normalized) < 2 || match == nil {
		return 0
	}
	total := 0
	for index, seconds := range []int{minReleaseAgeSeconds, 24 * 60 * 60, 60 * 60, 60, 1} {
		if amount, err := strconv.Atoi(match[index+1]); err == nil {
			total += amount * seconds
		}
	}
	return total
}

func isAbsoluteTimestamp(value string) bool {
	normalized := strings.TrimSpace(unquote(value))
	return normalized != "" && normalized[0] >= '0' && normalized[0] <= '9' && strings.Contains(normalized, "-")
}

var (
	hexQuotedCooldownPattern = regexp.MustCompile(`cooldown\s*,\s*<<"([^"]+)"`)
	hexDaysCooldownPattern   = regexp.MustCompile(`cooldown\s*,\s*(\d+)`)
)

// hexCooldownSeconds reads the cooldown from hex.config, which stores Erlang
// terms such as {cooldown,<<"7d">>}. or {cooldown,7}. (days).
func hexCooldownSeconds(contents string) int {
	if match := hexQuotedCooldownPattern.FindStringSubmatch(contents); match != nil {
		return durationSeconds(match[1])
	}
	if match := hexDaysCooldownPattern.FindStringSubmatch(contents); match != nil {
		return integerValue(match[1]) * 24 * 60 * 60
	}
	return 0
}

func stripComment(line string) string {
	doubleQuoted := false
	singleQuoted := false
	for index, character := range line {
		if character == '"' && !singleQuoted {
			doubleQuoted = !doubleQuoted
		}
		if character == '\'' && !doubleQuoted {
			singleQuoted = !singleQuoted
		}
		if !doubleQuoted && !singleQuoted && (character == '#' || character == ';') {
			return line[:index]
		}
	}
	return line
}

func unquote(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 2 {
		return trimmed
	}
	if (strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) ||
		(strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) {
		return trimmed[1 : len(trimmed)-1]
	}
	return trimmed
}
