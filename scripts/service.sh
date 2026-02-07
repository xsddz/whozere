#!/bin/bash
#
# whozere - Service management script
# Usage: ./service.sh [install|uninstall|start|stop|restart|status]
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="whozere"
BINARY_PATH="/usr/local/bin/whozere"
CONFIG_PATH="/usr/local/etc/whozere/config.yaml"
SYSTEMD_DIR="/etc/systemd/system"
LAUNCHD_DIR="$HOME/Library/LaunchAgents"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

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
    exit 1
}

# Check prerequisites
check_prereqs() {
    if [ ! -f "$BINARY_PATH" ]; then
        error "whozere binary not found at $BINARY_PATH. Please install first."
    fi
    if [ ! -f "$CONFIG_PATH" ]; then
        error "Config file not found at $CONFIG_PATH. Please create it first."
    fi
}

# === Linux (systemd) ===

systemd_install() {
    check_prereqs
    info "Installing systemd service..."
    
    cat << EOF | sudo tee "$SYSTEMD_DIR/${SERVICE_NAME}.service" > /dev/null
[Unit]
Description=whozere - Login Detection & Notification
After=network.target

[Service]
Type=simple
ExecStart=$BINARY_PATH -config $CONFIG_PATH
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    success "Systemd service installed and enabled"
    info "Start with: sudo systemctl start $SERVICE_NAME"
}

systemd_uninstall() {
    info "Uninstalling systemd service..."
    
    sudo systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    sudo systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    sudo rm -f "$SYSTEMD_DIR/${SERVICE_NAME}.service"
    sudo systemctl daemon-reload
    
    success "Systemd service uninstalled"
}

systemd_start() {
    sudo systemctl start "$SERVICE_NAME"
    success "Service started"
}

systemd_stop() {
    sudo systemctl stop "$SERVICE_NAME"
    success "Service stopped"
}

systemd_restart() {
    sudo systemctl restart "$SERVICE_NAME"
    success "Service restarted"
}

systemd_status() {
    systemctl status "$SERVICE_NAME" --no-pager || true
}

# === macOS (launchd) ===

PLIST_NAME="com.${SERVICE_NAME}.agent.plist"
PLIST_PATH="$LAUNCHD_DIR/$PLIST_NAME"

launchd_install() {
    check_prereqs
    info "Installing launchd service..."
    
    mkdir -p "$LAUNCHD_DIR"
    
    cat << EOF > "$PLIST_PATH"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.${SERVICE_NAME}.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
        <string>-config</string>
        <string>$CONFIG_PATH</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/${SERVICE_NAME}.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/${SERVICE_NAME}.err</string>
</dict>
</plist>
EOF

    success "Launchd plist installed at $PLIST_PATH"
    info "Start with: launchctl load $PLIST_PATH"
}

launchd_uninstall() {
    info "Uninstalling launchd service..."
    
    launchctl unload "$PLIST_PATH" 2>/dev/null || true
    rm -f "$PLIST_PATH"
    
    success "Launchd service uninstalled"
}

launchd_start() {
    launchctl load "$PLIST_PATH"
    success "Service started"
}

launchd_stop() {
    launchctl unload "$PLIST_PATH"
    success "Service stopped"
}

launchd_restart() {
    launchctl unload "$PLIST_PATH" 2>/dev/null || true
    launchctl load "$PLIST_PATH"
    success "Service restarted"
}

launchd_status() {
    if launchctl list | grep -q "$SERVICE_NAME"; then
        success "Service is running"
        launchctl list | grep "$SERVICE_NAME"
    else
        warn "Service is not running"
    fi
}

# === Main ===

usage() {
    echo "Usage: $0 [install|uninstall|start|stop|restart|status]"
    echo ""
    echo "Commands:"
    echo "  install    Install as system service (auto-start on boot)"
    echo "  uninstall  Remove system service"
    echo "  start      Start the service"
    echo "  stop       Stop the service"
    echo "  restart    Restart the service"
    echo "  status     Show service status"
    exit 1
}

main() {
    case "$OS" in
        linux)
            if ! command -v systemctl &> /dev/null; then
                error "systemctl not found. This script requires systemd."
            fi
            
            case "${1:-}" in
                install)   systemd_install ;;
                uninstall) systemd_uninstall ;;
                start)     systemd_start ;;
                stop)      systemd_stop ;;
                restart)   systemd_restart ;;
                status)    systemd_status ;;
                *)         usage ;;
            esac
            ;;
        darwin)
            case "${1:-}" in
                install)   launchd_install ;;
                uninstall) launchd_uninstall ;;
                start)     launchd_start ;;
                stop)      launchd_stop ;;
                restart)   launchd_restart ;;
                status)    launchd_status ;;
                *)         usage ;;
            esac
            ;;
        *)
            error "Unsupported OS: $OS. Windows users should use Task Scheduler manually."
            ;;
    esac
}

main "$@"
