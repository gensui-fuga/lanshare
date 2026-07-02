#!/bin/bash
# LanShare - Cross-platform build script
# Builds for Linux, Windows, macOS, Android

set -e

VERSION="1.0.0"
BUILD_DIR="$(dirname "$0")/../dist"
SOURCE_DIR="$(dirname "$0")/.."
mkdir -p "$BUILD_DIR"/{linux,windows,darwin,android}

echo "🚀 LanShare v${VERSION} Cross-Platform Build"
echo "==========================================="
echo ""

# Detect Go
if ! command -v go &>/dev/null; then
    echo "❌ Go not found. Please install Go 1.21+"
    exit 1
fi
echo "✓ Go $(go version | grep -oP 'go\S+')"

cd "$SOURCE_DIR"

# Common build flags
LDFLAGS="-s -w -X main.Version=${VERSION}"

# =============================================
# Linux (amd64 + arm64)
# =============================================
echo ""
echo "📦 Linux..."

# amd64
echo "   → amd64..."
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/linux/lanshare-linux-amd64" .
echo "     ✓ $(ls -lh "$BUILD_DIR/linux/lanshare-linux-amd64" | awk '{print $5}')"

# arm64
echo "   → arm64..."
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/linux/lanshare-linux-arm64" .
echo "     ✓ $(ls -lh "$BUILD_DIR/linux/lanshare-linux-arm64" | awk '{print $5}')"

# =============================================
# Windows (amd64)
# =============================================
echo ""
echo "📦 Windows..."

echo "   → amd64..."
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS -H windowsgui" -o "$BUILD_DIR/windows/lanshare.exe" .
echo "     ✓ $(ls -lh "$BUILD_DIR/windows/lanshare.exe" | awk '{print $5}')"

# =============================================
# macOS (amd64 + arm64)
# =============================================
echo ""
echo "📦 macOS..."

# amd64 (Intel)
echo "   → amd64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/darwin/lanshare-darwin-amd64" .
echo "     ✓ $(ls -lh "$BUILD_DIR/darwin/lanshare-darwin-amd64" | awk '{print $5}')"

# arm64 (Apple Silicon)
echo "   → arm64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/darwin/lanshare-darwin-arm64" .
echo "     ✓ $(ls -lh "$BUILD_DIR/darwin/lanshare-darwin-arm64" | awk '{print $5}')"

# Create macOS .app bundle
echo ""
echo "   → Creating .app bundle..."
APP_DIR="$BUILD_DIR/darwin/LanShare.app"
mkdir -p "$APP_DIR/Contents/MacOS"

# Use arm64 as default binary (universal would need lipo)
cp "$BUILD_DIR/darwin/lanshare-darwin-arm64" "$APP_DIR/Contents/MacOS/lanshare"

cat > "$APP_DIR/Contents/Info.plist" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>lanshare</string>
    <key>CFBundleIdentifier</key>
    <string>com.lanshare.app</string>
    <key>CFBundleName</key>
    <string>LanShare</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.12</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST

echo "     ✓ LanShare.app created"
echo "     (run: open $APP_DIR or copy to /Applications)"

# =============================================
# Android (Termux)
# =============================================
echo ""
echo "📦 Android (Termux)..."

# arm64
echo "   → arm64..."
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/android/lanshare-android-arm64" .
echo "     ✓ $(ls -lh "$BUILD_DIR/android/lanshare-android-arm64" | awk '{print $5}')"

# armv7a (for older devices)
echo "   → armv7a..."
GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/android/lanshare-android-armv7a" .
echo "     ✓ $(ls -lh "$BUILD_DIR/android/lanshare-android-armv7a" | awk '{print $5}')"

# x86_64 (emulator)
echo "   → x86_64 (emulator)..."
GOOS=android GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$LDFLAGS" -o "$BUILD_DIR/android/lanshare-android-x86_64" .
echo "     ✓ $(ls -lh "$BUILD_DIR/android/lanshare-android-x86_64" | awk '{print $5}')"

# =============================================
# Summary
# =============================================
echo ""
echo "============================================"
echo "✅ Build complete!"
echo ""
echo "📁 Output:"
echo "   Linux:"
echo "     - dist/linux/lanshare-linux-amd64"
echo "     - dist/linux/lanshare-linux-arm64"
echo "   Windows:"
echo "     - dist/windows/lanshare.exe"
echo "   macOS:"
echo "     - dist/darwin/lanshare-darwin-amd64"
echo "     - dist/darwin/lanshare-darwin-arm64"
echo "     - dist/darwin/LanShare.app"
echo "   Android (Termux):"
echo "     - dist/android/lanshare-android-arm64"
echo "     - dist/android/lanshare-android-armv7a"
echo "     - dist/android/lanshare-android-x86_64"
echo ""
echo "📱 iOS: 使用 Web UI (所有設備瀏覽器訪問)"
echo "   iPhone 4 (iOS 7) 完全相容"
echo "============================================"
