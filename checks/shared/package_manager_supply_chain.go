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
	path       string
	binaries   []string
	validate   func(string) []string
	passDetail string
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
		if contents, ok := p.readConfig(config.path); ok {
			failures = append(failures, config.validate(contents)...)
			continue
		}
		if p.anyBinaryInstalled(config.binaries...) {
			failures = append(failures, config.path+" is missing")
		}
	}

	return failures
}

func (p *PackageManagerSupplyChain) passingDetails() []string {
	var details []string

	for _, config := range p.configs() {
		if contents, ok := p.readConfig(config.path); ok && len(config.validate(contents)) == 0 {
			details = append(details, config.passDetail)
		}
	}
	if len(details) == 0 {
		return []string{p.PassedMessage()}
	}
	return details
}

func (p *PackageManagerSupplyChain) hasConfigFile() bool {
	for _, config := range p.configs() {
		if p.fileExists(config.path) {
			return true
		}
	}
	return false
}

func (p *PackageManagerSupplyChain) hasPackageManagerBinary() bool {
	return p.anyBinaryInstalled("npm", "yarn", "bun", "uv")
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
			path:       filepath.Join(home, ".npmrc"),
			binaries:   []string{"npm"},
			validate:   validateNpmrc,
			passDetail: "~/.npmrc delays npm-compatible package releases and pins exact versions",
		},
		{
			path:       filepath.Join(home, ".yarnrc.yml"),
			binaries:   []string{"yarn"},
			validate:   validateYarnrc,
			passDetail: "~/.yarnrc.yml delays Yarn package releases",
		},
		{
			path:       filepath.Join(home, ".bunfig.toml"),
			binaries:   []string{"bun"},
			validate:   validateBunfig,
			passDetail: "~/.bunfig.toml delays Bun package releases",
		},
		{
			path:       p.uvConfigPath(),
			binaries:   []string{"uv"},
			validate:   validateUv,
			passDetail: p.uvConfigPath() + " excludes Python packages newer than 7 days",
		},
		{
			path:       filepath.Join(home, ".pypirc"),
			validate:   validatePypirc,
			passDetail: "~/.pypirc has no plaintext credentials",
		},
	}
}

func (p *PackageManagerSupplyChain) readConfig(path string) (string, bool) {
	contents, err := p.readFile(path)
	if err != nil {
		return "", false
	}
	return string(contents), true
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

	if integerValue(values["min-release-age"]) < 7 {
		failures = append(failures, "~/.npmrc min-release-age is below 7 days")
	}
	if integerValue(values["minimum-release-age"]) < minReleaseAgeMinutes {
		failures = append(failures, "~/.npmrc minimum-release-age is below 10080 minutes")
	}
	if strings.ToLower(values["save-exact"]) != "true" {
		failures = append(failures, "~/.npmrc save-exact is not enabled")
	}

	return failures
}

func validateYarnrc(contents string) []string {
	values := keyValuePairs(contents)
	if integerValue(values["npmminimalagegate"]) < minReleaseAgeMinutes {
		return []string{"~/.yarnrc.yml npmMinimalAgeGate is below 10080 minutes"}
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
	if durationSeconds(values["pip.exclude-newer"]) < minReleaseAgeSeconds {
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

	switch strings.TrimPrefix(normalized, number) {
	case "s":
		return amount
	case "m":
		return amount * 60
	case "h":
		return amount * 60 * 60
	case "d":
		return amount * 24 * 60 * 60
	case "w":
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
