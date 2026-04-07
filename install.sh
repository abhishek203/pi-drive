#!/bin/bash
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
esac

GITHUB_REPO="abhishek203/pi-drive"
BINARY="pidrive-${OS}-${ARCH}"

# Get latest release tag
VERSION=$(curl -sSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest release version."
  exit 1
fi

DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY}"
CHECKSUMS_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/checksums.txt"

echo "Installing pidrive ${VERSION} for $OS/$ARCH..."

# Download binary from GitHub releases
curl -sSLo "/tmp/${BINARY}" "${DOWNLOAD_URL}"

# Verify checksum from GitHub releases (fail if unavailable)
echo "Verifying checksum..."
EXPECTED=$(curl -sSL "${CHECKSUMS_URL}" | grep "${BINARY}" | awk '{print $1}')
if [ -z "$EXPECTED" ]; then
  echo "Error: could not fetch or find checksum for ${BINARY}. Aborting."
  rm -f "/tmp/${BINARY}"
  exit 1
fi

if command -v sha256sum &>/dev/null; then
  ACTUAL=$(sha256sum "/tmp/${BINARY}" | awk '{print $1}')
else
  ACTUAL=$(shasum -a 256 "/tmp/${BINARY}" | awk '{print $1}')
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Checksum mismatch! Expected: $EXPECTED, Got: $ACTUAL"
  rm -f "/tmp/${BINARY}"
  exit 1
fi
echo "Checksum OK."

# Install binary to ~/.local/bin (no sudo needed)
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "${INSTALL_DIR}"
mv "/tmp/${BINARY}" "${INSTALL_DIR}/pidrive"
chmod +x "${INSTALL_DIR}/pidrive"

# Add to PATH if not already there
if ! echo "$PATH" | grep -q "${INSTALL_DIR}"; then
  echo "Add to your PATH: export PATH=\"\${HOME}/.local/bin:\$PATH\""
fi

# Check for davfs2 on Linux
if [ "$OS" = "linux" ]; then
  if ! command -v mount.davfs &>/dev/null; then
    echo ""
    echo "Note: davfs2 is needed for WebDAV mount support."
    echo "  Install it with: sudo apt install davfs2"
  fi
fi

echo ""
echo "pidrive installed!"
echo ""
if [ "$OS" = "darwin" ]; then
  MOUNT_PATH="$HOME/drive/my/"
else
  MOUNT_PATH="/drive/my/"
fi

echo "Next steps:"
echo "  pidrive register --email you@company.com --name \"My Agent\" --server https://pidrive.ressl.ai"
echo "  pidrive mount"
echo "  ls ${MOUNT_PATH}"
