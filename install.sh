#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="Sergeydigl3/zapret-nix"
TEMP_DIR=$(mktemp -d)

# Cleanup on exit
cleanup() {
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Print colored output
info() {
  echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
  echo -e "${RED}[ERROR]${NC} $1" >&2
}

warn() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

# Check if running with sudo
check_permissions() {
  if [[ $EUID -ne 0 ]]; then
    error "This script needs to be run with sudo"
    exit 1
  fi
}

# Detect system architecture
detect_arch() {
  local machine
  machine=$(uname -m)

  case "$machine" in
    x86_64)
      echo "amd64"
      ;;
    aarch64|arm64)
      echo "arm64"
      ;;
    armv7l)
      echo "armhf"
      ;;
    armv6l)
      echo "armhf"
      ;;
    i386|i686)
      echo "i386"
      ;;
    *)
      error "Unsupported architecture: $machine"
      exit 1
      ;;
  esac
}

# Detect package manager and distro
detect_package_manager() {
  if command -v apt-get &> /dev/null; then
    echo "deb"
  elif command -v dnf &> /dev/null; then
    echo "rpm"
  elif command -v yum &> /dev/null; then
    echo "rpm"
  elif command -v apk &> /dev/null; then
    echo "apk"
  elif command -v pacman &> /dev/null; then
    echo "archlinux"
  else
    error "Could not detect package manager"
    exit 1
  fi
}

# Get latest release version
get_latest_version() {
  local version
  version=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')

  if [[ -z "$version" ]]; then
    error "Failed to fetch latest release version"
    exit 1
  fi

  echo "$version"
}

# Download release package
download_package() {
  local version=$1
  local pm=$2
  local arch=$3
  local filename extension url

  case "$pm" in
    deb)
      extension="deb"
      filename="zapret-nix_${version}_${arch}.deb"
      ;;
    rpm)
      extension="rpm"
      filename="zapret-nix-${version}-1.${arch}.rpm"
      ;;
    apk)
      extension="apk"
      filename="zapret-nix-${version}-r0.${arch}.apk"
      ;;
    archlinux)
      extension="pkg.tar.zst"
      filename="zapret-nix-${version}-1-${arch}.pkg.tar.zst"
      ;;
    *)
      error "Unsupported package manager: $pm"
      exit 1
      ;;
  esac

  url="https://github.com/$REPO/releases/download/v${version}/$filename"

  info "Downloading zapret-nix v$version ($pm package)..."

  if ! curl -fL "$url" -o "$TEMP_DIR/$filename"; then
    error "Failed to download package from $url"
    exit 1
  fi

  echo "$TEMP_DIR/$filename"
}

# Install package
install_package() {
  local package=$1
  local pm=$2

  info "Installing package..."

  case "$pm" in
    deb)
      apt-get update
      dpkg -i "$package" || apt-get install -f -y
      ;;
    rpm)
      rpm -ivh "$package" || dnf install -y "$package" || yum install -y "$package"
      ;;
    apk)
      apk add --allow-untrusted "$package"
      ;;
    archlinux)
      pacman -U --noconfirm "$package"
      ;;
  esac
}

# Verify installation
verify_installation() {
  info "Verifying installation..."

  if command -v zapret &> /dev/null; then
    local version
    version=$(zapret --version 2>/dev/null || echo "unknown")
    info "Installation successful! Version: $version"
    return 0
  else
    warn "Could not verify installation"
    return 1
  fi
}

# Main installation
main() {
  echo "=========================================="
  echo "   zapret-nix Installer"
  echo "=========================================="
  echo ""

  check_permissions

  local arch pm version package

  pm=$(detect_package_manager)
  arch=$(detect_arch)

  info "Detected system: $pm/$arch"

  version=$(get_latest_version)
  info "Latest version: v$version"

  package=$(download_package "$version" "$pm" "$arch")

  install_package "$package" "$pm"

  verify_installation

  echo ""
  info "zapret-nix installed successfully!"
  echo "Run 'zapret --help' to get started"
}

main
