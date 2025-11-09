#!/bin/bash
# Installation script for dis-core systemd service

set -e

SERVICE_FILE="dis-core.service"
SYSTEMD_DIR="/etc/systemd/system"
BINARY_DIR="/usr/local/bin"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Installing dis-core as systemd service..."

# Build the binary
echo "Building dis-core binary..."
cd "$PROJECT_DIR"
go build -o dis-core ./cmd/dis-core

# Install binary
echo "Installing binary to $BINARY_DIR..."
sudo cp dis-core "$BINARY_DIR/"
sudo chmod +x "$BINARY_DIR/dis-core"

# Install systemd service
echo "Installing systemd service file..."
sudo cp "scripts/$SERVICE_FILE" "$SYSTEMD_DIR/"

# Reload systemd
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable service
echo "Enabling dis-core service..."
sudo systemctl enable dis-core

echo "Installation complete!"
echo ""
echo "To start the service: sudo systemctl start dis-core"
echo "To check status:     sudo systemctl status dis-core"
echo "To view logs:        journalctl -u dis-core -f"
echo ""
echo "Note: Ensure DIS_DB_DSN environment variable is set in /etc/environment or in the service file."
