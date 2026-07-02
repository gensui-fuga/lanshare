#!/bin/bash
# LanShare Debian/Ubuntu package builder
set -e

VERSION="1.0.0"
ARCH="amd64"
PKG_NAME="lanshare_${VERSION}_${ARCH}"

SCRIPT_DIR="$(dirname "$0")"
SOURCE_DIR="$(dirname "$SCRIPT_DIR")/.."
BUILD_DIR="$SOURCE_DIR/build/dist"
DEB_DIR="$SCRIPT_DIR"

echo "📦 Building .deb package..."

# Build binary
cd "$SOURCE_DIR"
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$BUILD_DIR/lanshare" .

# Create package structure
PKG_ROOT="$SOURCE_DIR/dist/packaging/lanshare"
rm -rf "$PKG_ROOT"
mkdir -p "$PKG_ROOT/DEBIAN"
mkdir -p "$PKG_ROOT/usr/bin"
mkdir -p "$PKG_ROOT/usr/share/applications"
mkdir -p "$PKG_ROOT/usr/share/icons/hicolor/256x256/apps"
mkdir -p "$PKG_ROOT/usr/share/doc/lanshare"

# Copy control file
cp "$DEB_DIR/DEBIAN/control" "$PKG_ROOT/DEBIAN/"

# Install binary
cp "$BUILD_DIR/lanshare" "$PKG_ROOT/usr/bin/lanshare"
chmod 755 "$PKG_ROOT/usr/bin/lanshare"

# Desktop entry
cat > "$PKG_ROOT/usr/share/applications/lanshare.desktop" << 'DESKTOP'
[Desktop Entry]
Name=LanShare
Comment=Cross-Platform Local Network File Sharing
Exec=lanshare
Icon=lanshare
Terminal=false
Type=Application
Categories=Network;FileTransfer;
StartupNotify=true
DESKTOP

# Icon (placeholder - generates SVG icon embedded as PNG)
# For now create a simple icon
convert -size 256x256 xc:'#7c3aed' -fill white -font Helvetica -pointsize 100 -gravity center -annotate 0 'LS' "$PKG_ROOT/usr/share/icons/hicolor/256x256/apps/lanshare.png" 2>/dev/null || true

# Changelog
cat > "$PKG_ROOT/usr/share/doc/lanshare/changelog" << 'CHLOG'
lanshare (1.0.0) stable; urgency=medium

  * Initial release
  * Cross-platform LAN file sharing
  * iOS 7 (iPhone 4) compatible web UI
  * Desktop GUI with Fyne

 -- gensui-fuga <gensui-fuga@users.noreply.github.com>  Tue, 30 Jun 2026 23:00:00 +0800
CHLOG

gzip -9 -n "$PKG_ROOT/usr/share/doc/lanshare/changelog" 2>/dev/null || true

# Build .deb
dpkg-deb --build "$PKG_ROOT" "$SOURCE_DIR/dist/packaging/lanshare_${VERSION}_${ARCH}.deb"

echo "✅ .deb package created: dist/packaging/lanshare_${VERSION}_${ARCH}.deb"
