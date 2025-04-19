$ErrorActionPreference = "Stop"

# Stop the ParetoSecurity process
Stop-Process -Name "paretosecurity-tray" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "paretosecurity" -Force -ErrorAction SilentlyContinue

# Remove installation directory
$RoamingDir = [Environment]::GetFolderPath("ApplicationData")
$InstallPath = Join-Path $RoamingDir "ParetoSecurity"
Remove-Item -Recurse -Force -Path $InstallPath -ErrorAction SilentlyContinue

# Remove uninstaller registry entry
Remove-Item -Path "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\ParetoSecurity" -Force -ErrorAction SilentlyContinue
