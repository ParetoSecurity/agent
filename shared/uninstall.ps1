$ErrorActionPreference = "Stop"

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

# Remove installation directory
$RoamingDir = [Environment]::GetFolderPath("ApplicationData")
$InstallPath = Join-Path $RoamingDir "ParetoSecurity"
Remove-Item -Recurse -Force -Path $InstallPath -ErrorAction SilentlyContinue

# Remove uninstaller registry entry
Remove-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\ParetoSecurity" -Force -ErrorAction SilentlyContinue

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
