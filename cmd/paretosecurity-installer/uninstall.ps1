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

# Remove URI handler
$URLHandlerKey = "HKCU:\Software\Classes\paretosecurity"
Remove-Item -Path $URLHandlerKey -Recurse -Force -ErrorAction SilentlyContinue

# Remove desktop shortcut
$DesktopDir = [Environment]::GetFolderPath("Desktop")
$DesktopShortcut = Join-Path $DesktopDir "Pareto Security.lnk"
Remove-Item -Path $DesktopShortcut -Force -ErrorAction SilentlyContinue

# Remove startup shortcut
$StartupDir = [Environment]::GetFolderPath("Startup")
$StartupShortcut = Join-Path $StartupDir "Pareto Security.lnk"
Remove-Item -Path $StartupShortcut -Force -ErrorAction SilentlyContinue
