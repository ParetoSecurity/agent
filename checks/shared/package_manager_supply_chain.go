package shared

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const minReleaseAgeSeconds = 7 * 24 * 60 * 60
const minReleaseAgeMinutes = 7 * 24 * 60

type packageManagerConfig struct {
	paths      []string
	binaries   []string
	validate   func(string, string) []string
	missing    func([]string) string
	passDetail func(string) string
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
			failures = append(failures, config.validate(contents, path)...)
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
	return p.anyBinaryInstalled("npm", "yarn", "pnpm", "bun", "uv")
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
		{
			paths:      []string{filepath.Join(home, ".npmrc")},
			binaries:   []string{"npm"},
			validate:   p.validateNpmConfig,
			missing:    firstConfigPathMissing,
			passDetail: func(string) string { return "~/.npmrc delays npm-compatible package releases and pins exact versions" },
		},
		{
			paths:      []string{filepath.Join(home, ".yarnrc.yml")},
			binaries:   []string{"yarn"},
			validate:   func(contents string, _ string) []string { return validateYarnrc(contents) },
			missing:    firstConfigPathMissing,
			passDetail: func(string) string { return "~/.yarnrc.yml delays Yarn package releases" },
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
		{
			paths:      []string{p.uvConfigPath()},
			binaries:   []string{"uv"},
			validate:   func(contents string, _ string) []string { return validateUv(contents) },
			missing:    firstConfigPathMissing,
			passDetail: func(path string) string { return path + " excludes Python packages newer than 7 days" },
		},
		{
			paths:      []string{filepath.Join(home, ".pypirc")},
			validate:   func(contents string, _ string) []string { return validatePypirc(contents) },
			missing:    firstConfigPathMissing,
			passDetail: func(string) string { return "~/.pypirc has no plaintext credentials" },
		},
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
		return version
	}
	path, err := p.lookPath(name)
	if err != nil {
		return ""
	}
	output, err := p.runCommand(path, "--version")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
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
		appData := p.AppDataDir
		if appData == "" {
			appData = p.env("APPDATA")
		}
		if appData == "" {
			appData = filepath.Join(p.homeDir(), "AppData", "Roaming")
		}
		return filepath.Join(appData, "uv", "uv.toml")
	}
	configHome := p.env("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(p.homeDir(), ".config")
	}
	return filepath.Join(configHome, "uv", "uv.toml")
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

func (p *PackageManagerSupplyChain) validateNpmConfig(contents string, _ string) []string {
	failures := validateNpmrc(contents)
	if len(failures) != 0 || !p.anyBinaryInstalled("npm") || !npmConfigUsesMinReleaseAge(contents) {
		return failures
	}
	version := p.packageManagerVersion("npm")
	if version == "" {
		failures = append(failures, "npm version could not be determined, so ~/.npmrc min-release-age could not be verified")
		return failures
	}
	if !versionAtLeast(version, "11.14.0") {
		failures = append(failures, "npm is older than 11.14.0 and does not enforce ~/.npmrc min-release-age")
	}
	return failures
}

func npmConfigUsesMinReleaseAge(contents string) bool {
	_, ok := keyValuePairs(contents)["min-release-age"]
	return ok
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
	if integerValue(values["npmminimalagegate"]) < minReleaseAgeMinutes {
		return []string{"~/.yarnrc.yml npmMinimalAgeGate is below 10080 minutes"}
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
	if integerValue(values["install.minimumReleaseAge"]) < minReleaseAgeSeconds {
		return []string{"~/.bunfig.toml minimumReleaseAge is below 604800 seconds"}
	}
	return nil
}

func validateUv(contents string) []string {
	values := scopedKeyValuePairs(contents)
	var excludeNewer int
	if value, ok := values["exclude-newer"]; ok {
		excludeNewer = durationSeconds(value)
	} else {
		excludeNewer = durationSeconds(values["pip.exclude-newer"])
	}
	if excludeNewer < minReleaseAgeSeconds {
		return []string{"uv exclude-newer is below 7 days"}
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

func durationSeconds(value string) int {
	normalized := strings.ToLower(strings.TrimSpace(unquote(value)))
	if seconds, err := strconv.Atoi(normalized); err == nil {
		return seconds
	}

	number := ""
	for _, character := range normalized {
		if character < '0' || character > '9' {
			break
		}
		number += string(character)
	}
	amount, err := strconv.Atoi(number)
	if err != nil {
		return 0
	}

	switch strings.TrimSpace(strings.TrimPrefix(normalized, number)) {
	case "s", "second", "seconds":
		return amount
	case "m", "minute", "minutes":
		return amount * 60
	case "h", "hour", "hours":
		return amount * 60 * 60
	case "d", "day", "days":
		return amount * 24 * 60 * 60
	case "w", "week", "weeks":
		return amount * minReleaseAgeSeconds
	default:
		return 0
	}
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
