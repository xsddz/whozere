#!/bin/bash
#
# whozere - Uninstallation script
# Usage: curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/scripts/uninstall.sh | bash
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/usr/local/etc/whozere"
SYSTEMD_DIR="/etc/systemd/system"
LAUNCHD_DIR="$HOME/Library/LaunchAgents"
SERVICE_NAME="whozere"

# Print functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Stop running service
stop_service() {
    info "Stopping whozere service..."
    
    # Check for systemd (Linux)
    if command -v systemctl &> /dev/null; then
        if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
            sudo systemctl stop "$SERVICE_NAME" 2>/dev/null || true
            sudo systemctl disable "$SERVICE_NAME" 2>/dev/null || true
            info "Stopped systemd service"
        fi
        if [ -f "$SYSTEMD_DIR/${SERVICE_NAME}.service" ]; then
            sudo rm -f "$SYSTEMD_DIR/${SERVICE_NAME}.service"
            sudo systemctl daemon-reload
            info "Removed systemd service file"
        fi
    fi
    
    # Check for launchd (macOS)
    if command -v launchctl &> /dev/null; then
        local PLIST="com.${SERVICE_NAME}.agent.plist"
        if launchctl list | grep -q "$SERVICE_NAME" 2>/dev/null; then
            launchctl unload "$LAUNCHD_DIR/$PLIST" 2>/dev/null || true
            info "Stopped launchd service"
        fi
        if [ -f "$LAUNCHD_DIR/$PLIST" ]; then
            rm -f "$LAUNCHD_DIR/$PLIST"
            info "Removed launchd plist file"
        fi
    fi
    
    # Kill any running process
    pkill -f "whozere" 2>/dev/null || true
}

# Remove binary and helper scripts
remove_binary() {
    local BINARY="$INSTALL_DIR/whozere"
    local SERVICE_SCRIPT="$INSTALL_DIR/whozere-service"
    local UNINSTALL_SCRIPT="$INSTALL_DIR/whozere-uninstall"
    
    info "Removing whozere files..."
    
    for FILE in "$BINARY" "$SERVICE_SCRIPT" "$UNINSTALL_SCRIPT"; do
        if [ -f "$FILE" ]; then
            if [ -w "$INSTALL_DIR" ]; then
                rm -f "$FILE"
            else
                sudo rm -f "$FILE"
            fi
            success "Removed $FILE"
        fi
    done
}

# Remove config (with confirmation)
remove_config() {
    if [ -d "$CONFIG_DIR" ]; then
        echo ""
        read -p "Remove configuration directory $CONFIG_DIR? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if [ -w "$(dirname $CONFIG_DIR)" ]; then
                rm -rf "$CONFIG_DIR"
            else
                sudo rm -rf "$CONFIG_DIR"
            fi
            success "Removed configuration directory"
        else
            info "Kept configuration directory"
        fi
    fi
}

# Main
main() {
    echo ""
    echo "     _       ____  ______  __________ ____  ____"
    echo "    | |     / / / / / __ \\/__  / ____/ __ \\/ __/"
    echo "    | | /| / / /_/ / / / / / / / __/ / /_/ / _/  "
    echo "    | |/ |/ / __  / /_/ / / / / /___/ _, _/ /___"
    echo "    |__/|__/_/ /_/\\____/ /_/ /_____/_/ |_/_____/"
    echo ""
    echo "    Uninstaller"
    echo ""

    stop_service
    remove_binary
    
    # Only prompt for config removal in interactive mode
    if [ -t 0 ]; then
        remove_config
    else
        info "Non-interactive mode: keeping configuration at $CONFIG_DIR"
    fi

    echo ""
    success "whozere has been uninstalled"
}

main "$@"
