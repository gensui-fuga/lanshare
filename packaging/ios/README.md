# LanShare on iOS

LanShare cannot be compiled as a native iOS app without macOS + Xcode.

However, the **Web UI** works perfectly on all iOS devices, including **iPhone 4 (iOS 7.1.2)**.

## Using on iPhone 4 / iPad 2 / iPhone 4s

1. Run LanShare on any desktop in your LAN:
   ```bash
   ./lanshare
   ```

2. Open Safari on your iOS device and visit:
   ```
   http://YOUR_DESKTOP_IP:8080
   ```

3. The Web UI is fully compatible with iOS 7 Safari:
   - Upload files from your device
   - Download shared files
   - No installation needed

## Certificate Fix for iOS 7

If you get SSL warnings (not applicable - LanShare uses plain HTTP):

Old iOS devices (iOS 9 and earlier) have an expired root certificate (DST Root CA X3) that causes HTTPS errors on most websites.

**Since LanShare uses HTTP, this is NOT an issue for our Web UI.** Just connect directly via HTTP.

## For Jailbroken Devices

If your iPhone 4 is jailbroken, you can run the Go binary directly:

1. Install `Terminal` and `Entropy` from Cydia (or any SSH server)
2. Transfer the Linux ARM binary to the device via SCP
3. Run `./lanshare` via SSH

## Building for iOS (requires macOS)

If you have a Mac, you can build a proper IPA:

```bash
# Prerequisites: Xcode + Rust
cargo install cargo-ipa
cargo ipa --target aarch64-apple-ios --release
```

This requires porting the Go code to Rust (or using `gomobile`), which is outside the scope of this project.

---

**Bottom line**: For iPhone 4 / old iOS devices, just use the Web UI via Safari. Zero installation required.
