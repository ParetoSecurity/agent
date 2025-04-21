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

# Download and unzip the latest release
Write-Host "Extracting ParetoSecurity from provided zip..."
Expand-Archive -Path $ZipPath -DestinationPath $InstallPath -Force

# Remove the zip file after extraction
Write-Host "Removing the zip file..."
Remove-Item -Path $ZipPath -Force

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

# Add uninstaller registry entry
Write-Host "Adding uninstaller registry entry..."
$UninstallKey = "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\ParetoSecurity"
if (-Not (Test-Path -Path $UninstallKey)) {
    New-Item -Path $UninstallKey | Out-Null
}
Set-ItemProperty -Path $UninstallKey -Name "DisplayName" -Value "Pareto Security"
Set-ItemProperty -Path $UninstallKey -Name "DisplayVersion" -Value $DisplayVersion
Set-ItemProperty -Path $UninstallKey -Name "Publisher" -Value "Niteo GmbH"
Set-ItemProperty -Path $UninstallKey -Name "InstallLocation" -Value $InstallPath
Set-ItemProperty -Path $UninstallKey -Name "UninstallString" -Value "powershell.exe -ExecutionPolicy Bypass -File $InstallPath\uninstall.ps1"
Set-ItemProperty -Path $UninstallKey -Name "DisplayIcon" -Value (Join-Path $InstallPath "paretosecurity-tray.exe,0")
Set-ItemProperty -Path $UninstallKey -Name "HelpLink" -Value "https://paretosecurity.com/help"
Set-ItemProperty -Path $UninstallKey -Name "URLInfoAbout" -Value "https://paretosecurity.com"

Write-Host "Installation completed successfully."
