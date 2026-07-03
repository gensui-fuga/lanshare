# 📁 LanShare — 全平台區域網路檔案共享

> 毛玻璃風格 · 自適應主題色 · iOS 7 相容 · Android 原生 · 零設定

## 🎯 一句話

開 App → 得網址 → 任何設備瀏覽器存取。同一區域網路，任何裝置都能上傳/下載檔案。

## ✨ 特色

| 特性 | 說明 |
|------|------|
| 🧊 **毛玻璃 UI** | `backdrop-filter: blur(24px)` 半透明面板 |
| 🖼️ **自訂背景** | 支援 JPG/PNG/GIF/WebP/BMP，上傳即套用 |
| 🎨 **自適應主題色** | 取背景主色，自動調整文字/卡片/強調色 |
| 📂 **上下載分離** | 上傳目錄 + 下載目錄獨立管理 |
| 📱 **App 內操作** | Android 原生 WebView，應用內完整操作 |
| 🌐 **雙模式** | App 內操作 + LAN 瀏覽器存取，同時運行 |
| 🔒 **純區域網路** | 不走外網，無帳號，無追蹤 |

## 📲 平台支援

| 平台 | 類型 | 大小 |
|------|------|------|
| **Android** | APK (WebView) | 2.6 MB |
| **Linux x64** | CLI 二進制 | 6.2 MB |
| **Linux arm64** | CLI 二進制 | 5.8 MB |
| **Windows x64** | CLI 二進制 | 6.4 MB |
| **macOS x64** | CLI 二進制 | 6.3 MB |
| **iPhone 4** | Safari Web | — |
| **任何瀏覽器** | Web UI | — |

## 🚀 使用方式

### Android (推薦)

```bash
# 下載 APK → 安裝 → 打開
adb install LanShare-1.1.0.apk
```

App 自動啟動伺服器，內建 WebView 顯示完整界面。點擊狀態列可複製網址或開瀏覽器。

### 桌面端

```bash
# Linux
./lanshare                           # 預設 port 8080，暫存目錄
./lanshare 9000                      # 自訂 port
./lanshare ./shared                  # 單一目錄（上下載共用）
./lanshare 9000 ./uploads ./downloads  # 分離上下載目錄

# macOS / Windows 同理
```

### Arch Linux

```bash
yay -S lanshare-bin
```

### iPhone 4 / iPad 2 (iOS 7)

Safari 開啟 Android 或桌面端顯示的網址 `http://IP:8080`

## 🔧 環境變數

```bash
LANSHARE_UPLOAD_DIR=/path/to/uploads
LANSHARE_DOWNLOAD_DIR=/path/to/downloads
LANSHARE_PORT=8080
```

## 🎮 Web UI 功能

- 📥 **下載區** — 瀏覽、下載、刪除檔案
- 📤 **上傳區** — 點擊或拖曳上傳，最大 4GB
- ⚙️ **設定** — 修改上下載路徑、上傳背景圖片
- 📊 **統計** — 檔案數量、總大小、運行時間

## 📦 技術棧

- **伺服器**: Go (CGO_ENABLED=0 靜態編譯)
- **Web UI**: 純 HTML/CSS/JS，無框架
- **Android**: WebView + 原生 Service
- **背景自適應**: Canvas API 取主色

## 🔨 自行編譯

```bash
git clone https://github.com/gensui-fuga/lanshare
cd lanshare

# 桌面端
go build -o lanshare .

# Android APK (需要 Android SDK build-tools)
./build/build-apk.sh
```
