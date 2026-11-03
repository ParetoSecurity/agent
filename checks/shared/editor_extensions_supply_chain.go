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

const minExtensionCooldownHours = 7 * 24

type editorConfig struct {
	name string
	// settingsDir is the editor's folder under the platform config directory,
	// for example "Code" for ~/.config/Code/User/settings.json.
	settingsDir string
	// flatpakID is the Flathub application ID on Linux, if there is one.
	flatpakID string
	binaries  []string
	// installPaths lists app locations. If one exists, the editor is present
	// even when its CLI is not on PATH.
	installPaths func(*EditorExtensionsSupplyChain) []string
	setting      string
	// minimumVersion is set when only newer editor versions enforce the setting.
	minimumVersion string
}

// EditorExtensionsSupplyChain verifies code editors delay newly published extension releases.
type EditorExtensionsSupplyChain struct {
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
func (e *EditorExtensionsSupplyChain) Name() string {
	return "Editor extensions delay new releases"
}

// Run executes the check.
func (e *EditorExtensionsSupplyChain) Run() error {
	failures := e.validationFailures()
	if len(failures) > 0 {
		e.passed = false
		e.status = strings.Join(failures, "; ")
		return nil
	}

	e.passed = true
	e.status = strings.Join(e.passingDetails(), "; ")
	return nil
}

// Passed returns true if the check passed.
func (e *EditorExtensionsSupplyChain) Passed() bool {
	return e.passed
}

// IsRunnable returns true if a supported editor or its settings file exists.
func (e *EditorExtensionsSupplyChain) IsRunnable() bool {
	for _, editor := range e.editors() {
		if _, _, ok := e.activeSettings(editor); ok {
			return true
		}
		if e.isInstalled(editor) {
			return true
		}
	}
	return false
}

// UUID returns the UUID of the check.
func (e *EditorExtensionsSupplyChain) UUID() string {
	return "d53a7b3b-2484-4d2f-966e-0ef1e9b64e48"
}

// PassedMessage returns the message to return if the check passed.
func (e *EditorExtensionsSupplyChain) PassedMessage() string {
	return "Editor extension supply-chain protection is configured"
}

// FailedMessage returns the message to return if the check failed.
func (e *EditorExtensionsSupplyChain) FailedMessage() string {
	return "Editor extensions install new releases immediately"
}

// RequiresRoot returns true if the check needs root access.
func (e *EditorExtensionsSupplyChain) RequiresRoot() bool {
	return false
}

// Status returns details for the current check state.
func (e *EditorExtensionsSupplyChain) Status() string {
	if e.status != "" {
		return e.status
	}
	if !e.IsRunnable() {
		return "No supported editor or settings file found"
	}
	if e.Passed() {
		return e.PassedMessage()
	}
	return e.FailedMessage()
}

func (e *EditorExtensionsSupplyChain) validationFailures() []string {
	var failures []string

	for _, editor := range e.editors() {
		contents, path, ok := e.activeSettings(editor)
		if !ok {
			if e.isInstalled(editor) {
				failures = append(failures, e.displayPath(e.settingsPaths(editor)[0])+" is missing")
			}
			continue
		}
		editorFailures := validateEditorSettings(contents, e.displayPath(path), editor.setting)
		if len(editorFailures) == 0 && editor.minimumVersion != "" && jsoncValue(contents, editor.setting) != "" {
			editorFailures = e.versionGateFailures(editor)
		}
		failures = append(failures, editorFailures...)
	}

	return failures
}

func (e *EditorExtensionsSupplyChain) passingDetails() []string {
	var details []string

	for _, editor := range e.editors() {
		contents, path, ok := e.activeSettings(editor)
		if !ok || len(validateEditorSettings(contents, path, editor.setting)) > 0 {
			continue
		}
		if isExtensionAutoUpdateDisabled(contents) {
			details = append(details, e.displayPath(path)+" has automatic extension updates disabled")
		} else {
			details = append(details, e.displayPath(path)+" delays "+editor.name+" extension updates")
		}
	}
	if len(details) == 0 {
		return []string{e.PassedMessage()}
	}
	return details
}

func (e *EditorExtensionsSupplyChain) versionGateFailures(editor editorConfig) []string {
	binary := e.firstInstalledBinary(editor.binaries)
	if binary == "" {
		return nil
	}
	version := e.editorVersion(binary)
	if version == "" {
		return []string{editor.name + " version could not be determined, so " + editor.setting + " could not be verified"}
	}
	if !versionAtLeast(version, editor.minimumVersion) {
		return []string{editor.name + " is older than " + editor.minimumVersion + " and does not enforce " + editor.setting}
	}
	return nil
}

func (e *EditorExtensionsSupplyChain) editors() []editorConfig {
	return []editorConfig{
		// Visual Studio Code family
		{
			name:           "Visual Studio Code",
			settingsDir:    "Code",
			flatpakID:      "com.visualstudio.code",
			binaries:       []string{"code"},
			installPaths:   vscodeInstallPaths("Visual Studio Code.app", "Microsoft VS Code"),
			setting:        "extensions.autoUpdateDelay",
			minimumVersion: "1.125.0",
		},
		{
			name:           "Visual Studio Code - Insiders",
			settingsDir:    "Code - Insiders",
			binaries:       []string{"code-insiders"},
			installPaths:   vscodeInstallPaths("Visual Studio Code - Insiders.app", "Microsoft VS Code Insiders"),
			setting:        "extensions.autoUpdateDelay",
			minimumVersion: "1.125.0",
		},
		{
			name:           "VSCodium",
			settingsDir:    "VSCodium",
			flatpakID:      "com.vscodium.codium",
			binaries:       []string{"codium"},
			installPaths:   vscodeInstallPaths("VSCodium.app", "VSCodium"),
			setting:        "extensions.autoUpdateDelay",
			minimumVersion: "1.125.0",
		},

		// Other editors
		{
			name:         "Cursor",
			settingsDir:  "Cursor",
			binaries:     []string{"cursor"},
			installPaths: vscodeInstallPaths("Cursor.app", "cursor"),
			setting:      "extensions.installCooldownHours",
		},
	}
}

// vscodeInstallPaths lists where an editor based on VS Code installs. That is
// the app bundle on macOS and the per-user folder on Windows.
func vscodeInstallPaths(macBundle string, windowsDir string) func(*EditorExtensionsSupplyChain) []string {
	return func(e *EditorExtensionsSupplyChain) []string {
		switch e.goos() {
		case "darwin":
			return []string{
				filepath.Join("/Applications", macBundle),
				filepath.Join(e.homeDir(), "Applications", macBundle),
			}
		case "windows":
			return []string{filepath.Join(e.localAppDataDir(), "Programs", windowsDir)}
		default:
			return nil
		}
	}
}

// settingsPaths lists the editor's user settings.json locations. The first
// match wins. On Linux the Flatpak build keeps its config under ~/.var/app.
func (e *EditorExtensionsSupplyChain) settingsPaths(editor editorConfig) []string {
	settings := filepath.Join(editor.settingsDir, "User", "settings.json")
	switch e.goos() {
	case "darwin":
		return []string{filepath.Join(e.homeDir(), "Library", "Application Support", settings)}
	case "windows":
		return []string{filepath.Join(e.appDataDir(), settings)}
	default:
		paths := []string{filepath.Join(e.configHome(), settings)}
		if editor.flatpakID != "" {
			paths = append(paths, filepath.Join(e.homeDir(), ".var", "app", editor.flatpakID, "config", settings))
		}
		return paths
	}
}

func (e *EditorExtensionsSupplyChain) isInstalled(editor editorConfig) bool {
	if e.firstInstalledBinary(editor.binaries) != "" {
		return true
	}
	for _, path := range editor.installPaths(e) {
		if e.fileExists(path) {
			return true
		}
	}
	if editor.flatpakID != "" && e.goos() == "linux" {
		return e.fileExists(filepath.Join(e.homeDir(), ".var", "app", editor.flatpakID))
	}
	return false
}

func (e *EditorExtensionsSupplyChain) activeSettings(editor editorConfig) (string, string, bool) {
	for _, path := range e.settingsPaths(editor) {
		if contents, err := e.readFile(path); err == nil {
			return string(contents), path, true
		}
	}
	return "", "", false
}

func (e *EditorExtensionsSupplyChain) firstInstalledBinary(names []string) string {
	for _, name := range names {
		if _, err := e.lookPath(name); err == nil {
			return name
		}
	}
	return ""
}

// editorVersion reads the version from `<cli> --version`. The first line of
// that output is the version number, for example "1.125.0".
func (e *EditorExtensionsSupplyChain) editorVersion(name string) string {
	if version, ok := e.Versions[name]; ok {
		return versionToken(version)
	}
	path, err := e.lookPath(name)
	if err != nil {
		return ""
	}
	output, err := e.runCommand(path, "--version")
	if err != nil {
		return ""
	}
	return versionToken(string(output))
}

// displayPath abbreviates paths under the home directory to ~/...
func (e *EditorExtensionsSupplyChain) displayPath(path string) string {
	home := e.homeDir()
	if relative, err := filepath.Rel(home, path); home != "" && err == nil && !strings.HasPrefix(relative, "..") {
		return "~/" + filepath.ToSlash(relative)
	}
	return path
}

func (e *EditorExtensionsSupplyChain) readFile(path string) ([]byte, error) {
	if e.ReadFile != nil {
		return e.ReadFile(path)
	}
	return os.ReadFile(path)
}

func (e *EditorExtensionsSupplyChain) fileExists(path string) bool {
	if e.FileExists != nil {
		return e.FileExists(path)
	}
	_, err := os.Stat(path)
	return err == nil
}

func (e *EditorExtensionsSupplyChain) lookPath(name string) (string, error) {
	if e.LookPath != nil {
		return e.LookPath(name)
	}
	return exec.LookPath(name)
}

func (e *EditorExtensionsSupplyChain) runCommand(name string, args ...string) ([]byte, error) {
	if e.RunCommand != nil {
		return e.RunCommand(name, args...)
	}
	return exec.Command(name, args...).Output()
}

func (e *EditorExtensionsSupplyChain) homeDir() string {
	if e.HomeDir != "" {
		return e.HomeDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func (e *EditorExtensionsSupplyChain) configHome() string {
	if configHome := e.env("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	return filepath.Join(e.homeDir(), ".config")
}

func (e *EditorExtensionsSupplyChain) appDataDir() string {
	if e.AppDataDir != "" {
		return e.AppDataDir
	}
	if appData := e.env("APPDATA"); appData != "" {
		return appData
	}
	return filepath.Join(e.homeDir(), "AppData", "Roaming")
}

func (e *EditorExtensionsSupplyChain) localAppDataDir() string {
	if localAppData := e.env("LOCALAPPDATA"); localAppData != "" {
		return localAppData
	}
	return filepath.Join(e.homeDir(), "AppData", "Local")
}

func (e *EditorExtensionsSupplyChain) env(name string) string {
	if e.Getenv != nil {
		return e.Getenv(name)
	}
	return os.Getenv(name)
}

func (e *EditorExtensionsSupplyChain) goos() string {
	if e.GOOS != "" {
		return e.GOOS
	}
	return runtime.GOOS
}

// validateEditorSettings passes when automatic extension updates are off, or
// when the cooldown setting holds releases back for at least 7 days.
func validateEditorSettings(contents string, path string, setting string) []string {
	if isExtensionAutoUpdateDisabled(contents) {
		return nil
	}
	hours, err := strconv.Atoi(jsoncValue(contents, setting))
	if err != nil {
		return []string{path + " " + setting + " is not set; set it to at least " + strconv.Itoa(minExtensionCooldownHours) + " hours"}
	}
	if hours < minExtensionCooldownHours {
		return []string{path + " " + setting + " is below " + strconv.Itoa(minExtensionCooldownHours) + " hours; set it to at least " + strconv.Itoa(minExtensionCooldownHours)}
	}
	return nil
}

// isExtensionAutoUpdateDisabled accepts the legacy boolean false and the
// current "off" value. If the key is not there, auto-update is on.
func isExtensionAutoUpdateDisabled(contents string) bool {
	value := jsoncValue(contents, "extensions.autoUpdate")
	return value == "false" || value == "off"
}

// jsoncValue reads a top-level "key": value token from settings.json. That
// file is JSONC, so it can contain comments and trailing commas.
func jsoncValue(contents string, key string) string {
	pattern := regexp.MustCompile(`"` + regexp.QuoteMeta(key) + `"\s*:\s*([^\n,}]+)`)
	match := pattern.FindStringSubmatch(stripJSONCComments(contents))
	if match == nil {
		return ""
	}
	return unquote(strings.TrimSpace(match[1]))
}

// stripJSONCComments removes // line comments and /* */ block comments. It
// keeps the contents of strings unchanged.
func stripJSONCComments(contents string) string {
	var result strings.Builder
	inString := false

	for index := 0; index < len(contents); index++ {
		character := contents[index]
		switch {
		case inString:
			result.WriteByte(character)
			if character == '\\' && index+1 < len(contents) {
				index++
				result.WriteByte(contents[index])
			} else if character == '"' {
				inString = false
			}
		case character == '"':
			inString = true
			result.WriteByte(character)
		case character == '/' && index+1 < len(contents) && contents[index+1] == '/':
			for index < len(contents) && contents[index] != '\n' {
				index++
			}
			if index < len(contents) {
				result.WriteByte('\n')
			}
		case character == '/' && index+1 < len(contents) && contents[index+1] == '*':
			end := strings.Index(contents[index+2:], "*/")
			if end < 0 {
				return result.String()
			}
			index += end + 3
		default:
			result.WriteByte(character)
		}
	}

	return result.String()
}
