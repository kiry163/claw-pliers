#!/bin/bash

set -e

BIN_PATH="./claw-mail-monitor"
CONFIG_PATH=""
UPGRADE_ONLY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bin)
      BIN_PATH="$2"
      shift 2
      ;;
    --config)
      CONFIG_PATH="$2"
      shift 2
      ;;
    --upgrade)
      UPGRADE_ONLY=1
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

if [ ! -f "$BIN_PATH" ]; then
  echo "Binary $BIN_PATH not found"
  exit 1
fi

OS_NAME=$(uname -s)
TARGET_USER="${SUDO_USER:-root}"
HOME_DIR=""

if [ "$OS_NAME" = "Darwin" ]; then
  if [ "$TARGET_USER" = "root" ]; then
    HOME_DIR="/var/root"
  else
    HOME_DIR=$(dscl . -read "/Users/$TARGET_USER" NFSHomeDirectory | awk '{print $2}')
  fi
else
  if [ "$TARGET_USER" = "root" ]; then
    HOME_DIR="/root"
  else
    HOME_DIR=$(getent passwd "$TARGET_USER" | cut -d: -f6)
    if [ -z "$HOME_DIR" ]; then
      HOME_DIR="/home/$TARGET_USER"
    fi
  fi
fi

if [ -n "$CONFIG_PATH" ]; then
  CONFIG_DIR=$(dirname "$CONFIG_PATH")
else
  CONFIG_DIR="$HOME_DIR/.config/claw-mail-monitor"
  CONFIG_PATH="$CONFIG_DIR/config.yaml"
fi

CACHE_DIR="$HOME_DIR/.cache/claw-mail-monitor"

mkdir -p "$CONFIG_DIR"
mkdir -p "$CACHE_DIR"

if [ "$UPGRADE_ONLY" -eq 0 ]; then
  if [ ! -f "$CONFIG_PATH" ]; then
    if [ -f "./config.example.yaml" ]; then
      cp ./config.example.yaml "$CONFIG_PATH"
    fi
  fi
fi

if [ -f "$CONFIG_PATH" ]; then
  chmod 644 "$CONFIG_PATH"
fi

chmod 755 "$(dirname "$CONFIG_DIR")" || true
chmod 755 "$CONFIG_DIR" || true

cp "$BIN_PATH" /usr/local/bin/claw-mail-monitor
chmod +x /usr/local/bin/claw-mail-monitor

echo "Cleaning existing services..."
if command -v /usr/local/bin/claw-mail-monitor >/dev/null 2>&1; then
  CLAW_MAIL_MONITOR_CONFIG="$CONFIG_PATH" /usr/local/bin/claw-mail-monitor service uninstall >/dev/null 2>&1 || true
fi
if [ -f "$HOME_DIR/Library/LaunchAgents/claw-mail-monitor.plist" ]; then
  rm -f "$HOME_DIR/Library/LaunchAgents/claw-mail-monitor.plist"
fi

echo "Installing service..."
CLAW_MAIL_MONITOR_CONFIG="$CONFIG_PATH" /usr/local/bin/claw-mail-monitor service install
CLAW_MAIL_MONITOR_CONFIG="$CONFIG_PATH" /usr/local/bin/claw-mail-monitor service start

echo "Service status:"
CLAW_MAIL_MONITOR_CONFIG="$CONFIG_PATH" /usr/local/bin/claw-mail-monitor service status || true

if [ "$OS_NAME" = "Darwin" ]; then
  echo "launchctl status:"
  launchctl print system/claw-mail-monitor 2>/dev/null || true
fi

echo "Installation complete!"
echo "Config: $CONFIG_PATH"
