package checks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/log"
)

type PasswordManagerCheck struct {
	passed bool
}

func (pmc *PasswordManagerCheck) Name() string {
	return "Password Manager Presence"
}

func (pmc *PasswordManagerCheck) Run() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Debug("Failed to get user home directory, falling back to USERPROFILE")
		userHome = os.Getenv("USERPROFILE")
	}

	paths := []string{
		filepath.Join(userHome, "AppData", "Local", "1Password", "app", "8", "1Password.exe"),
		filepath.Join(userHome, "AppData", "Local", "Programs", "Bitwarden", "Bitwarden.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES"), "KeePass Password Safe 2", "KeePass.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "KeePass Password Safe 2", "KeePass.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES"), "KeePassXC", "KeePassXC.exe"),
		filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "KeePassXC", "KeePassXC.exe"),
	}

	log.WithField("paths", paths).Debug("Checking for installed password managers")

	for _, path := range paths {
		log.WithField("path", path).Debug("Checking password manager path")
		if _, err := osStat(path); err == nil {
			log.WithField("path", path).Debug("Found password manager")
			pmc.passed = true
			return nil
		}
	}

	log.Debug("No installed password managers found, checking browser extensions")
	pmc.passed = checkForBrowserExtensions()
	return nil
}

func checkForBrowserExtensions() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Debug("Failed to get user home directory for browser extensions, falling back to USERPROFILE")
		home = os.Getenv("USERPROFILE")
	}

	extensionPaths := map[string]string{
		"Google Chrome":  filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Extensions"),
		"Microsoft Edge": filepath.Join(home, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Extensions"),
		"Edge Beta":      filepath.Join(home, "AppData", "Local", "Microsoft", "Edge Beta", "User Data", "Default", "Extensions"),
		"Edge Dev":       filepath.Join(home, "AppData", "Local", "Microsoft", "Edge Dev", "User Data", "Default", "Extensions"),
		"Brave Browser":  filepath.Join(home, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions"),
		"Opera":          filepath.Join(home, "AppData", "Roaming", "Opera Software", "Opera Stable", "Extensions"),
		"Opera GX":       filepath.Join(home, "AppData", "Roaming", "Opera Software", "Opera GX Stable", "Extensions"),
		"Vivaldi":        filepath.Join(home, "AppData", "Local", "Vivaldi", "User Data", "Default", "Extensions"),
	}

	browserExtensions := []string{
		"hdokiejnpimakedhajhdlcegeplioahd", // LastPass
		"ghmbeldphafepmbegfdlkpapadhbakde", // ProtonPass
		"eiaeiblijfjekdanodkjadfinkhbfgcd", // nordpass
		"nngceckbapebfimnlniiiahkandclbl",  // bitwarden
		"aeblfdkhhhdcdjpifhhbdiojplfjncoa", // 1password
		"fdjamakpfbbddfjaooikfcpapjohcfmg", // dashlane
		"dppgmdbiimibapkepcbdbmkaabgiofem", // 1Password on Edge
		"gehmmocbbkpblljhkekmfhjpfbkclbph", // Dashlane on Edge
		"jbkfoedolllekgbhcbcoahefnbanhhlh", // Bitwarden on Edge
		"bbcinlkgjjkejfdpemiealijmmooekmp", // LastPass on Edge
	}

	// Check Chromium-based browsers
	for browser, extPath := range extensionPaths {
		log.WithField("browser", browser).WithField("path", extPath).Debug("Checking browser extensions path")
		if _, err := os.Stat(extPath); err == nil {
			entries, err := os.ReadDir(extPath)
			if err == nil {
				log.WithField("browser", browser).WithField("extensionCount", len(entries)).Debug("Found browser extensions directory")
				for _, entry := range entries {
					name := strings.ToLower(entry.Name())
					for _, ext := range browserExtensions {
						if strings.Contains(name, strings.ToLower(ext)) {
							log.WithField("browser", browser).WithField("extension", ext).Debug("Found password manager browser extension")
							return true
						}
					}
				}
			} else {
				log.WithField("browser", browser).WithError(err).Debug("Failed to read browser extensions directory")
			}
		}
	}

	// Check Firefox separately due to different extension structure
	if checkFirefoxExtensions(home) {
		return true
	}

	return false
}

func checkFirefoxExtensions(home string) bool {
	profilesPath := filepath.Join(home, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles")

	log.WithField("path", profilesPath).Debug("Checking Firefox profiles path")
	if _, err := os.Stat(profilesPath); err != nil {
		log.WithError(err).Debug("Firefox profiles directory not found")
		return false
	}

	profiles, err := os.ReadDir(profilesPath)
	if err != nil {
		log.WithError(err).Debug("Failed to read Firefox profiles directory")
		return false
	}

	log.WithField("profileCount", len(profiles)).Debug("Found Firefox profiles")

	// Firefox addon IDs for password managers
	firefoxAddonIDs := []string{
		"@lastpass-password-manager",             // LastPass
		"@proton-pass",                           // ProtonPass
		"nordpass@nordpass.com",                  // NordPass
		"{446900e4-71c2-419f-a6a7-df9c091e268b}", // Bitwarden
		"{d634138d-c276-4fc8-924b-40a0ea21d284}", // 1Password
		"extension@dashlane.com",                 // Dashlane
	}

	for _, profile := range profiles {
		if profile.IsDir() {
			extensionsPath := filepath.Join(profilesPath, profile.Name(), "extensions")
			log.WithField("profile", profile.Name()).WithField("path", extensionsPath).Debug("Checking Firefox profile extensions")
			if _, err := os.Stat(extensionsPath); err == nil {
				extensions, err := os.ReadDir(extensionsPath)
				if err == nil {
					log.WithField("profile", profile.Name()).WithField("extensionCount", len(extensions)).Debug("Found Firefox extensions directory")
					for _, ext := range extensions {
						extName := strings.ToLower(ext.Name())
						for _, addonID := range firefoxAddonIDs {
							if strings.Contains(extName, strings.ToLower(addonID)) {
								log.WithField("profile", profile.Name()).WithField("addon", addonID).Debug("Found password manager Firefox extension")
								return true
							}
						}
					}
				} else {
					log.WithField("profile", profile.Name()).WithError(err).Debug("Failed to read Firefox extensions directory")
				}
			}
		}
	}

	return false
}

func (pmc *PasswordManagerCheck) Passed() bool {
	return pmc.passed
}

func (pmc *PasswordManagerCheck) IsRunnable() bool {
	return true
}

func (pmc *PasswordManagerCheck) UUID() string {
	return "f962c423-fdf5-428a-a57a-827abc9b253e"
}

func (pmc *PasswordManagerCheck) PassedMessage() string {
	return "Password manager is present"
}

func (pmc *PasswordManagerCheck) FailedMessage() string {
	return "No password manager found"
}

func (pmc *PasswordManagerCheck) RequiresRoot() bool {
	return false
}

func (pmc *PasswordManagerCheck) Status() string {
	if pmc.Passed() {
		return pmc.PassedMessage()
	}
	return pmc.FailedMessage()
}
