package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestDiskEncryption_Run(t *testing.T) {
	psCmd := "Get-BitLockerVolume | Select-Object MountPoint, ProtectionStatus, VolumeType | ConvertTo-Json"

	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "BitLocker enabled on OS volume",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     `{"MountPoint":"C:","ProtectionStatus":1,"VolumeType":1}`,
				},
			},
			expectedPassed: true,
			expectedStatus: "BitLocker is enabled on C:",
		},
		{
			name: "BitLocker disabled on OS volume",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     `{"MountPoint":"C:","ProtectionStatus":0,"VolumeType":1}`,
				},
			},
			expectedPassed: false,
			expectedStatus: "BitLocker is not enabled on OS volume C:",
		},
		{
			name: "Multiple volumes, OS volume encrypted",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     `[{"MountPoint":"C:","ProtectionStatus":1,"VolumeType":1},{"MountPoint":"D:","ProtectionStatus":0,"VolumeType":0}]`,
				},
			},
			expectedPassed: true,
			expectedStatus: "BitLocker is enabled on C:",
		},
		{
			name: "Multiple volumes, OS volume not encrypted",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     `[{"MountPoint":"C:","ProtectionStatus":0,"VolumeType":1},{"MountPoint":"D:","ProtectionStatus":1,"VolumeType":0}]`,
				},
			},
			expectedPassed: false,
			expectedStatus: "BitLocker is not enabled on OS volume C:",
		},
		{
			name: "No OS volume, but data volume encrypted",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     `{"MountPoint":"D:","ProtectionStatus":1,"VolumeType":0}`,
				},
			},
			expectedPassed: true,
			expectedStatus: "BitLocker is enabled on D:",
		},
		{
			name: "No volumes encrypted",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     `{"MountPoint":"D:","ProtectionStatus":0,"VolumeType":0}`,
				},
			},
			expectedPassed: false,
			expectedStatus: "No encrypted volumes found",
		},
		{
			name: "Empty output",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     "",
				},
			},
			expectedPassed: false,
			expectedStatus: "No BitLocker volumes found",
		},
		{
			name: "Command error",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     "",
					Err:     assert.AnError,
				},
			},
			expectedPassed: false,
			expectedStatus: "Failed to query BitLocker status",
		},
		{
			name: "Invalid JSON",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", psCmd},
					Out:     "not json",
				},
			},
			expectedPassed: false,
			expectedStatus: "Failed to parse BitLocker status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mockCommands

			d := &DiskEncryption{}
			err := d.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, d.Passed())
			assert.Equal(t, tt.expectedStatus, d.Status())
		})
	}
}

func TestDiskEncryption_Metadata(t *testing.T) {
	d := &DiskEncryption{}
	assert.Equal(t, "Disk encryption is enabled", d.Name())
	assert.Equal(t, "Disk encryption is enabled", d.PassedMessage())
	assert.Equal(t, "Disk encryption is disabled", d.FailedMessage())
	assert.Equal(t, "c2cae85c-0335-4708-a428-3a16fd407912", d.UUID())
	assert.True(t, d.IsRunnable())
	assert.True(t, d.RequiresRoot())
}
