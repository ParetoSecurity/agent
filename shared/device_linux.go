//go:build linux

package shared

type ReportingDevice struct {
	MachineUUID string `json:"machineUUID"` // e.g. 123e4567-e89b-12d3-a456-426614174000
	MachineName string `json:"machineName"` // e.g. MacBook-Pro.local
	Auth        string `json:"auth"`
	OSVersion   string `json:"linuxOSVersion"` // e.g. Ubuntu 20.04
	ModelName   string `json:"modelName"`      // e.g. MacBook Pro
	ModelSerial string `json:"modelSerial"`    // e.g. C02C1234
}

// SystemSerial retrieves the system's serial number by reading the contents of
// "/sys/class/dmi/id/product_serial" using the RunCommand helper function.
// It returns the serial number as a string, or an error if the command fails.
func SystemSerial() (string, error) {
	serial, err := RunCommand("cat", "/sys/class/dmi/id/product_serial")
	if err != nil {
		return "", err
	}
	return serial, nil
}

// SystemDevice retrieves the system's device name by reading the contents of
// "/sys/class/dmi/id/product_name" using the RunCommand function. It returns
// the device name as a string, or an error if the command fails.
func SystemDevice() (string, error) {
	device, err := RunCommand("cat", "/sys/class/dmi/id/product_name")
	if err != nil {
		return "", err
	}
	return device, nil
}
