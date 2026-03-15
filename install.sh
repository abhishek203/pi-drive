#!/bin/bash
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
esac

GITHUB_REPO="ResslAI-Salesforce/pi-drive"
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

# Install binary (requires sudo for /usr/local/bin)
sudo mv "/tmp/${BINARY}" /usr/local/bin/pidrive
sudo chmod +x /usr/local/bin/pidrive

# Install davfs2 on Linux (needed for WebDAV mount)
if [ "$OS" = "linux" ]; then
  if ! command -v mount.davfs &>/dev/null; then
    echo "Installing davfs2 (WebDAV mount support)..."
    if command -v apt &>/dev/null; then
      sudo apt update -qq && sudo apt install -y -qq davfs2
    elif command -v yum &>/dev/null; then
      sudo yum install -y davfs2
    else
      echo "Please install davfs2 manually for WebDAV mount support."
    fi
  fi
fi

echo ""
echo "pidrive installed!"
echo ""
echo "Next steps:"
echo "  pidrive register --email you@company.com --name \"My Agent\" --server https://pidrive.ressl.ai"
echo "  pidrive mount"
echo "  ls /drive/my/"
