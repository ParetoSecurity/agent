//go:build !windows

package checks

import "fmt"

func defaultSystemDrive() string {
	return ""
}

func defaultListFixedDrives() ([]string, error) {
	return nil, nil
}

func defaultGetBitLockerProtectionStatus(path string) (int, error) {
	return 0, fmt.Errorf("bitlocker status is only supported on Windows")
}
