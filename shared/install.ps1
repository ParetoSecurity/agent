param (
    [string]$ZipPath,
    [switch]$WithStartup,
    [string]$DisplayVersion = "1.0.0" # Default version
)

$ErrorActionPreference = "Stop"

# Define paths
$RoamingDir = [Environment]::GetFolderPath("ApplicationData")
$InstallPath = Join-Path $RoamingDir "ParetoSecurity"
$DesktopShortcut = Join-Path ([Environment]::GetFolderPath("Desktop")) "Pareto Security.lnk"
$StartupShortcut = Join-Path $RoamingDir "Microsoft\Windows\Start Menu\Programs\Startup\Pareto Security.lnk"

# Create installation directory
if (-Not (Test-Path -Path $InstallPath)) {
    New-Item -ItemType Directory -Path $InstallPath | Out-Null
}

# Close running instances of ParetoSecurity applications
Write-Host "Closing running instances of ParetoSecurity..."
$procNames = @("paretosecurity-tray.exe", "paretosecurity.exe", "paretosecurity-tray", "paretosecurity")
foreach ($name in $procNames) {
    do {
        $procs = Get-Process -Name $name -ErrorAction SilentlyContinue
        if ($procs) {
            $procs | Stop-Process -Force -ErrorAction SilentlyContinue
            Start-Sleep -Milliseconds 300
        }
    } while ($procs)
}

# Download and unzip the latest release
Write-Host "Extracting ParetoSecurity from provided zip..."
Expand-Archive -Path $ZipPath -DestinationPath $InstallPath -Force

# Create desktop shortcut
Write-Host "Creating desktop shortcut..."
$WScriptShell = New-Object -ComObject WScript.Shell
$Shortcut = $WScriptShell.CreateShortcut($DesktopShortcut)
$Shortcut.TargetPath = Join-Path $InstallPath "paretosecurity-tray.exe"
$Shortcut.Description = "Pareto Security"
$Shortcut.Save()

# Create startup shortcut if requested
if ($WithStartup) {
    Write-Host "Creating startup shortcut..."
    $StartupShortcutObj = $WScriptShell.CreateShortcut($StartupShortcut)
    $StartupShortcutObj.TargetPath = Join-Path $InstallPath "paretosecurity-tray.exe"
    $StartupShortcutObj.Description = "Pareto Security"
    $StartupShortcutObj.Save()
}

# Register paretosecurity:// URL handler
Write-Host "Registering paretosecurity:// URL handler..."
$URLHandlerKey = "HKCU:\Software\Classes\paretosecurity"
if (-Not (Test-Path -Path $URLHandlerKey)) {
    New-Item -Path $URLHandlerKey -Force | Out-Null
}
Set-ItemProperty -Path $URLHandlerKey -Name "(Default)" -Value "URL:ParetoSecurity Protocol"
Set-ItemProperty -Path $URLHandlerKey -Name "URL Protocol" -Value ""

$CommandKey = Join-Path $URLHandlerKey "shell\open\command"
if (-Not (Test-Path -Path $CommandKey)) {
    New-Item -Path $CommandKey -Force | Out-Null
}
Set-ItemProperty -Path $CommandKey -Name "(Default)" -Value ('"' + $InstallPath + '\paretosecurity-tray.exe" link "%1"')

# Launch ParetoSecurity tray application
Write-Host "Launching ParetoSecurity tray application..."
Start-Process -FilePath (Join-Path $InstallPath "paretosecurity-tray.exe")

# Add uninstaller registry entry
Write-Host "Adding uninstaller registry entry..."
$UninstallKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\ParetoSecurity"
if (-Not (Test-Path -Path $UninstallKey)) {
    New-Item -Path $UninstallKey -Force | Out-Null
}
Set-ItemProperty -Path $UninstallKey -Name "DisplayName" -Value "Pareto Security"
Set-ItemProperty -Path $UninstallKey -Name "DisplayVersion" -Value $DisplayVersion
Set-ItemProperty -Path $UninstallKey -Name "Publisher" -Value "Niteo GmbH"
Set-ItemProperty -Path $UninstallKey -Name "InstallLocation" -Value $InstallPath
Set-ItemProperty -Path $UninstallKey -Name "UninstallString" -Value "powershell.exe -ExecutionPolicy Bypass -File $InstallPath\uninstall.ps1"
Set-ItemProperty -Path $UninstallKey -Name "DisplayIcon" -Value (Join-Path $InstallPath "paretosecurity-tray.exe,0")
Set-ItemProperty -Path $UninstallKey -Name "HelpLink" -Value "https://paretosecurity.com/help"
Set-ItemProperty -Path $UninstallKey -Name "URLInfoAbout" -Value "https://paretosecurity.com"