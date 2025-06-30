package checks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

type PasswordManagerCheck struct {
	passed bool
}

func (pmc *PasswordManagerCheck) Name() string {
	return "Password Manager Presence"
}

func (pmc *PasswordManagerCheck) Run() error {
	appNames := []string{
		"1Password.app",
		"1Password 8.app",
		"1Password 7.app",
		"Bitwarden.app",
		"Dashlane.app",
		"KeePassXC.app",
		"KeePassX.app",
	}

	if checkInstalledApplications(appNames) || checkForBrowserExtensions() {
		pmc.passed = true
	} else {
		pmc.passed = false
	}
	return nil
}

func checkInstalledApplications(appNames []string) bool {
	searchPaths := []string{
		"/Applications",
		"/System/Applications",
		filepath.Join(os.Getenv("HOME"), "Applications"),
	}

	for _, path := range searchPaths {
		if contents, err := os.ReadDir(path); err == nil {
			for _, entry := range contents {
				if entry.IsDir() && lo.Contains(appNames, entry.Name()) {
					return true
				}
			}
		}
	}
	return false
}

func checkForBrowserExtensions() bool {
	home := os.Getenv("HOME")
	extensionPaths := map[string]string{
		"Google Chrome":  filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "Extensions"),
		"Firefox":        filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles"),
		"Microsoft Edge": filepath.Join(home, "Library", "Application Support", "Microsoft Edge", "Default", "Extensions"),
		"Brave Browser":  filepath.Join(home, "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Extensions"),
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
