# LanShare 📁

> Cross-Platform Local Network File Sharing — **Works on iPhone 4 (iOS 7) too**

[![Build Status](https://github.com/gensui-fuga/lanshare/actions/workflows/release.yml/badge.svg)](https://github.com/gensui-fuga/lanshare/actions)
[![License: GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8)](https://go.dev/)

LanShare is a lightweight, zero-configuration file sharing tool that works entirely on your local network. **No internet required. No accounts. No cloud.**

## 🌟 Features

- **📱 iPhone 4 (iOS 7) Compatible** — Specially designed Web UI works on old Safari
- **🖥️ Cross-Platform** — Windows, Linux, macOS, Android (Termux), iOS (Safari)
- **🎨 Desktop GUI** — Native system tray app with Fyne (optional)
- **📤 Drag & Drop** — Upload files from any device
- **🔒 Local Only** — Files never leave your LAN. No data leaves your network.
- **⚡ Zero Config** — Download, run, share. That's it.
- **🌐 Web UI** — No installation needed on client devices

## 🚀 Quick Start

### Desktop (Windows / Linux / macOS)

```bash
# Download from Releases, then:
./lanshare

# Open http://localhost:8080 in your browser
# Or share the LAN IP with other devices
```

### iPhone 4 / iPad 2 / Any iOS Device

1. Run LanShare on **any computer** in your home
2. Open **Safari** on your iPhone 4
3. Visit `http://YOUR_COMPUTER_IP:8080`
4. Upload / download files instantly

> ✅ **Works on iOS 7.1.2** — no app installation, no jailbreak needed

### Android (Termux)

```bash
pkg install wget
wget https://github.com/gensui-fuga/lanshare/releases/latest/download/lanshare-android-arm64
chmod +x lanshare-android-arm64
mv lanshare-android-arm64 $PREFIX/bin/lanshare
lanshare
```

## 📦 Downloads

| Platform | Format | Notes |
|----------|--------|-------|
| **Linux x86_64** | Binary | Portable, no dependencies |
| **Linux ARM64** | Binary | Raspberry Pi, etc. |
| **Linux** | `.deb` | Debian / Ubuntu / Mint |
| **Linux** | PKGBUILD | Arch Linux (AUR soon) |
| **Windows** | `.exe` | Portable, no install needed |
| **macOS (Intel)** | Binary | macOS 10.12+ |
| **macOS (Apple Silicon)** | Binary | Native M1/M2/M3 |
| **macOS** | `.app` | Drag to Applications |
| **Android** | Binary | Termux required |
| **iOS** | 🌐 Web UI | Safari on any iOS version |

## 🖥️ Usage

### Basic

```bash
# Default (port 8080, temp directory)
lanshare

# Custom port
lanshare 9000

# Custom shared directory
lanshare ~/shared

# Both
lanshare 9000 ~/shared
```

### Environment Variables

```bash
LANSHARE_DIR=~/shared LANSHARE_PORT=9090 lanshare
LANSHARE_OPEN_BROWSER=1 lanshare  # Auto-open browser
```

### Desktop GUI (Fyne)

```bash
# Build with GUI support
go build -tags gui -o lanshare-gui .
./lanshare-gui
```

The GUI version provides:
- System tray icon with context menu
- File list with auto-refresh
- Server status display
- One-click "Open in Browser"
- Copy LAN URL to clipboard

## 🐧 Arch Linux

```bash
# From AUR (once published)
yay -S lanshare

# Or build from PKGBUILD
cd packaging/arch
makepkg -si
```

## 📱 Screenshots

```
📱 iPhone 4 (iOS 7.1.2) Web UI:

┌────────────────────────┐
│   📁 LanShare          │
├────────────────────────┤
│ 📤 上傳檔案             │
│ [選擇檔案] [上傳]       │
├────────────────────────┤
│ 📥 檔案列表             │
│ report.pdf    2.4 MB   │
│ photo.jpg     1.2 MB   │
│ backup.zip    156 MB   │
└────────────────────────┘
```

## 🏗️ Building from Source

```bash
git clone https://github.com/gensui-fuga/lanshare.git
cd lanshare

# Build for current platform
go build -o lanshare .

# Cross-compile all platforms
./build/build-all.sh

# Build .deb package
./build/build-deb.sh
```

### Prerequisites

- Go 1.21+
- For GUI: `fyne.io/fyne/v2` + system OpenGL libraries
- For .deb: `dpkg-deb`
- For Windows installer: NSIS (optional)

## 🔧 Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  iPhone 4   │     │   Android    │     │  Desktop    │
│  (Safari)   │     │  (Termux)    │     │  (Native)   │
└──────┬──────┘     └──────┬───────┘     └──────┬──────┘
       │                   │                    │
       └───────────────────┼────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  LanShare   │
                    │  Server     │
                    │ (:8080)     │
                    └─────────────┘
                    Shared Directory
```

## 🤝 Contributing

PRs welcome! See [issues](https://github.com/gensui-fuga/lanshare/issues) for ideas.

## 📄 License

GNU General Public License v3.0

## 🙏 Credits

- [Fyne](https://fyne.io/) - GUI toolkit
- Inspired by [LocalSend](https://github.com/localsend/localsend)

---

**Made for the iPhone 4 that refuses to die** 🫡
