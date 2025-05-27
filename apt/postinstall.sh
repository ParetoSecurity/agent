#!/bin/bash
set -e

# Check for systemd
if command -v systemctl >/dev/null 2>&1; then
    # Reload systemd and enable socket
    systemctl daemon-reload
    systemctl enable paretosecurity.service
    systemctl enable --now paretosecurity.socket
fi
