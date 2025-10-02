// Package linux provides checks for Linux systems.
package checks

import (
	"regexp"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/samber/lo"
)

// Autologin checks for autologin misconfiguration.
type Autologin struct {
	passed bool
	status string
}

// isGreeterCommand checks if a command is a greeter (login screen) rather than a desktop session
func isGreeterCommand(cmd string) bool {
	greeters := []string{
		"agreety",
		"cosmic-greeter",
		"ddlm",
		"emptty",
		"greetd-egui",
		"greetd-gtkgreet",
		"greetd-mini-greeter",
		"greetd-mini-wl-greeter",
		"greetly",
		"gtkgreet",
		"iced-greet",
		"ly",
		"marine_greetdm",
		"meow-greeter",
		"ncgreet",
		"nwg-hello",
		"nyow-greeter",
		"octobacillus",
		"qmlgreet",
		"qtgreet",
		"regreet",
		"salut",
		"tuigreet",
		"waygreet",
		"waycratedm",
		"wlgreet",
	}

	cmdLower := strings.ToLower(cmd)

	// Split by whitespace and check each part
	parts := strings.Fields(cmdLower)
	foundGreeter := lo.ContainsBy(parts, func(part string) bool {
		// Remove path if present
		// /usr/bin/tuigreet -> tuigreet
		if idx := strings.LastIndex(part, "/"); idx != -1 {
			part = part[idx+1:]
		}
		return lo.Contains(greeters, part)
	})

	if foundGreeter {
		return true
	}

	// Also check if it's starting sway/cage with a greeter config
	// (common pattern for gtkgreet/regreet)
	if strings.Contains(cmdLower, "sway") && strings.Contains(cmdLower, "greet") {
		return true
	}
	if strings.Contains(cmdLower, "cage") && strings.Contains(cmdLower, "greet") {
		return true
	}

	return false
}

// checkGreetdSessionAutologin checks if a greetd session has autologin configured
func checkGreetdSessionAutologin(configPath, sessionName string) bool {
	cmd, cmdExists := shared.GetTOMLSectionKey(configPath, sessionName, "command")
	if !cmdExists {
		return false
	}

	user, userExists := shared.GetTOMLSectionKey(configPath, sessionName, "user")

	// Autologin is configured if:
	// 1. Command is not a greeter AND
	// 2. User is specified and not "greeter"
	return !isGreeterCommand(cmd) && userExists && user != "" && user != "greeter"
}

// getGreetdConfigPaths finds greetd config paths
func getGreetdConfigPaths() []string {
	candidates := []string{
		"/etc/greetd/config.toml",
		"/etc/greetd/greetd.toml",
	}

	// Try to get config path from systemd service (for NixOS)
	if output, err := shared.RunCommand("systemctl", "cat", "greetd.service"); err == nil {
		// Look for ExecStart line with config path
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ExecStart=") {
				// Extract config path from command line args
				if strings.Contains(line, "--config") {
					parts := regexp.MustCompile(`--config\s+([^\s]+)`).FindStringSubmatch(line)
					if len(parts) > 1 {
						candidates = append([]string{parts[1]}, candidates...)
					}
				}
			}
		}
	}

	var existingPaths []string
	for _, path := range candidates {
		if _, err := osStat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	return existingPaths
}

// Name returns the name of the check
func (f *Autologin) Name() string {
	return "Automatic login is disabled"
}

// Run executes the check
func (f *Autologin) Run() error {
	f.passed = true

	// Check KDE (SDDM) autologin
	sddmFiles, _ := filepathGlob("/etc/sddm.conf.d/*.conf")
	sddmFiles = append(sddmFiles, "/etc/sddm.conf") // Add main config to list

	for _, file := range sddmFiles {
		content, err := shared.ReadFile(file)
		if err == nil {
			contentStr := string(content)

			// SDDM uses User= under [Autologin] section
			// Check for User= setting (any non-empty user means autologin is configured)
			if regexp.MustCompile(`(?m)^\s*User\s*=\s*\S+`).MatchString(contentStr) {
				f.passed = false
				f.status = "SDDM autologin user is configured"
				return nil
			}

			// Also check for Session= which might indicate autologin setup
			if regexp.MustCompile(`(?m)^\s*\[Autologin\]`).MatchString(contentStr) &&
				regexp.MustCompile(`(?m)^\s*Session\s*=\s*\S+`).MatchString(contentStr) {
				f.passed = false
				f.status = "SDDM autologin session is configured"
				return nil
			}
		}
	}

	// Check GNOME (GDM) autologin
	gdmPaths := []string{"/etc/gdm3/custom.conf", "/etc/gdm/custom.conf"}
	for _, path := range gdmPaths {
		if content, err := shared.ReadFile(path); err == nil {
			contentStr := string(content)

			if strings.Contains(contentStr, "AutomaticLoginEnable=true") {
				f.passed = false
				f.status = "AutomaticLoginEnable=true in GDM is enabled"
				return nil
			}

			// Check for NixOS-style timed login settings (any of these indicate autologin)
			if strings.Contains(contentStr, "TimedLoginEnable=true") {
				f.passed = false
				f.status = "TimedLoginEnable=true in GDM is enabled"
				return nil
			}

			// Check if TimedLogin user is set
			if regexp.MustCompile(`TimedLogin=\S+`).MatchString(contentStr) {
				f.passed = false
				f.status = "TimedLogin user is configured in GDM"
				return nil
			}

			// Check if TimedLoginDelay is set (any value, including 0)
			if regexp.MustCompile(`TimedLoginDelay=\d+`).MatchString(contentStr) {
				f.passed = false
				f.status = "TimedLoginDelay is configured in GDM"
				return nil
			}
		}
	}

	// Check GNOME (GDM) autologin using dconf
	output, err := shared.RunCommand("dconf", "read", "/org/gnome/login-screen/enable-automatic-login")
	if err == nil && strings.TrimSpace(string(output)) == "true" {
		f.passed = false
		f.status = "Automatic login is enabled in GNOME"
		return nil
	}

	// Check for NixOS getty autologin marker file
	if _, err := osStat("/run/agetty.autologged"); err == nil {
		f.passed = false
		f.status = "Getty autologin detected (NixOS /run/agetty.autologged exists)"
		return nil
	}

	// Check systemd getty service overrides for autologin
	gettyOverrides, _ := filepathGlob("/etc/systemd/system/getty@*.service.d/*.conf")
	gettyOverrides = append(gettyOverrides, "/etc/systemd/system/getty@.service.d/overrides.conf")
	gettyOverrides = append(gettyOverrides, "/etc/systemd/system/serial-getty@.service.d/overrides.conf")

	for _, file := range gettyOverrides {
		if content, err := shared.ReadFile(file); err == nil {
			if strings.Contains(string(content), "--autologin") {
				f.passed = false
				f.status = "Getty autologin detected in systemd service override"
				return nil
			}
		}
	}

	// Check LightDM autologin
	lightdmFiles, _ := filepathGlob("/etc/lightdm/lightdm.conf.d/*.conf")
	lightdmFiles = append(lightdmFiles, "/etc/lightdm/lightdm.conf")

	for _, file := range lightdmFiles {
		if content, err := shared.ReadFile(file); err == nil {
			contentStr := string(content)

			// Check for autologin-user setting (ignoring commented lines)
			// Uses (?m) for multiline mode where ^ matches start of line
			if regexp.MustCompile(`(?m)^\s*autologin-user\s*=\s*\S+`).MatchString(contentStr) {
				f.passed = false
				f.status = "LightDM autologin user is configured"
				return nil
			}

			// Also check for autologin-guest (guest session autologin)
			if regexp.MustCompile(`(?m)^\s*autologin-guest\s*=\s*true`).MatchString(contentStr) {
				f.passed = false
				f.status = "LightDM guest autologin is enabled"
				return nil
			}

			// Check for autologin-session
			if regexp.MustCompile(`(?m)^\s*autologin-session\s*=\s*\S+`).MatchString(contentStr) {
				f.passed = false
				f.status = "LightDM autologin session is configured"
				return nil
			}
		}
	}

	// Check greetd autologin
	greetdConfigPaths := getGreetdConfigPaths()
	for _, configPath := range greetdConfigPaths {
		// Check initial_session (autologin on first boot)
		if checkGreetdSessionAutologin(configPath, "initial_session") {
			f.passed = false
			f.status = "greetd initial_session autologin is configured"
			return nil
		}

		// Check default_session (autologin after logout)
		if checkGreetdSessionAutologin(configPath, "default_session") {
			f.passed = false
			f.status = "greetd default_session autologin is configured"
			return nil
		}
	}

	return nil
}

// Passed returns the status of the check
func (f *Autologin) Passed() bool {
	return f.passed
}

// IsRunnable returns whether Autologin is runnable.
func (f *Autologin) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *Autologin) UUID() string {
	return "f962c423-fdf5-428a-a57a-816abc9b253e"
}

// PassedMessage returns the message to return if the check passed
func (f *Autologin) PassedMessage() string {
	return "Automatic login is off"
}

// FailedMessage returns the message to return if the check failed
func (f *Autologin) FailedMessage() string {
	return "Automatic login is on"
}

// RequiresRoot returns whether the check requires root access
func (f *Autologin) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *Autologin) Status() string {
	if !f.Passed() {
		return f.status
	}
	return f.PassedMessage()
}
