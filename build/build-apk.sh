#!/bin/bash
# Build LanShare APK using gomobile
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
MOBILE_DIR="$SCRIPT_DIR/mobile-apk"

echo "📦 Building LanShare APK (gomobile)"

# Create mobile app directory
rm -rf "$MOBILE_DIR"
mkdir -p "$MOBILE_DIR"

# Write the mobile app
cat > "$MOBILE_DIR/main.go" << 'EOF'
package main

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/geom"
)

var (
	serverPort = 8080
	serverIP   = "127.0.0.1"
	logLines   []string
	appCtx     *app.App
)

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func addLog(msg string) {
	logLines = append(logLines, msg)
	if len(logLines) > 20 {
		logLines = logLines[len(logLines)-20:]
	}
}

func main() {
	// Start HTTP server
	go func() {
		shareDir := filepath.Join(os.TempDir(), "lanshare")
		os.MkdirAll(shareDir, 0755)
		serverIP = getLocalIP()

		// Quick server implementation
		mux := http.NewServeMux()
		mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"version": "1.0.0",
				"ip": serverIP,
				"port": serverPort,
				"share_dir": shareDir,
				"status": "running",
			})
		})
		// ... serve files from shareDir

		ln, err := net.Listen("tcp", ":"+strconv.Itoa(serverPort))
		if err != nil {
			addLog("ERROR: " + err.Error())
			return
		}
		serverPort = ln.Addr().(*net.TCPAddr).Port
		addLog("LanShare: http://" + serverIP + ":" + strconv.Itoa(serverPort))
		http.Serve(ln, mux)
	}()

	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		appCtx = &a

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				if e.Crosses(lifecycle.StageVisible) == lifecycle.CrossOn {
					glctx, _ = e.DrawContext.(gl.Context)
					a.Send(paint.Event{})
				} else {
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				if glctx == nil || sz.WidthPx <= 0 || sz.HeightPx <= 0 {
					a.Send(paint.Event{})
					continue
				}
				render(glctx, sz)
				a.Publish()
				a.Send(paint.Event{})
			}
		}
	})
}
EOF

# Initialize module
cd "$MOBILE_DIR"
go mod init lanshare-mobile
go mod tidy 2>&1 | tail -3

# Build APK
export PATH="$HOME/go/bin:$PATH"
gomobile build -target android -androidapi 21 -o "$PROJECT_DIR/dist/android/LanShare-1.0.0.apk" . 2>&1

echo ""
echo "✅ APK built!"
ls -lh "$PROJECT_DIR/dist/android/LanShare-1.0.0.apk"
