package checks

import (
	"errors"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestWindowsDefender_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Windows Defender active (productState 397568)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `{"displayName":"Windows Defender","instanceGuid":"{D68DDC3A-831F-4fae-9E44-DA132C1ACF46}","pathToSignedProductExe":"windowsdefender://","pathToSignedReportingExe":"%ProgramFiles%\\Windows Defender\\MsMpeng.exe","productState":"397568","timestamp":"Tue, 01 Jul 2025 08:16:52 GMT"}`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "Third-party antivirus active (Norton Security)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `{"displayName":"Norton Security","instanceGuid":"{12345678-1234-1234-1234-123456789012}","pathToSignedProductExe":"C:\\Program Files\\Norton Security\\Engine\\norton.exe","pathToSignedReportingExe":"C:\\Program Files\\Norton Security\\Engine\\norton.exe","productState":"266256","timestamp":"Tue, 01 Jul 2025 08:16:52 GMT"}`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "Antivirus disabled (real-time protection off - productState 262144)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `{"displayName":"Windows Defender","instanceGuid":"{D68DDC3A-831F-4fae-9E44-DA132C1ACF46}","pathToSignedProductExe":"windowsdefender://","pathToSignedReportingExe":"%ProgramFiles%\\Windows Defender\\MsMpeng.exe","productState":"262144","timestamp":"Tue, 01 Jul 2025 08:16:52 GMT"}`,
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "No active antivirus or EDR software detected",
		},
		{
			name: "Antivirus enabled but definitions outdated (productState 266256)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `{"displayName":"Norton Security","instanceGuid":"{12345678-1234-1234-1234-123456789012}","pathToSignedProductExe":"C:\\Program Files\\Norton Security\\Engine\\norton.exe","pathToSignedReportingExe":"C:\\Program Files\\Norton Security\\Engine\\norton.exe","productState":"266256","timestamp":"Tue, 01 Jul 2025 08:16:52 GMT"}`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "Multiple antivirus products with one disabled (array output)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `[{"displayName":"Windows Defender","instanceGuid":"{D68DDC3A-831F-4fae-9E44-DA132C1ACF46}","pathToSignedProductExe":"windowsdefender://","pathToSignedReportingExe":"%ProgramFiles%\\Windows Defender\\MsMpeng.exe","productState":"262144","timestamp":"Tue, 01 Jul 2025 08:16:52 GMT"},{"displayName":"Kaspersky Internet Security","instanceGuid":"{12345678-1234-1234-1234-123456789012}","pathToSignedProductExe":"C:\\Program Files\\Kaspersky\\kaspersky.exe","pathToSignedReportingExe":"C:\\Program Files\\Kaspersky\\kaspersky.exe","productState":"266240","timestamp":"Tue, 01 Jul 2025 08:16:52 GMT"}]`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "Empty output (no antivirus detected)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     "",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "No active antivirus or EDR software detected",
		},
		{
			name: "PowerShell command fails",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     "",
					Err:     errors.New("powershell failed"),
				},
			},
			expectedPassed: false,
			expectedStatus: "No active antivirus or EDR software detected",
		},
		{
			name: "Invalid JSON output",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `invalid-json`,
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "No active antivirus or EDR software detected",
		},
		{
			name: "AVG Antivirus active (real-world example - productState 266240)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `{"displayName":"AVG Antivirus","instanceGuid":"{9CA89623-EDDE-E3D8-41CD-9C8961430EF3}","pathToSignedProductExe":"C:\\Program Files\\AVG\\Antivirus\\wsc_proxy.exe","pathToSignedReportingExe":"C:\\Program Files\\AVG\\Antivirus\\wsc_proxy.exe","productState":"266240","timestamp":"Tue, 01 Jul 2025 08:52:21 GMT"}`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "Windows Defender disabled (real-world example - productState 393472)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `{"displayName":"Windows Defender","instanceGuid":"{D68DDC3A-831F-4fae-9E44-DA132C1ACF46}","pathToSignedProductExe":"windowsdefender://","pathToSignedReportingExe":"%ProgramFiles%\\Windows Defender\\MsMpeng.exe","productState":"393472","timestamp":"Tue, 01 Jul 2025 08:52:33 GMT"}`,
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "No active antivirus or EDR software detected",
		},
		{
			name: "Real-world multiple products (AVG active, Defender disabled)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     `[{"displayName":"AVG Antivirus","instanceGuid":"{9CA89623-EDDE-E3D8-41CD-9C8961430EF3}","pathToSignedProductExe":"C:\\Program Files\\AVG\\Antivirus\\wsc_proxy.exe","pathToSignedReportingExe":"C:\\Program Files\\AVG\\Antivirus\\wsc_proxy.exe","productState":"266240","timestamp":"Tue, 01 Jul 2025 08:52:21 GMT"},{"displayName":"Windows Defender","instanceGuid":"{D68DDC3A-831F-4fae-9E44-DA132C1ACF46}","pathToSignedProductExe":"windowsdefender://","pathToSignedReportingExe":"%ProgramFiles%\\Windows Defender\\MsMpeng.exe","productState":"393472","timestamp":"Tue, 01 Jul 2025 08:52:33 GMT"}]`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "CrowdStrike Falcon EDR detected (no traditional AV)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     "",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Process -Name CSFalconService -ErrorAction SilentlyContinue | Select-Object Name"},
					Out:     "Name\n----\nCSFalconService",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "SentinelOne EDR detected via service",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     "",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Process -Name SentinelAgent -ErrorAction SilentlyContinue | Select-Object Name"},
					Out:     "",
					Err:     errors.New("process not found"),
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'SentinelAgent' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "Status\n------\nRunning",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "Carbon Black EDR detected via registry",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     "",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Process -Name cb -ErrorAction SilentlyContinue | Select-Object Name"},
					Out:     "",
					Err:     errors.New("process not found"),
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'CbDefense' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "",
					Err:     errors.New("service not found"),
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Test-Path 'HKLM:\\SOFTWARE\\CarbonBlack'"},
					Out:     "True",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus or EDR software is active",
		},
		{
			name: "No antivirus or EDR detected (comprehensive check)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json"},
					Out:     "",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Process -Name CSFalconService -ErrorAction SilentlyContinue | Select-Object Name"},
					Out:     "",
					Err:     errors.New("process not found"),
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'CSAgent' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "",
					Err:     errors.New("service not found"),
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Test-Path 'HKLM:\\SYSTEM\\CrowdStrike'"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Test-Path '$env:ProgramFiles\\CrowdStrike'"},
					Out:     "False",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "No active antivirus or EDR software detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mockCommands

			check := &WindowsDefender{}
			err := check.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, check.Passed())
			assert.Equal(t, tt.expectedStatus, check.Status())
		})
	}
}
