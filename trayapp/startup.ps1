param (
    [string]$Action,           # "enable" or "disable"
    [string]$InstallPath = ""  # Path to the installation directory
)

$ErrorActionPreference = "Stop"

# Define paths
$RoamingDir = [Environment]::GetFolderPath("ApplicationData")
$DefaultInstallPath = Join-Path $RoamingDir "ParetoSecurity"
$StartupShortcut = Join-Path $RoamingDir "Microsoft\Windows\Start Menu\Programs\Startup\Pareto Security.lnk"

# Use provided install path or default
if ($InstallPath -eq "") {
    $InstallPath = $DefaultInstallPath
}

$TrayExePath = Join-Path $InstallPath "paretosecurity-tray.exe"

function Enable-Startup {
    Write-Host "Enabling startup shortcut..."
    
    # Ensure the startup directory exists
    $StartupDir = Split-Path $StartupShortcut -Parent
    if (-Not (Test-Path -Path $StartupDir)) {
        New-Item -ItemType Directory -Path $StartupDir -Force | Out-Null
    }
    
    # Verify the target executable exists
    if (-Not (Test-Path -Path $TrayExePath)) {
        throw "Target executable not found: $TrayExePath"
    }
    
    # Create the shortcut
    $WScriptShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WScriptShell.CreateShortcut($StartupShortcut)
    $Shortcut.TargetPath = $TrayExePath
    $Shortcut.Description = "Pareto Security"
    $Shortcut.WorkingDirectory = $InstallPath
    $Shortcut.Save()
    
    Write-Host "Startup shortcut created successfully at: $StartupShortcut"
}

function Disable-Startup {
    Write-Host "Disabling startup shortcut..."
    
    if (Test-Path -Path $StartupShortcut) {
        Remove-Item -Path $StartupShortcut -Force
        Write-Host "Startup shortcut removed successfully"
    } else {
        Write-Host "Startup shortcut was not found, nothing to remove"
    }
}

# Main execution
switch ($Action.ToLower()) {
    "enable" {
        Enable-Startup
        exit 0
    }
    "disable" {
        Disable-Startup
        exit 0
    }
    default {
        Write-Host "Invalid action. Use 'enable' or 'disable'"
        exit 2
    }
}