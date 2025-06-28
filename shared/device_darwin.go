//go:build darwin
// +build darwin

package shared

import "strings"

type ReportingDevice struct {
	MachineUUID string `json:"machineUUID"` // e.g. 123e4567-e89b-12d3-a456-426614174000
	MachineName string `json:"machineName"` // e.g. MacBook-Pro.local
	Auth        string `json:"auth"`
	OSVersion   string `json:"macOSVersion"` // e.g. Ubuntu 20.04
	ModelName   string `json:"modelName"`    // e.g. MacBook Pro
	ModelSerial string `json:"modelSerial"`  // e.g. C02C1234
}

// SystemSerial retrieves the system's serial number by executing a shell command
// that queries hardware information on Darwin (macOS) systems. It returns the
// serial number as a string, or an error if the command fails.
func SystemSerial() (string, error) {
	// Use bash to properly handle pipes
	serial, err := RunCommand("bash", "-c", "system_profiler SPHardwareDataType | grep 'Serial Number (system)' | awk '{print $4}'")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(serial), nil
}

// SystemDevice retrieves the system's device model name by executing a shell command
// that queries the hardware data using 'system_profiler' and processes the output.
// It returns the device model name as a string, or an error if the command fails.
func SystemDevice() (string, error) {
	// Use bash to properly handle pipes and get everything after "Model Name:"
	device, err := RunCommand("bash", "-c", "system_profiler SPHardwareDataType | grep 'Model Name:' | sed 's/.*Model Name: //'")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(device), nil
}
