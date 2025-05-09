#!/bin/sh
# Copyright (c) Tailscale Inc
# Copyright (c) 2024 The Brave Authors
# Copyright (c) 2025 The ParetoSecurity Authors
# SPDX-License-Identifier: BSD-3-Clause
#
# This script installs the ParetoSecurity using the OS's package manager
# Requires: coreutils, grep, sh and one of sudo/doas/run0/pkexec/sudo-rs
# Source: https://github.com/brave/install.sh

set -eu

BASE_URL="https://github.com/ParetoSecurity/agent/releases/latest/download/paretosecurity_"

main() {

    # Determine architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64|aarch64) ;;
        *) error "Unsupported architecture: $ARCH";;
    esac

    # Locate necessary tools
    sudo=""
    if [[ "$(whoami)" != "root" ]]; then
        sudo="$(first_of sudo doas run0 pkexec sudo-rs)" || error "Please install sudo/doas/run0/pkexec/sudo-rs to proceed."
    fi

    case "$(first_of curl wget)" in
        wget) curl="wget -qO- --trust-server-names";;
        *) curl="curl -fsSL";;
    esac

    echo "Starting installation of Pareto Security..."

    # Package manager detection and installation
    if available apt-get; then
        install_apt
    elif available dnf; then
        install_dnf
    elif available pacman; then
        install_pacman
    elif available yum; then
        install_yum
    else
        error "Could not find a supported package manager (apt, dnf, pacman, yum)."
    fi

    echo "Pareto Security has been installed successfully."
    echo "From now on, your package manager will automatically handle updates."
    echo "To initiate the application, kindly execute the ‘paretosecurity check’ command. Alternatively, you may proceed with joining the team from the dashboard."
}

install_apt() {
    TEMP_DIR=$(mktemp -d)
    echo "Using apt-get..."
    if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
        PACKAGE="paretosecurity_amd64.deb"
        ARCH_SUFFIX="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        PACKAGE="paretosecurity_arm64.deb"
        ARCH_SUFFIX="arm64"
    else
        error "Unsupported architecture: $ARCH"
    fi
    
    if dpkg -s "paretosecurity" >/dev/null 2>&1; then
        echo "Pareto Security is already installed, updating..."
        show $curl "${BASE_URL}${ARCH_SUFFIX}.deb" -o "$TEMP_DIR/$PACKAGE"
        show $sudo dpkg -i "$TEMP_DIR/$PACKAGE"
        show $sudo apt-get install -f -y # Fix dependencies
    else
        echo "Installing Pareto Security..."
        show $curl "${BASE_URL}${ARCH_SUFFIX}.deb" -o "$TEMP_DIR/$PACKAGE"
        show $sudo dpkg -i "$TEMP_DIR/$PACKAGE"
        show $sudo apt-get install -f -y # Fix dependencies
    fi
    echo "Cleaning up..."
    rm -rf "$TEMP_DIR"
}

install_dnf() {
    TEMP_DIR=$(mktemp -d)
    echo "Using dnf..."
    if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
        PACKAGE="paretosecurity_amd64.rpm"
        ARCH_SUFFIX="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        PACKAGE="paretosecurity_arm64.rpm"
        ARCH_SUFFIX="arm64"
    else
        error "Unsupported architecture: $ARCH"
    fi
    
    if rpm -q "paretosecurity" >/dev/null 2>&1; then
        echo "Pareto Security is already installed, updating..."
        show $sudo rpm -e "paretosecurity"
        show $curl "${BASE_URL}${ARCH_SUFFIX}.rpm" -o "$TEMP_DIR/$PACKAGE"
        show $sudo rpm -i "$TEMP_DIR/$PACKAGE"
    else
        echo "Installing Pareto Security..."
        show $curl "${BASE_URL}${ARCH_SUFFIX}.rpm" -o "$TEMP_DIR/$PACKAGE"
        show $sudo rpm -i "$TEMP_DIR/$PACKAGE"
    fi
    echo "Cleaning up..."
    rm -rf "$TEMP_DIR"
}

install_pacman() {
    TEMP_DIR=$(mktemp -d)
    echo "Using pacman..."
    if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
        PACKAGE="paretosecurity_amd64.archlinux.pkg.tar.zst"
        ARCH_SUFFIX="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        PACKAGE="paretosecurity_arm64.archlinux.pkg.tar.zst"
        ARCH_SUFFIX="arm64"
    else
        error "Unsupported architecture: $ARCH"
    fi
    
    if pacman -Q "paretosecurity" >/dev/null 2>&1; then
       echo "Pareto Security is already installed, updating..."
       show $curl "${BASE_URL}${ARCH_SUFFIX}.archlinux.pkg.tar.zst" -o "$TEMP_DIR/$PACKAGE"
       show $sudo pacman -U --noconfirm "$TEMP_DIR/$PACKAGE"
    else
       echo "Installing Pareto Security..."
       show $curl "${BASE_URL}${ARCH_SUFFIX}.archlinux.pkg.tar.zst" -o "$TEMP_DIR/$PACKAGE"
       show $sudo pacman -U --noconfirm "$TEMP_DIR/$PACKAGE"
    fi
    echo "Cleaning up..."
    rm -rf "$TEMP_DIR"
}

install_yum() {
    TEMP_DIR=$(mktemp -d)
    echo "Using yum..."
    if [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
        PACKAGE="paretosecurity_amd64.rpm"
        ARCH_SUFFIX="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        PACKAGE="paretosecurity_arm64.rpm"
        ARCH_SUFFIX="arm64"
    else
        error "Unsupported architecture: $ARCH"
    fi
    
    show $curl "${BASE_URL}${ARCH_SUFFIX}.rpm" -o "$TEMP_DIR/$PACKAGE"
    show $sudo yum -y install "$TEMP_DIR/$PACKAGE"
    echo "Cleaning up..."
    rm -rf "$TEMP_DIR"
}

# Helpers
available() { command -v "${1:?}" >/dev/null; }
first_of() { for c in "${@:?}"; do if available "$c"; then echo "$c"; return 0; fi; done; return 1; }
show() { (set -x; "${@:?}"); }
error() { exec >&2; printf "Error: "; printf "%s\n" "${@:?}"; exit 1; }

main