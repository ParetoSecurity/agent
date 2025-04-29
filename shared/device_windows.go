//go:build windows
// +build windows

package shared

type ReportingDevice struct {
	MachineUUID string `json:"machineUUID"` // e.g. 123e4567-e89b-12d3-a456-426614174000
	MachineName string `json:"machineName"` // e.g. MacBook-Pro.local
	Auth        string `json:"auth"`
	OSVersion   string `json:"windowsOSVersion"` // e.g. Windows 10 Pro (24H2)
	ModelName   string `json:"modelName"`        // e.g. MacBook Pro
	ModelSerial string `json:"modelSerial"`      // e.g. C02C1234
}

// SystemSerial retrieves the system's BIOS serial number by executing a PowerShell command.
// It returns the serial number as a string, or an error if the command fails.
func SystemSerial() (string, error) {
	serial, err := RunCommand("powershell", "-Command", `(Get-WmiObject -Class Win32_BIOS).SerialNumber`)
	if err != nil {
		return "", err
	}
	return serial, nil
}

// SystemDevice retrieves the model name of the current Windows computer system
// by executing a PowerShell command that queries the Win32_ComputerSystem WMI class.
// It returns the device model as a string, or an error if the command fails.
func SystemDevice() (string, error) {
	device, err := RunCommand("powershell", "-Command", `(Get-WmiObject -Class Win32_ComputerSystem).Model`)
	if err != nil {
		return "", err
	}
	return device, nil
}
