#!/bin/bash

BASE_URL="https://github.com/ParetoSecurity/agent/releases/latest/download/paretosecurity_"

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit
fi
ARCH=$(uname -m)

echo "Starting installation of Pareto Security..."

# Check if the script is running on Ubuntu, Debian, or Pop!_OS
if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "$ID_LIKE" == *"debian"* ]]; then
        TEMP_DIR=$(mktemp -d)
        echo "Downloading Pareto Security package for $ARCH..."
        if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
            wget -q --show-progress -O "$TEMP_DIR/paretosecurity_amd64.deb" "${BASE_URL}amd64.deb"
            echo "Installing package..."
            dpkg -i "$TEMP_DIR/paretosecurity_amd64.deb"
        elif [[ "$ARCH" == "aarch64" ]]; then
            wget -q --show-progress -O "$TEMP_DIR/paretosecurity_arm64.deb" "${BASE_URL}arm64.deb"
            echo "Installing package..."
            dpkg -i "$TEMP_DIR/paretosecurity_arm64.deb"
        else
            echo "Unsupported architecture: $ARCH"
            exit 1
        fi
        echo "Cleaning up..."
        rm -rf "$TEMP_DIR"

    elif [[ "$ID" == *"rhel"* || "$ID" == *"fedora"* ]]; then
        TEMP_DIR=$(mktemp -d)
        echo "Downloading Pareto Security package for $ARCH..."
        if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
            wget -q --show-progress -O "$TEMP_DIR/paretosecurity_amd64.rpm" "${BASE_URL}amd64.rpm"
            echo "Installing package..."
            pacman -U "$TEMP_DIR/paretosecurity_amd64.archlinux.pkg.tar.zst"
        elif [[ "$ARCH" == "aarch64" ]]; then
            wget -q --show-progress -O "$TEMP_DIR/paretosecurity_arm64.rpm" "${BASE_URL}arm64.rpm"
            echo "Installing package..."
            pacman -U "$TEMP_DIR/paretosecurity_arm64.archlinux.pkg.tar.zst"
        else
            echo "Unsupported architecture: $ARCH"
            exit 1
        fi
    elif [[ "$ID_LIKE" == *"arch"* ]]; then
        TEMP_DIR=$(mktemp -d)
        echo "Downloading Pareto Security package for $ARCH..."
        if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
            wget -q --show-progress -O "$TEMP_DIR/paretosecurity_amd64.rpm" "${BASE_URL}amd64.archlinux.pkg.tar.zst"
            echo "Installing package..."
            rpm -i "$TEMP_DIR/paretosecurity_amd64.rpm"
        elif [[ "$ARCH" == "aarch64" ]]; then
            wget -q --show-progress -O "$TEMP_DIR/paretosecurity_arm64.rpm" "${BASE_URL}arm64.archlinux.pkg.tar.zst"
            echo "Installing package..."
            rpm -i "$TEMP_DIR/paretosecurity_arm64.rpm"
        else
            echo "Unsupported architecture: $ARCH"
            exit 1
        fi
        echo "Cleaning up..."
        rm -rf "$TEMP_DIR"
    fi
fi

echo "Pareto Security has been installed successfully."
echo "From now on, your package manager will automatically handle updates."
echo "To initiate the application, kindly execute the ‘paretosecurity check’ command. Alternatively, you may proceed with joining the team from the dashboard."
