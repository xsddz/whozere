#!/bin/bash
#
# whozere - One-line installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/xsddz/whozere/main/install.sh | bash
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="xsddz/whozere"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/usr/local/etc/whozere"

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

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    case $OS in
        linux|darwin)
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported OS: $OS"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/' || echo "")
    
    if [ -z "$VERSION" ]; then
        warn "Could not fetch latest version, using 'latest'"
        VERSION="latest"
    fi
    
    info "Latest version: $VERSION"
}

# Download and install
install() {
    local BINARY="whozere"
    local DOWNLOAD_URL

    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/whozere-${PLATFORM}"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/whozere-${PLATFORM}"
    fi

    if [ "$OS" = "windows" ]; then
        BINARY="whozere.exe"
        DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
    fi

    info "Downloading from: $DOWNLOAD_URL"
    
    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download binary
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY" || error "Download failed"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$BINARY" || error "Download failed"
    else
        error "Neither curl nor wget found"
    fi

    # Make executable
    chmod +x "$TMP_DIR/$BINARY"

    # Install binary
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
    else
        info "Need sudo to install to $INSTALL_DIR"
        sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
    fi

    success "Binary installed to $INSTALL_DIR/$BINARY"
}

# Install example config
install_config() {
    local CONFIG_URL="https://raw.githubusercontent.com/${REPO}/main/config.example.yaml"
    
    # Create config directory
    if [ ! -d "$CONFIG_DIR" ]; then
        if [ -w "$(dirname $CONFIG_DIR)" ]; then
            mkdir -p "$CONFIG_DIR"
        else
            sudo mkdir -p "$CONFIG_DIR"
        fi
    fi

    # Download example config
    local CONFIG_FILE="$CONFIG_DIR/config.example.yaml"
    if [ -w "$CONFIG_DIR" ]; then
        curl -fsSL "$CONFIG_URL" -o "$CONFIG_FILE" 2>/dev/null || warn "Could not download example config"
    else
        sudo curl -fsSL "$CONFIG_URL" -o "$CONFIG_FILE" 2>/dev/null || warn "Could not download example config"
    fi

    if [ -f "$CONFIG_FILE" ]; then
        success "Example config installed to $CONFIG_FILE"
    fi
}

# Verify installation
verify() {
    if command -v whozere &> /dev/null; then
        success "Installation complete!"
        echo ""
        whozere -version
        echo ""
        info "Next steps:"
        echo "  1. Copy and edit config: cp $CONFIG_DIR/config.example.yaml $CONFIG_DIR/config.yaml"
        echo "  2. Configure your notification channels in config.yaml"
        echo "  3. Test: whozere -config $CONFIG_DIR/config.yaml -test"
        echo "  4. Run: whozere -config $CONFIG_DIR/config.yaml"
    else
        error "Installation failed - whozere not found in PATH"
    fi
}

# Main
main() {
    echo ""
    echo "  ╦ ╦╦ ╦╔═╗╔═╗╔═╗╦═╗╔═╗"
    echo "  ║║║╠═╣║ ║╔═╝║╣ ╠╦╝║╣ "
    echo "  ╚╩╝╩ ╩╚═╝╚═╝╚═╝╩╚═╚═╝"
    echo "  Who's here? - Login Detection & Notification"
    echo ""

    detect_platform
    get_latest_version
    install
    install_config
    verify
}

main "$@"
