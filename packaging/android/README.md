# LanShare on Android

LanShare runs on Android through **Termux** (terminal emulator). The Go binary is cross-compiled for Android's Linux kernel.

## Installation via Termux

### 1. Install Termux

Get it from [F-Droid](https://f-droid.org/packages/com.termux/) (recommended) or GitHub Releases.

### 2. Download the binary

```bash
# In Termux, for 64-bit devices:
pkg install wget
wget https://github.com/gensui-fuga/lanshare/releases/latest/download/lanshare-android-arm64

# For 32-bit (armv7a) devices:
wget https://github.com/gensui-fuga/lanshare/releases/latest/download/lanshare-android-armv7a
```

### 3. Make executable and run

```bash
chmod +x lanshare-android-arm64
mv lanshare-android-arm64 $PREFIX/bin/lanshare
lanshare
```

### 4. Access from other devices

Open browser on any device in the same network:
```
http://YOUR_ANDROID_IP:8080
```

## As a full APK (requires Android SDK)

For a native APK with GUI, you'd need:
- Android SDK + NDK
- gomobile

```bash
# Install gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Build APK
gomobile bind -target android -o lanshare.apk .
```

This produces a basic APK. For a full-featured app, you'd need to write an Android UI wrapper.

## Auto-start on boot (Termux:Boot)

1. Install `Termux:Boot` from F-Droid
2. Create `~/.termux/boot/lanshare.sh`:
   ```bash
   #!/data/data/com.termux/files/usr/bin/bash
   termux-wake-lock
   lanshare 8080 /data/data/com.termux/files/home/shared
   ```
3. Make executable: `chmod +x ~/.termux/boot/lanshare.sh`
