//go:build windows
// +build windows

package shared

type ReportingDevice struct {
	MachineUUID string `json:"machineUUID"` // e.g. 123e4567-e89b-12d3-a456-426614174000
	MachineName string `json:"machineName"` // e.g. MacBook-Pro.local
	Auth        string `json:"auth"`
	OSVersion   string `json:"windowOSVersion"` // e.g. Windows 10 Pro (24H2)
	ModelName   string `json:"modelName"`       // e.g. MacBook Pro
	ModelSerial string `json:"modelSerial"`     // e.g. C02C1234
}
