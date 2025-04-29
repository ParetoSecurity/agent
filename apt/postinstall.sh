#!/bin/bash
set -e

# Check if the script is running on Ubuntu, Debian, or Pop!_OS
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "$ID" == "ubuntu" || "$ID" == "debian" || "$ID" == "pop" ]]; then
        # Create keyrings directory
        mkdir -p --mode=0755 /usr/share/keyrings
        # Download and install GPG key
        curl -fsSL https://pkg.paretosecurity.com/paretosecurity.gpg | tee /usr/share/keyrings/paretosecurity.gpg >/dev/null
        # Add Pareto repository
        echo 'deb [signed-by=/usr/share/keyrings/paretosecurity.gpg] https://pkg.paretosecurity.com/debian stable main' | tee /etc/apt/sources.list.d/pareto.list >/dev/null
    elif [[ "$ID" == *"rhel"* || "$ID" == *"fedora"* ]]; then
        # Download and install GPG key
        rpm --import https://pkg.paretosecurity.com/paretosecurity.asc
        curl -fsSl https://pkg.paretosecurity.com/rpm/paretosecurity.repo | tee /etc/yum.repos.d/paretosecurity.repo >/dev/null
    elif [[ "$ID_LIKE" == "arch" ]]; then
        # Download and install GPG key
        curl -fsSL https://pkg.paretosecurity.com/paretosecurity.gpg | pacman-key --add -
        pacman-key --lsign-key info@niteo.co >/dev/null
        # Add Pareto repository if not already present
        if ! grep -q "\[paretosecurity\]" /etc/pacman.conf; then
            echo '[paretosecurity]' | tee -a /etc/pacman.conf >/dev/null
            echo "Server = https://pkg.paretosecurity.com/aur/stable/$(uname -m)" | tee -a /etc/pacman.conf >/dev/null
        fi
    fi
fi

# Check for systemd
if command -v systemctl >/dev/null 2>&1; then

    # Stop and disable old pareto-core services if they exist
    systemctl stop pareto-core.service 2>/dev/null || true
    systemctl disable pareto-core.service 2>/dev/null || true
    systemctl stop pareto-core.socket 2>/dev/null || true
    systemctl disable pareto-core.socket 2>/dev/null || true

    # Reload systemd and enable socket
    systemctl daemon-reload
    systemctl enable paretosecurity.socket

fi
