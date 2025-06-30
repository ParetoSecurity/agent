package checks

import (
	"os"
	"path/filepath"
	"strings"
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

	for _, path := range paths {
		if _, err := osStat(path); err == nil {
			pmc.passed = true
			return nil
		}
	}

	pmc.passed = checkForBrowserExtensions()
	return nil
}

func checkForBrowserExtensions() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("USERPROFILE")
	}

	extensionPaths := map[string]string{
		"Google Chrome":  filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Extensions"),
		"Firefox":        filepath.Join(home, "AppData", "Roaming", "Mozilla", "Firefox", "Profiles"),
		"Microsoft Edge": filepath.Join(home, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Extensions"),
		"Brave Browser":  filepath.Join(home, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions"),
	}

	browserExtensions := []string{
		"hdokiejnpimakedhajhdlcegeplioahd", // LastPass
		"ghmbeldphafepmbegfdlkpapadhbakde", // ProtonPass
		"eiaeiblijfjekdanodkjadfinkhbfgcd", // nordpass
		"nngceckbapebfimnlniiiahkandclbl",  // bitwarden
		"aeblfdkhhhdcdjpifhhbdiojplfjncoa", // 1password
		"fdjamakpfbbddfjaooikfcpapjohcfmg", // dashlane
	}

	for _, extPath := range extensionPaths {
		if _, err := os.Stat(extPath); err == nil {
			entries, err := os.ReadDir(extPath)
			if err == nil {
				for _, entry := range entries {
					name := strings.ToLower(entry.Name())
					for _, ext := range browserExtensions {
						if strings.Contains(name, strings.ToLower(ext)) {
							return true
						}
					}
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
