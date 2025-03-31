package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestEncryptingFS_Name(t *testing.T) {
	e := &EncryptingFS{}
	expectedName := "Filesystem encryption is enabled"
	if e.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, e.Name())
	}
}

func TestEncryptingFS_Status(t *testing.T) {
	e := &EncryptingFS{}
	expectedStatus := "Block device encryption is disabled"
	if e.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, e.Status())
	}
}

func TestEncryptingFS_UUID(t *testing.T) {
	e := &EncryptingFS{}
	expectedUUID := "21830a4e-84f1-48fe-9c5b-beab436b2cdb"
	if e.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, e.UUID())
	}
}

func TestEncryptingFS_Passed(t *testing.T) {
	e := &EncryptingFS{passed: true}
	expectedPassed := true
	if e.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, e.Passed())
	}
}

func TestEncryptingFS_FailedMessage(t *testing.T) {
	e := &EncryptingFS{}
	expectedFailedMessage := "Block device encryption is disabled"
	if e.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, e.FailedMessage())
	}
}

func TestEncryptingFS_PassedMessage(t *testing.T) {
	e := &EncryptingFS{}
	expectedPassedMessage := "Block device encryption is enabled"
	if e.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, e.PassedMessage())
	}
}
func TestMaybeCryptoViaLuks(t *testing.T) {
	tests := []struct {
		name     string
		mocks    []shared.RunCommandMock
		expected bool
	}{
		{
			name: "LUKS encryption detected",
			mocks: []shared.RunCommandMock{
				{
					Command: "lsblk",
					Args:    []string{"-o", "TYPE,MOUNTPOINT"},
					Out:     "TYPE MOUNTPOINT\npart /boot\ncrypt /\npart [SWAP]",
					Err:     nil,
				},
			},
			expected: true,
		},
		{
			name: "LUKS encryption detected but only for home",
			mocks: []shared.RunCommandMock{
				{
					Command: "lsblk",
					Args:    []string{"-o", "TYPE,MOUNTPOINT"},
					Out:     "TYPE MOUNTPOINT\npart /boot\ncrypt /home\npart [SWAP]",
					Err:     nil,
				},
			},
			expected: true,
		},
		{
			name: "LUKS encryption detected for LVM",
			mocks: []shared.RunCommandMock{
				{
					Command: "lsblk",
					Args:    []string{"-o", "TYPE,MOUNTPOINT"},
					Out:     "TYPE MOUNTPOINT\npart /boot\ncrypt \npart [SWAP]",
					Err:     nil,
				},
			},
			expected: true,
		},
		{
			name: "No LUKS encryption",
			mocks: []shared.RunCommandMock{
				{
					Command: "lsblk",
					Args:    []string{"-o", "TYPE,MOUNTPOINT"},
					Out:     "TYPE MOUNTPOINT\npart /boot\npart /\npart [SWAP]",
					Err:     nil,
				},
			},
			expected: false,
		},
		{
			name: "Command error",
			mocks: []shared.RunCommandMock{
				{
					Command: "lsblk",
					Args:    []string{"-o", "TYPE,MOUNTPOINT"},
					Out:     "",
					Err:     assert.AnError,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mocks
			result := maybeCryptoViaLuks()
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestMaybeCryptoViaKernel(t *testing.T) {
	// Save the original ReadFile function and restore it after the test

	tests := []struct {
		name        string
		cmdlineData string
		readFileErr error
		expected    bool
	}{
		{
			name:        "Kernel crypto parameters found",
			cmdlineData: "BOOT_IMAGE=/boot/vmlinuz-5.10.0-kali-amd64 root=/dev/mapper/vgkali-root cryptdevice=UUID=123:vg:root",
			readFileErr: nil,
			expected:    true,
		},
		{
			name:        "No kernel crypto parameters",
			cmdlineData: "BOOT_IMAGE=/boot/vmlinuz-5.10.0-kali-amd64 root=/dev/mapper/vgkali-root ro quiet",
			readFileErr: nil,
			expected:    false,
		},
		{
			name:        "Crypto parameters with wrong format",
			cmdlineData: "BOOT_IMAGE=/boot/vmlinuz cryptdevice=wrongformat",
			readFileErr: nil,
			expected:    false,
		},
		{
			name:        "Crypto parameters but not for root",
			cmdlineData: "BOOT_IMAGE=/boot/vmlinuz cryptdevice=UUID=123:vg:home",
			readFileErr: nil,
			expected:    false,
		},
		{
			name:        "Error reading cmdline file",
			cmdlineData: "",
			readFileErr: assert.AnError,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the ReadFile function
			shared.ReadFileMocks = map[string]string{
				"/proc/cmdline": tt.cmdlineData,
			}

			result := maybeCryptoViaKernel()
			assert.Equal(t, tt.expected, result)
		})
	}
}
