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
            show $sudo dnf config-manager addrepo --overwrite --from-repofile=https://pkg.paretosecurity.com/rpm/paretosecurity.repo
        else
            show $sudo dnf install -y 'dnf-command(config-manager)'
            show $sudo dnf config-manager --add-repo https://pkg.paretosecurity.com/rpm/paretosecurity.repo
        fi
        show $sudo dnf install -y paretosecurity

    elif available pacman; then
        if pacman -Ss paretosecurity >/dev/null 2>&1; then
            show $sudo pacman -Sy --needed --noconfirm paretosecurity
        else
            show curl -fsSL https://pkg.paretosecurity.com/paretosecurity.gpg | show $sudo pacman-key --add -
            show $sudo pacman-key --lsign-key info@niteo.co
            show $sudo pacman -Sy --needed --noconfirm paretosecurity
        fi

    elif available zypper; then
        show $sudo zypper --non-interactive addrepo --gpgcheck --repo https://pkg.paretosecurity.com/rpm/paretosecurity.repo
        show $sudo zypper --non-interactive --gpg-auto-import-keys refresh
        show $sudo zypper --non-interactive install paretosecurity

    elif available yum; then
        available yum-config-manager || show $sudo yum install yum-utils -y
        show $sudo yum-config-manager -y --add-repo https://pkg.paretosecurity.com/rpm/paretosecurity.repo
        show $sudo yum install paretosecurity -y

    elif available rpm-ostree; then
        available curl || available wget || error "Please install curl/wget to proceed."
        show $curl https://pkg.paretosecurity.com/rpm/paretosecurity.repo|\
            show $sudo install -DTm644 /dev/stdin /etc/yum.repos.d/paretosecurity.repo
        show $sudo rpm-ostree install -y --idempotent paretosecurity

    else
        error "Could not find a supported package manager. Only apt/dnf/pacman(+paru/pikaur/yay)/rpm-ostree/yum/zypper are supported." "" \
            "If you'd like us to support your system better, please file an issue at" \
            "https://github.com/paretosecurity/agent/issues and include the following information:" "" \
            "$(uname -srvmo || true)" "" \
            "$(cat /etc/os-release || true)"
    fi

    echo "Pareto Security has been installed successfully."
    echo "From now on, your package manager will automatically handle updates."
    echo "To initiate the application, kindly execute the ‘paretosecurity check’ command. Alternatively, you may proceed with joining the team from the dashboard."

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