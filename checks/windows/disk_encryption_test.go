package checks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskEncryption_Run(t *testing.T) {
	tests := []struct {
		name           string
		systemDrive    string
		fixedDrives    []string
		statusByDrive  map[string]int
		statusErr      error
		expectedPassed bool
		expectedStatus string
	}{
		{
			name:           "BitLocker enabled on OS volume",
			systemDrive:    "C:",
			statusByDrive:  map[string]int{"C:": 1},
			expectedPassed: true,
			expectedStatus: "BitLocker is enabled on C:",
		},
		{
			name:           "BitLocker disabled on OS volume",
			systemDrive:    "C:",
			statusByDrive:  map[string]int{"C:": 0},
			expectedPassed: false,
			expectedStatus: "BitLocker is not enabled on OS volume C: (status: 0 (Unknown Status))",
		},
		{
			name:           "Multiple volumes, OS volume encrypted",
			systemDrive:    "C:",
			statusByDrive:  map[string]int{"C:": 1, "D:": 0},
			expectedPassed: true,
			expectedStatus: "BitLocker is enabled on C:",
		},
		{
			name:           "Multiple volumes, OS volume not encrypted",
			systemDrive:    "C:",
			statusByDrive:  map[string]int{"C:": 0, "D:": 1},
			expectedPassed: false,
			expectedStatus: "BitLocker is not enabled on OS volume C: (status: 0 (Unknown Status))",
		},
		{
			name:           "No OS volume, but data volume encrypted",
			systemDrive:    "",
			fixedDrives:    []string{"D:"},
			statusByDrive:  map[string]int{"D:": 1},
			expectedPassed: true,
			expectedStatus: "BitLocker is enabled on D:",
		},
		{
			name:           "No volumes encrypted",
			systemDrive:    "",
			fixedDrives:    []string{"D:"},
			statusByDrive:  map[string]int{"D:": 0},
			expectedPassed: false,
			expectedStatus: "No encrypted volumes found (last checked D:: status 0 (Unknown Status))",
		},
		{
			name:           "Empty output",
			systemDrive:    "",
			fixedDrives:    []string{},
			expectedPassed: false,
			expectedStatus: "No BitLocker volumes found",
		},
		{
			name:           "Command error",
			systemDrive:    "C:",
			statusErr:      errors.New("boom"),
			expectedPassed: false,
			expectedStatus: "Failed to query BitLocker status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origSystemDrive := bitLockerSystemDrive
			origFixedDrives := bitLockerListFixedDrives
			origProtection := bitLockerProtectionStatusFn
			defer func() {
				bitLockerSystemDrive = origSystemDrive
				bitLockerListFixedDrives = origFixedDrives
				bitLockerProtectionStatusFn = origProtection
			}()

			bitLockerSystemDrive = func() string {
				return tt.systemDrive
			}
			bitLockerListFixedDrives = func() ([]string, error) {
				return tt.fixedDrives, nil
			}
			bitLockerProtectionStatusFn = func(path string) (int, error) {
				if tt.statusErr != nil {
					return 0, tt.statusErr
				}
				if status, ok := tt.statusByDrive[path]; ok {
					return status, nil
				}
				return 0, nil
			}

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
	assert.False(t, d.RequiresRoot())
}
