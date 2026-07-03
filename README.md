# LanShare 📁

> Cross-Platform Local Network File Sharing — **Works on iPhone 4 (iOS 7) too**

[![License: GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![AUR](https://img.shields.io/aur/version/lanshare-bin)](https://aur.archlinux.org/packages/lanshare-bin)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8)](https://go.dev/)

LanShare is a lightweight, zero-configuration file sharing tool that works entirely on your local network. **No internet required. No accounts. No cloud.**

## 🌟 Features

- **📱 iPhone 4 (iOS 7) Compatible** — Web UI works on old Safari, no app install needed
- **🤖 Standalone Android APK** — Native UI with file manager, no Termux required
- **🖥️ Cross-Platform** — Windows, Linux, macOS, Android APK, Android Termux, iOS Safari
- **📂 Built-in File Manager** — Upload, download, delete files from any device
- **⚙️ Configurable Storage** — Change save location from the settings panel
- **🌐 IP Display** — Shows your LAN IP at the top of every interface
- **🔒 Local Only** — Files never leave your LAN. No cloud, no tracking.
- **⚡ Zero Config** — Download, run, share. That's it.

## 🚀 Quick Start

### 🤖 Android (APK)

Download from [Releases](https://github.com/gensui-fuga/lanshare/releases):

```bash
# Install via ADB
adb install LanShare-1.0.0.apk
# Or just tap the APK on your phone
```

Open the app → tap **啟動伺服器** → open browser on any device → visit the shown IP.

### 🖥️ Desktop (Windows / Linux / macOS)

```bash
# Download from Releases, then:
./lanshare
# Open http://localhost:8080 in your browser
```

### 📱 iPhone 4 / iPad 2 / Any iOS Device

1. Run LanShare on **any computer** in your home
2. Open **Safari** on your iPhone 4
3. Visit `http://YOUR_COMPUTER_IP:8080`
4. Upload / download files instantly

> ✅ **Works on iOS 7.1.2** — no app installation, no jailbreak needed

### 🐧 Arch Linux

```bash
yay -S lanshare-bin
lanshare
```

## 📥 Downloads

Download from **[GitHub Releases](https://github.com/gensui-fuga/lanshare/releases)**:

| Platform | File | Notes |
|----------|------|-------|
| **🤖 Android** | `LanShare-1.0.0.apk` | Standalone APK, native UI, no Termux |
| **🐧 Linux x86_64** | `lanshare-linux-amd64` | Portable binary |
| **🐧 Linux ARM64** | (build from source) | Raspberry Pi, etc. |
| **🪟 Windows** | `lanshare-windows-amd64.exe` | Portable, no install |
| **🍎 macOS Intel** | `lanshare-darwin-amd64` | macOS 10.12+ |
| **🍎 macOS Apple Silicon** | `lanshare-darwin-arm64` | Native M1/M2/M3 |
| **📱 iOS** | — | Use Web UI via Safari |

### AUR

```bash
yay -S lanshare-bin
```

### Debian / Ubuntu

```bash
# From Releases, download and install:
sudo dpkg -i lanshare_1.0.0_amd64.deb
```

## 🖥️ Usage

### Basic

```bash
# Default (port 8080, temp directory)
lanshare

# Custom port
lanshare 9000

# Custom shared directory
lanshare ~/shared
```

### Environment Variables

```bash
LANSHARE_DIR=~/shared LANSHARE_PORT=9090 lanshare
LANSHARE_OPEN_BROWSER=1 lanshare  # Auto-open browser
```

### Settings via Web UI

Open `http://localhost:8080` in any browser:

| Tab | Features |
|-----|----------|
| 📄 **檔案** | File list, delete, auto-refresh |
| 📤 **上傳** | Drag & drop, progress bar, multi-file |
| ⚙️ **設定** | IP info, file count, uptime, storage path |

## 📱 Android APK Features

The standalone APK (`LanShare-1.0.0.apk`) provides:

```
┌──────────────────────────┐
│  📁 LanShare    ○ 已停止  │
├──────────────────────────┤
│  IP: 192.168.1.5  8080   │
├──────────────────────────┤
│  📄 檔案  │  ⚙️ 設定     │
├──────────────────────────┤
│  📂 檔案管理              │
│  🖼️ photo.jpg   1.2 MB   │
│  📄 report.pdf  2.4 MB   │
│  📦 backup.zip  156 MB   │
├──────────────────────────┤
│  [▶ 啟動伺服器]  [🌐 開啟]│
└──────────────────────────┘
```

- **No Termux required** — works standalone
- **Configurable storage** — change save path in settings
- **IP display** — shows your LAN IP at all times
- **Background service** — runs even when you switch apps
- **Notification** — shows server status in notification bar

## Web UI (iOS 7 / iPhone 4)

The Web UI is specifically designed for old browsers:

- **ES5 only** — no modern JavaScript, works on iOS 7 Safari
- **Table layout** — no flexbox, reliable on old rendering engines
- **Dark theme** — easy on the eyes, consistent look
- **Full functionality** — file manager, upload, settings, all from the browser

## 🔧 Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  iPhone 4   │     │   Android    │     │  Desktop    │
│  (Safari)   │     │  (APK / Web) │     │  (Native)   │
└──────┬──────┘     └──────┬───────┘     └──────┬──────┘
       │                   │                    │
       └───────────────────┼────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  LanShare   │
                    │  Server     │
                    │  (:8080)    │
                    └─────────────┘
                    Shared Directory
```

## 🏗️ Building from Source

```bash
git clone https://github.com/gensui-fuga/lanshare.git
cd lanshare

# Build for current platform
go build -o lanshare .

# Cross-compile all platforms
./build/build-all.sh

# Build Android APK (requires gomobile)
./build/build-apk.sh
```

### Prerequisites

- Go 1.21+
- For GUI: `fyne.io/fyne/v2` + system OpenGL
- For Android APK: `gomobile` + Android NDK

## 📦 Packages

| Package | Command |
|---------|---------|
| **Arch Linux (AUR)** | `yay -S lanshare-bin` |
| **Debian/Ubuntu** | `sudo dpkg -i lanshare_1.0.0_amd64.deb` |

## 🗺️ Roadmap

- [x] Cross-platform CLI server
- [x] Web UI (iOS 7 compatible)
- [x] Standalone Android APK with native UI
- [x] Desktop GUI (Fyne)
- [x] AUR package
- [ ] macOS .dmg installer
- [ ] Windows NSIS installer
- [ ] iOS native app (requires macOS + Xcode)
- [ ] mDNS auto-discovery

## 📄 License

GNU General Public License v3.0

## 🙏 Credits

- [Fyne](https://fyne.io/) — Desktop GUI toolkit
- Inspired by [LocalSend](https://github.com/localsend/localsend)

---

**Made for the iPhone 4 that refuses to die** 🫡
