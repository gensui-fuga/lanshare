#!/bin/bash
# Build LanShare APK for Android
set -e

VERSION="1.0.0"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ANDROID_DIR="$SCRIPT_DIR"

ANDROID_SDK="/opt/android-sdk"
ANDROID_NDK="/opt/android-ndk"
BUILD_TOOLS="$ANDROID_SDK/build-tools/36.1.0"

if [ ! -d "$BUILD_TOOLS" ]; then
    BUILD_TOOLS=$(ls -d "$ANDROID_SDK/build-tools/"* 2>/dev/null | head -1)
fi

echo "📦 Building LanShare APK v$VERSION"
echo "==================================="

# Step 1: Build Go binary for Android
echo ""
echo "1️⃣  Cross-compiling Go binary..."
cd "$PROJECT_DIR"

GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$ANDROID_DIR/assets/lanshare-android-arm64" .
echo "   ✓ Binary: $(ls -lh "$ANDROID_DIR/assets/lanshare-android-arm64" | awk '{print $5}')"

# Also compile for armv7a (older devices)
GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$ANDROID_DIR/assets/lanshare-android-armv7a" . 2>/dev/null || true

# Step 2: Compile Java sources to DEX
echo ""
echo "2️⃣  Compiling Java sources..."

# Create R.java
"$ANDROID_SDK/platform-tools/aapt" package -f \
    -M "$ANDROID_DIR/AndroidManifest.xml" \
    -S "$ANDROID_DIR/res" \
    -I "$ANDROID_SDK/platforms/android-34/android.jar" \
    -J "$ANDROID_DIR/src" \
    -m 2>&1 | grep -v "W/ResourceType" || true

# Compile Java to class files
javac -d "$ANDROID_DIR/build/classes" \
    -bootclasspath "$ANDROID_SDK/platforms/android-34/android.jar" \
    -source 1.8 -target 1.8 \
    "$ANDROID_DIR/src/com/lanshare/MainActivity.java" \
    "$ANDROID_DIR/src/com/lanshare/ServerService.java" \
    "$ANDROID_DIR/src/com/lanshare/R.java" 2>&1 || {
    echo "   ⚠ Java compilation issue, retrying without R..."
    javac -d "$ANDROID_DIR/build/classes" \
        -bootclasspath "$ANDROID_SDK/platforms/android-34/android.jar" \
        -source 1.8 -target 1.8 \
        "$ANDROID_DIR/src/com/lanshare/MainActivity.java" \
        "$ANDROID_DIR/src/com/lanshare/ServerService.java" 2>&1
}

# Convert to DEX
mkdir -p "$ANDROID_DIR/build"
"$BUILD_TOOLS/d8" --release \
    --lib "$ANDROID_SDK/platforms/android-34/android.jar" \
    --output "$ANDROID_DIR/build/" \
    "$ANDROID_DIR/build/classes/"*.class 2>&1

echo "   ✓ DEX created"

# Step 3: Package APK
echo ""
echo "3️⃣  Packaging APK..."

# Create APK with aapt
"$ANDROID_SDK/platform-tools/aapt" package -f \
    -M "$ANDROID_DIR/AndroidManifest.xml" \
    -S "$ANDROID_DIR/res" \
    -A "$ANDROID_DIR/assets" \
    -I "$ANDROID_SDK/platforms/android-34/android.jar" \
    -F "$ANDROID_DIR/build/lanshare-unsigned.apk" \
    2>&1 | grep -v "W/ResourceType" || true

# Add DEX
cd "$ANDROID_DIR/build"
"$BUILD_TOOLS/aapt" add "lanshare-unsigned.apk" "classes.dex" 2>&1
cd "$PROJECT_DIR"

echo "   ✓ APK packaged"

# Step 4: Sign APK
echo ""
echo "4️⃣  Signing APK..."

# Generate keystore if not exists
KEYSTORE="$ANDROID_DIR/build/lanshare.keystore"
if [ ! -f "$KEYSTORE" ]; then
    keytool -genkey -v -keystore "$KEYSTORE" \
        -alias lanshare -keyalg RSA -keysize 2048 -validity 10000 \
        -storepass lanshare123 -keypass lanshare123 \
        -dname "CN=LanShare, OU=Dev, O=LanShare, L=Unknown, ST=Unknown, C=CN" 2>&1
fi

# Sign
"$BUILD_TOOLS/apksigner" sign \
    --ks "$KEYSTORE" \
    --ks-pass pass:lanshare123 \
    --key-pass pass:lanshare123 \
    --out "$PROJECT_DIR/dist/android/LanShare-$VERSION.apk" \
    "$ANDROID_DIR/build/lanshare-unsigned.apk" 2>&1

echo ""
echo "==================================="
echo "✅ APK created!"
echo "📁 Output: dist/android/LanShare-$VERSION.apk"
ls -lh "$PROJECT_DIR/dist/android/LanShare-$VERSION.apk"
echo ""
echo "📱 Install: adb install dist/android/LanShare-$VERSION.apk"
