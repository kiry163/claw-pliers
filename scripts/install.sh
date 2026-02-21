#!/bin/bash
set -e

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.claw-pliers}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
PORT="${PORT:-8080}"

echo "Installing claw-pliers $VERSION..."

mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"

if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -sL https://api.github.com/repos/kiry163/claw-pliers/releases/latest | grep -oP '"tag_name": "\K[^"]+')
fi

ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

TAR_NAME="claw-pliers_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/kiry163/claw-pliers/releases/download/${VERSION}/${TAR_NAME}"

echo "Downloading $DOWNLOAD_URL..."
curl -fsSL "$DOWNLOAD_URL" -o "/tmp/claw-pliers.tar.gz"

echo "Extracting..."
tar -xzf "/tmp/claw-pliers.tar.gz" -C "$BIN_DIR" claw-pliers
chmod +x "$BIN_DIR/claw-pliers"

rm -f /tmp/claw-pliers.tar.gz

echo "Creating default config..."
cat > "$INSTALL_DIR/config.yaml" << EOF
server:
  port: $PORT
  log_level: info

auth:
  local_key: "$(openssl rand -hex 16)"

includes:
  - name: file
    path: "./file-config.yaml"
  - name: mail
    path: "./mail-config.yaml"
  - name: image
    path: "./image-config.yaml"
EOF

echo "Config created at $INSTALL_DIR/config.yaml"

echo ""
echo "Installation complete!"
echo "Run 'claw-pliers --help' to get started"
