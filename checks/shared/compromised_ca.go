package shared

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
)

// compromisedRootCAs maps certificate fingerprints to display names.
// A fingerprint is the SHA-256 of the DER certificate in uppercase hex with colons.
// This is the format of `openssl x509 -fingerprint -sha256`.
// The entries are the roots above the eMudhra intermediates with leaked keys (October 2026).
// See https://emudhra.we-have-a-problem.com
var compromisedRootCAs = map[string]string{
	"86:A1:EC:BA:08:9C:4A:8D:3B:BE:27:34:C6:12:BA:34:1D:81:3E:04:3C:F9:E8:A8:62:CD:5C:57:A3:6B:BE:6B": "emSign ECC Root CA - G3 (eMudhra)",
	"D7:A7:A0:FB:5D:7E:27:31:D7:71:E9:48:4E:BC:DE:F7:1D:5F:0C:3E:0A:29:48:78:2B:C8:3E:E0:EA:69:9E:F4": "AAA Certificate Services (Sectigo)",
}

// System CA bundles. The Nix entries cover the current system and the default profile.
// The check does not scan all store paths.
var linuxCABundlePaths = []string{
	"/etc/ssl/certs/ca-certificates.crt",
	"/etc/pki/tls/certs/ca-bundle.crt",
	"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem",
	"/etc/ssl/ca-bundle.pem",
	"/var/lib/ca-certificates/ca-bundle.pem",
	"/etc/ssl/cert.pem",
	"/nix/var/nix/profiles/default/etc/ssl/certs/ca-bundle.crt",
}

var darwinCABundlePaths = []string{
	"/etc/ssl/cert.pem",
	"/opt/homebrew/etc/ca-certificates/cert.pem",
	"/usr/local/etc/ca-certificates/cert.pem",
	"/opt/homebrew/etc/openssl@3/cert.pem",
	"/usr/local/etc/openssl@3/cert.pem",
	"/etc/ssl/certs/ca-certificates.crt",
	"/nix/var/nix/profiles/default/etc/ssl/certs/ca-bundle.crt",
}

const maxCertifiBundles = 30

// CompromisedCA makes sure that local CA bundles do not trust the roots of the
// compromised eMudhra intermediates.
type CompromisedCA struct {
	passed    bool
	status    string
	HomeDir   string
	GOOS      string
	Blocklist map[string]string
	// BundlePaths replaces bundle discovery when it is set.
	BundlePaths []string
	ReadFile    func(string) ([]byte, error)
	Glob        func(string) ([]string, error)
}

// Name returns the name of the check.
func (c *CompromisedCA) Name() string {
	return "No compromised CAs are trusted"
}

// Run executes the check.
func (c *CompromisedCA) Run() error {
	var failures []string
	scanned := 0
	for _, path := range c.bundlePaths() {
		contents, err := c.readFile(path)
		if err != nil {
			continue
		}
		scanned++
		if found := compromisedCAsIn(contents, c.blocklist()); len(found) > 0 {
			failures = append(failures, fmt.Sprintf("%s: %s", c.displayPath(path), strings.Join(found, ", ")))
		}
	}

	c.passed = len(failures) == 0
	if c.passed {
		c.status = fmt.Sprintf("Scanned %d CA bundle(s)", scanned)
	} else {
		c.status = strings.Join(failures, "; ")
	}
	return nil
}

// Passed returns true if the check passed.
func (c *CompromisedCA) Passed() bool {
	return c.passed
}

// IsRunnable returns true if this system has a CA bundle.
func (c *CompromisedCA) IsRunnable() bool {
	return len(c.bundlePaths()) > 0
}

// UUID returns the UUID of the check.
func (c *CompromisedCA) UUID() string {
	return "9af95e09-1aa3-464b-8c32-d760c21e7f51"
}

// PassedMessage returns the message to return if the check passed.
func (c *CompromisedCA) PassedMessage() string {
	return "No compromised CAs are trusted"
}

// FailedMessage returns the message to return if the check failed.
func (c *CompromisedCA) FailedMessage() string {
	return "Compromised CAs are trusted"
}

// RequiresRoot returns true if the check needs root access.
func (c *CompromisedCA) RequiresRoot() bool {
	return false
}

// Status returns details for the current check state.
func (c *CompromisedCA) Status() string {
	if c.status != "" {
		return c.status
	}
	if c.Passed() {
		return c.PassedMessage()
	}
	return c.FailedMessage()
}

// compromisedCAsIn returns the names of the blocklisted certificates in a PEM bundle.
// The names are unique and sorted.
func compromisedCAsIn(contents []byte, blocklist map[string]string) []string {
	seen := map[string]bool{}
	var found []string
	for {
		var block *pem.Block
		block, contents = pem.Decode(contents)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		if name, ok := blocklist[certificateFingerprint(block.Bytes)]; ok && !seen[name] {
			seen[name] = true
			found = append(found, name)
		}
	}
	sort.Strings(found)
	return found
}

func certificateFingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}

// bundlePaths returns the bundles that exist.
// It keeps only the first path for each file, so the check shows a symlinked bundle one time.
func (c *CompromisedCA) bundlePaths() []string {
	candidates := c.BundlePaths
	if candidates == nil {
		candidates = slices.Concat(c.systemBundlePaths(), c.certifiBundlePaths())
	}

	seen := map[string]bool{}
	var paths []string
	for _, path := range candidates {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		if seen[resolved] {
			continue
		}
		seen[resolved] = true
		paths = append(paths, path)
	}
	return paths
}

func (c *CompromisedCA) systemBundlePaths() []string {
	switch c.goos() {
	case "linux":
		return linuxCABundlePaths
	case "darwin":
		return darwinCABundlePaths
	}
	return nil
}

func (c *CompromisedCA) certifiBundlePaths() []string {
	home := c.homeDir()
	var patterns []string
	switch c.goos() {
	case "linux":
		patterns = []string{
			"/usr/lib/python3*/site-packages/certifi/cacert.pem",
			"/usr/lib/python3/dist-packages/certifi/cacert.pem",
			"/usr/local/lib/python3*/site-packages/certifi/cacert.pem",
			"/usr/local/lib/python3*/dist-packages/certifi/cacert.pem",
			filepath.Join(home, ".local/lib/python3*/site-packages/certifi/cacert.pem"),
			filepath.Join(home, ".local/share/pipx/venvs/*/lib/python3*/site-packages/certifi/cacert.pem"),
		}
	case "darwin":
		patterns = []string{
			"/opt/homebrew/lib/python3*/site-packages/certifi/cacert.pem",
			"/usr/local/lib/python3*/site-packages/certifi/cacert.pem",
			"/Library/Frameworks/Python.framework/Versions/3.*/lib/python3*/site-packages/certifi/cacert.pem",
			filepath.Join(home, "Library/Python/3.*/lib/python/site-packages/certifi/cacert.pem"),
		}
	default:
		return nil
	}
	patterns = append(patterns,
		filepath.Join(home, ".local/pipx/venvs/*/lib/python3*/site-packages/certifi/cacert.pem"),
		filepath.Join(home, "anaconda3/lib/python3*/site-packages/certifi/cacert.pem"),
		filepath.Join(home, "miniconda3/lib/python3*/site-packages/certifi/cacert.pem"),
	)

	var paths []string
	for _, pattern := range patterns {
		matches, err := c.glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			if len(paths) >= maxCertifiBundles {
				return paths
			}
			if c.certifiUsesSystemBundle(match) {
				continue
			}
			paths = append(paths, match)
		}
	}
	return paths
}

// certifiUsesSystemBundle reports whether the certifi package of a cacert.pem
// was patched to load a system bundle, so it never reads that cacert.pem.
// Debian 12, Ubuntu 22.04 and Ubuntu 24.04 ship such a stale cacert.pem.
// The system bundle is scanned separately.
func (c *CompromisedCA) certifiUsesSystemBundle(cacert string) bool {
	core, err := c.readFile(filepath.Join(filepath.Dir(cacert), "core.py"))
	if err != nil {
		return false
	}
	for _, bundle := range c.systemBundlePaths() {
		if strings.Contains(string(core), bundle) {
			return true
		}
	}
	return false
}

func (c *CompromisedCA) displayPath(path string) string {
	home := c.homeDir()
	if home != "" && (path == home || strings.HasPrefix(path, home+string(filepath.Separator))) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func (c *CompromisedCA) blocklist() map[string]string {
	if c.Blocklist != nil {
		return c.Blocklist
	}
	return compromisedRootCAs
}

func (c *CompromisedCA) homeDir() string {
	if c.HomeDir != "" {
		return c.HomeDir
	}
	home, _ := os.UserHomeDir()
	return home
}

func (c *CompromisedCA) goos() string {
	if c.GOOS != "" {
		return c.GOOS
	}
	return runtime.GOOS
}

func (c *CompromisedCA) readFile(path string) ([]byte, error) {
	if c.ReadFile != nil {
		return c.ReadFile(path)
	}
	return os.ReadFile(path)
}

func (c *CompromisedCA) glob(pattern string) ([]string, error) {
	if c.Glob != nil {
		return c.Glob(pattern)
	}
	return filepath.Glob(pattern)
}
