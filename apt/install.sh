#!/bin/sh
# Copyright (c) Tailscale Inc
# Copyright (c) 2024 The Brave Authors
# Copyright (c) 2025 Pareto Security
# SPDX-License-Identifier: BSD-3-Clause
#
# This script installs the Pareto Security using the OS's package manager
# Requires: coreutils, grep, sh and one of sudo/doas/run0/pkexec/sudo-rs
# Source: https://github.com/brave/install.sh

APT_VER_MIN="1.1"

set -eu

# All the code is wrapped in a main function that gets called at the
# bottom of the file, so that a truncated partial download doesn't end
# up executing half a script.
main() {
    ## Check if the app can run on this system

    case "$(uname -m)" in
        aarch64|x86_64) ;;
        *) error "Unsupported architecture $(uname -m). Only 64-bit x86 or ARM machines are supported.";;
    esac

    ## Locate the necessary tools

    case "$(whoami)" in
        root) sudo="";;
        *) sudo="$(first_of sudo doas run0 pkexec sudo-rs)" || error "Please install sudo/doas/run0/pkexec/sudo-rs to proceed.";;
    esac

    case "$(first_of curl wget)" in
        wget) curl="wget -qO- --trust-server-names";;
        *) curl="curl -fsSL";;
    esac

    ## Install the app

    if available apt-get && apt_supported; then
        export DEBIAN_FRONTEND=noninteractive
        if ! available curl && ! available wget; then
            show $sudo apt-get update
            show $sudo apt-get install -y curl
        fi

        show $curl "https://pkg.paretosecurity.com/paretosecurity.gpg"|\
            show $sudo install -DTm644 /dev/stdin /usr/share/keyrings/paretosecurity.gpg
        echo 'deb [signed-by=/usr/share/keyrings/paretosecurity.gpg] https://pkg.paretosecurity.com/debian stable main' |\
            show $sudo install -DTm644 /dev/stdin /etc/apt/sources.list.d/pareto.list
        show $sudo apt-get update
        show $sudo apt-get install -y paretosecurity

    elif available dnf; then
        if dnf --version|grep -q dnf5; then
            show $sudo rpm --import https://pkg.paretosecurity.com/paretosecurity.asc
            show $sudo dnf config-manager addrepo --overwrite --from-repofile="https://pkg.paretosecurity.com/rpm/paretosecurity.repo"
        else
            show $sudo rpm --import https://pkg.paretosecurity.com/paretosecurity.asc
            show $sudo dnf install -y 'dnf-command(config-manager)'
            show $sudo dnf config-manager --add-repo "https://pkg.paretosecurity.com/rpm/paretosecurity.repo"
        fi

        if which paretosecurity >/dev/null 2>&1; then
            show $sudo dnf upgrade -y paretosecurity
        else
            show $sudo dnf install -y paretosecurity
        fi
        
    elif available pacman; then
        if pacman -Ss paretosecurity >/dev/null 2>&1; then
            show $sudo pacman -Sy --needed --noconfirm paretosecurity
        else
            show $curl https://pkg.paretosecurity.com/paretosecurity.gpg |\
                show $sudo pacman-key --add -
            show $sudo pacman-key --lsign-key info@niteo.co
            if ! grep -q "\[paretosecurity\]" /etc/pacman.conf; then
                show echo '[paretosecurity]' | show $sudo tee -a /etc/pacman.conf >/dev/null
                ARCH="$(uname -m)"
                if [ "$ARCH" = "x86_64" ]; then
                    ARCH="amd64"
                elif [ "$ARCH" = "aarch64" ]; then
                    ARCH="aarch64"
                else
                    error "Unsupported architecture $(uname -m). Only 64-bit x86 or ARM machines are supported."
                fi
                show echo "Server = https://pkg.paretosecurity.com/aur/stable/$ARCH" | show $sudo tee -a /etc/pacman.conf >/dev/null
            fi
            show $sudo pacman -Syu --needed --noconfirm paretosecurity
        fi

    else
        error "Could not find a supported package manager. Only apt/dnf/pacman are supported." "" \
            "If you'd like us to support your system better, please file an issue at" \
            "https://github.com/paretosecurity/agent/issues and include the following information:" "" \
            "$(uname -srvmo || true)" "" \
            "$(cat /etc/os-release || true)"
    fi

    echo "Pareto Security has been installed successfully."
    echo "From now on, your package manager will automatically handle the updates."
    echo "If you encounter any issues, please refer to the documentation at https://paretosecurity.com/docs"

    if [ ! -S /run/paretosecurity.sock ]; then
        echo "To use the Pareto Security CLI, please restart the device."
    else
        echo "You can also use the Pareto Security CLI by running 'paretosecurity check'."
    fi

}

# Helpers
available() { command -v "${1:?}" >/dev/null; }
first_of() { for c in "${@:?}"; do if available "$c"; then echo "$c"; return 0; fi; done; return 1; }
show() { (set -x; "${@:?}"); }
error() { exec >&2; printf "Error: "; printf "%s\n" "${@:?}"; exit 1; }
newer() { [ "$(printf "%s\n%s" "$1" "$2"|sort -V|head -n1)" = "${2:?}" ]; }
supported() { newer "$2" "${3:?}" || error "Unsupported ${1:?} version ${2:-<empty>}. Only $1 versions >=$3 are supported."; }
apt_supported() { supported apt "$(apt-get -v|head -n1|cut -d' ' -f2)" "${APT_VER_MIN:?}"; }

main