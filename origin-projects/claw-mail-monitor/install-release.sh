#!/bin/bash

set -euo pipefail

REPO="kiry163/claw-mail-monitor"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TMP_DIR="${TMP_DIR:-/tmp}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *) echo "unsupported" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported" ;;
  esac
}

OS_NAME=$(detect_os)
ARCH_NAME=$(detect_arch)

if [ "$OS_NAME" = "unsupported" ] || [ "$ARCH_NAME" = "unsupported" ]; then
  echo "Unsupported platform: $(uname -s) $(uname -m)" >&2
  exit 1
fi

ASSET="claw-mail-monitor_${OS_NAME}_${ARCH_NAME}"
if [ "$VERSION" = "latest" ]; then
	BASE_URL="https://github.com/$REPO/releases/latest/download"
else
	BASE_URL="https://github.com/$REPO/releases/download/$VERSION"
fi
BIN_URL="$BASE_URL/$ASSET"
SUM_URL="$BASE_URL/checksums.txt"
BIN_PATH="$TMP_DIR/$ASSET"
SUM_PATH="$TMP_DIR/checksums.txt"

echo "Downloading $BIN_URL"
curl -fsSL -o "$BIN_PATH" "$BIN_URL"
curl -fsSL -o "$SUM_PATH" "$SUM_URL"

LINE=$(grep " $ASSET$" "$SUM_PATH" || true)
if [ -z "$LINE" ]; then
  echo "Checksum entry not found for $ASSET" >&2
  exit 1
fi

EXPECTED_SUM=${LINE%% *}
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL_SUM=$(sha256sum "$BIN_PATH" | cut -d ' ' -f1)
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL_SUM=$(shasum -a 256 "$BIN_PATH" | cut -d ' ' -f1)
else
  echo "sha256sum or shasum is required" >&2
  exit 1
fi

if [ "$EXPECTED_SUM" != "$ACTUAL_SUM" ]; then
  echo "Checksum mismatch" >&2
  exit 1
fi

if [ ! -d "$INSTALL_DIR" ]; then
  if command -v sudo >/dev/null 2>&1; then
    sudo mkdir -p "$INSTALL_DIR"
  else
    mkdir -p "$INSTALL_DIR"
  fi
fi

if command -v sudo >/dev/null 2>&1 && [ "$(id -u)" -ne 0 ]; then
  sudo install -m 0755 "$BIN_PATH" "$INSTALL_DIR/claw-mail-monitor"
else
  install -m 0755 "$BIN_PATH" "$INSTALL_DIR/claw-mail-monitor"
fi

echo "Installed to $INSTALL_DIR/claw-mail-monitor"
echo "Run: claw-mail-monitor serve --listen 127.0.0.1:14630"
