package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCAPEM = `-----BEGIN CERTIFICATE-----
MIIBuDCCAV6gAwIBAgIUdhEp4UC4uHGc9GaoqBxIlBC4tP4wCgYIKoZIzj0EAwIw
QTELMAkGA1UEBhMCVVMxFDASBgNVBAoMC1BhcmV0byBUZXN0MRwwGgYDVQQDDBNQ
YXJldG8gVGVzdCBSb290IENBMCAXDTI2MTAyMjE3MDYzNVoYDzIxMjYwOTI4MTcw
NjM1WjBBMQswCQYDVQQGEwJVUzEUMBIGA1UECgwLUGFyZXRvIFRlc3QxHDAaBgNV
BAMME1BhcmV0byBUZXN0IFJvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC
AARN9sJe77N/gX02ql8gi/DnW2omQPIOEw+xGuEQj/jzlNhV+J96ovuAH/Yn6vOL
Q4s6333sv6ydv+5cqyJ1yU+QozIwMDAdBgNVHQ4EFgQUs7DZAA2y0jbhXGEiuyot
C2grffAwDwYDVR0TAQH/BAUwAwEB/zAKBggqhkjOPQQDAgNIADBFAiEAiVTIHW74
TsJxzxD+MRWnmk552DVdhvAkf2l5uTKZnX4CIFooebv/nOaCoWsdTADoPfCif56f
nvbL9mzIKkoBytHk
-----END CERTIFICATE-----
`

const testCAFingerprint = "E1:30:6F:26:F8:48:C5:44:FC:5B:47:62:46:38:3D:33:17:2A:3F:ED:6F:21:17:53:9E:E4:AF:E4:29:E3:34:AA"

var testCABlocklist = map[string]string{testCAFingerprint: "Pareto Test Root CA"}

func writeBundle(t *testing.T, path, contents string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	return path
}

func TestCompromisedCA_Metadata(t *testing.T) {
	check := CompromisedCA{}

	assert.Equal(t, "No compromised CAs are trusted", check.Name())
	assert.Equal(t, "9af95e09-1aa3-464b-8c32-d760c21e7f51", check.UUID())
	assert.Equal(t, "No compromised CAs are trusted", check.PassedMessage())
	assert.Equal(t, "Compromised CAs are trusted", check.FailedMessage())
	assert.False(t, check.RequiresRoot())
}

func TestCompromisedCA_Fingerprint(t *testing.T) {
	found := compromisedCAsIn([]byte(testCAPEM), testCABlocklist)
	assert.Equal(t, []string{"Pareto Test Root CA"}, found)

	assert.Empty(t, compromisedCAsIn([]byte(testCAPEM), compromisedRootCAs))
}

func TestCompromisedCA_SkipsMalformedBlocks(t *testing.T) {
	contents := "not a pem\n" + testCAPEM + "-----BEGIN CERTIFICATE-----\n!!!\n-----END CERTIFICATE-----\n" +
		"-----BEGIN PRIVATE KEY-----\nAAAA\n-----END PRIVATE KEY-----\n"
	assert.Equal(t, []string{"Pareto Test Root CA"}, compromisedCAsIn([]byte(contents), testCABlocklist))
}

func TestCompromisedCA_RunFailsAndGroupsRootsPerBundle(t *testing.T) {
	dir := t.TempDir()
	bad := writeBundle(t, filepath.Join(dir, "bad.pem"), testCAPEM+testCAPEM)
	clean := writeBundle(t, filepath.Join(dir, "clean.pem"), "")

	check := CompromisedCA{BundlePaths: []string{bad, clean}, Blocklist: testCABlocklist, HomeDir: "/nonexistent"}
	require.NoError(t, check.Run())

	assert.False(t, check.Passed())
	assert.Equal(t, bad+": Pareto Test Root CA", check.Status())
}

func TestCompromisedCA_RunPassesWithCleanBundles(t *testing.T) {
	dir := t.TempDir()
	a := writeBundle(t, filepath.Join(dir, "a.pem"), testCAPEM)
	b := writeBundle(t, filepath.Join(dir, "b.pem"), testCAPEM)

	check := CompromisedCA{BundlePaths: []string{a, b}}
	require.NoError(t, check.Run())

	assert.True(t, check.Passed())
	assert.Equal(t, "Scanned 2 CA bundle(s)", check.Status())
}

func TestCompromisedCA_SymlinkedBundleIsScannedOnce(t *testing.T) {
	dir := t.TempDir()
	bundle := writeBundle(t, filepath.Join(dir, "cert.pem"), testCAPEM)
	link := filepath.Join(dir, "openssl-cert.pem")
	require.NoError(t, os.Symlink(bundle, link))

	check := CompromisedCA{BundlePaths: []string{bundle, link}, Blocklist: testCABlocklist, HomeDir: "/nonexistent"}
	assert.Equal(t, []string{bundle}, check.bundlePaths())

	require.NoError(t, check.Run())
	assert.Equal(t, bundle+": Pareto Test Root CA", check.Status())
}

func TestCompromisedCA_IsRunnable(t *testing.T) {
	dir := t.TempDir()

	check := CompromisedCA{BundlePaths: []string{filepath.Join(dir, "missing.pem")}}
	assert.False(t, check.IsRunnable())

	check = CompromisedCA{BundlePaths: []string{writeBundle(t, filepath.Join(dir, "cert.pem"), "")}}
	assert.True(t, check.IsRunnable())

	check = CompromisedCA{GOOS: "windows", HomeDir: dir}
	assert.False(t, check.IsRunnable())
}

func TestCompromisedCA_FindsPipxCertifiAndShortensHome(t *testing.T) {
	home := t.TempDir()
	bundle := writeBundle(t, filepath.Join(home,
		".local/share/pipx/venvs/tool/lib/python3.12/site-packages/certifi/cacert.pem"), testCAPEM)

	check := CompromisedCA{GOOS: "linux", HomeDir: home, Blocklist: testCABlocklist}
	assert.Contains(t, check.certifiBundlePaths(), bundle)

	require.NoError(t, check.Run())
	assert.False(t, check.Passed())
	assert.Contains(t, check.Status(),
		"~/.local/share/pipx/venvs/tool/lib/python3.12/site-packages/certifi/cacert.pem: Pareto Test Root CA")
}

func TestCompromisedCA_SkipsCertifiPatchedToUseSystemBundle(t *testing.T) {
	tests := []struct {
		name    string
		corepy  string
		skipped bool
	}{
		{
			name:    "Debian 12 and Ubuntu 24.04",
			corepy:  "DEBIAN_CA_CERTS_PATH = '/etc/ssl/certs/ca-certificates.crt'\n\ndef where() -> str:\n    return DEBIAN_CA_CERTS_PATH\n",
			skipped: true,
		},
		{
			name:    "Ubuntu 22.04",
			corepy:  "This module returns the installation location of cacert.pem or its contents.\nDEBIAN_CA_CERTS_PATH = '/etc/ssl/certs/ca-certificates.crt'\n",
			skipped: true,
		},
		{
			name:    "upstream certifi",
			corepy:  "def where() -> str:\n    return str(files(\"certifi\").joinpath(\"cacert.pem\"))\n",
			skipped: false,
		},
		{
			name:    "no core.py",
			skipped: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			certifi := filepath.Join(home, ".local/lib/python3.12/site-packages/certifi")
			bundle := writeBundle(t, filepath.Join(certifi, "cacert.pem"), testCAPEM)
			if tt.corepy != "" {
				writeBundle(t, filepath.Join(certifi, "core.py"), tt.corepy)
			}

			check := CompromisedCA{GOOS: "linux", HomeDir: home, Blocklist: testCABlocklist}
			require.NoError(t, check.Run())

			if tt.skipped {
				assert.NotContains(t, check.certifiBundlePaths(), bundle)
				assert.True(t, check.Passed())
			} else {
				assert.Contains(t, check.certifiBundlePaths(), bundle)
				assert.False(t, check.Passed())
			}
		})
	}
}

func TestCompromisedCA_CapsCertifiBundles(t *testing.T) {
	var matches []string
	for i := 0; i < maxCertifiBundles+10; i++ {
		matches = append(matches, filepath.Join("/venv", string(rune('a'+i%26)), "cacert.pem"))
	}
	check := CompromisedCA{
		GOOS: "linux",
		Glob: func(string) ([]string, error) { return matches, nil },
	}
	assert.Len(t, check.certifiBundlePaths(), maxCertifiBundles)
}

func TestCompromisedCA_SystemBundlePathsPerPlatform(t *testing.T) {
	linux := CompromisedCA{GOOS: "linux"}
	assert.Contains(t, linux.systemBundlePaths(), "/etc/ssl/certs/ca-certificates.crt")
	assert.Contains(t, linux.systemBundlePaths(), "/etc/pki/tls/certs/ca-bundle.crt")

	darwin := CompromisedCA{GOOS: "darwin"}
	assert.Contains(t, darwin.systemBundlePaths(), "/etc/ssl/cert.pem")
	assert.Contains(t, darwin.systemBundlePaths(), "/opt/homebrew/etc/ca-certificates/cert.pem")

	for _, path := range append(linux.systemBundlePaths(), darwin.systemBundlePaths()...) {
		assert.NotContains(t, path, "/nix/store")
	}
}
