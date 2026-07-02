package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/gl"
)

// ====== Embedded Server ======
func startServer() {
	shareDir := filepath.Join(os.TempDir(), "lanshare")
	os.MkdirAll(shareDir, 0755)

	mux := http.NewServeMux()

	// Info endpoint
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"version":"1.0.0","ip":"%s","port":%d,"share_dir":"%s","status":"running"}`,
			getLocalIP(), portNum, shareDir)
	})

	// File list
	mux.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		entries, _ := os.ReadDir(shareDir)
		w.Write([]byte("["))
		for i, e := range entries {
			if i > 0 {
				w.Write([]byte(","))
			}
			info, _ := e.Info()
			fmt.Fprintf(w, `{"name":"%s","size":%d}`, e.Name(), info.Size())
		}
		w.Write([]byte("]"))
	})

	// Upload
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer file.Close()
		dst, _ := os.Create(filepath.Join(shareDir, header.Filename))
		defer dst.Close()
		buf := make([]byte, 32*1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				dst.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		fmt.Fprintf(w, `{"uploaded":["%s"]}`, header.Filename)
		fileCount++
	})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portNum))
	if err != nil {
		ln, _ = net.Listen("tcp", ":0")
	}
	portNum = ln.Addr().(*net.TCPAddr).Port
	serverIP = getLocalIP()
	serverOK = true

	log.Printf("LanShare server: http://%s:%d", serverIP, portNum)
	http.Serve(ln, mux)
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}

// ====== Global state for server ======
var (
	serverOK  bool
	serverIP  = "127.0.0.1"
	portNum   = 8080
	fileCount int
	startTime = time.Now()
)

// ====== OpenGL Rendering ======
// Simple ASCII bitmap font (8x8 pixels) for rendering text
var fontBitmap = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // space
	0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x00, // !
	0x66, 0x66, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, // "
	0x14, 0x14, 0x7E, 0x28, 0x7E, 0x50, 0x50, 0x00, // #
	0x3C, 0x52, 0x50, 0x3C, 0x0A, 0x4A, 0x3C, 0x00, // $
	0x62, 0x64, 0x08, 0x10, 0x26, 0x46, 0x00, 0x00, // %
	0x18, 0x24, 0x18, 0x28, 0x44, 0x44, 0x38, 0x00, // &
	0x18, 0x18, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, // '
	0x08, 0x10, 0x20, 0x20, 0x20, 0x10, 0x08, 0x00, // (
	0x10, 0x08, 0x04, 0x04, 0x04, 0x08, 0x10, 0x00, // )
	0x00, 0x24, 0x18, 0x7E, 0x18, 0x24, 0x00, 0x00, // *
	0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00, // +
	0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x08, 0x00, // ,
	0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00, // -
	0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00, // .
	0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x00, // /
	0x3C, 0x42, 0x46, 0x5A, 0x62, 0x42, 0x3C, 0x00, // 0
	0x08, 0x18, 0x28, 0x08, 0x08, 0x08, 0x3E, 0x00, // 1
	0x3C, 0x42, 0x02, 0x0C, 0x30, 0x40, 0x7E, 0x00, // 2
	0x3C, 0x42, 0x02, 0x1C, 0x02, 0x42, 0x3C, 0x00, // 3
	0x0C, 0x14, 0x24, 0x44, 0x7E, 0x04, 0x04, 0x00, // 4
	0x7E, 0x40, 0x7C, 0x02, 0x02, 0x42, 0x3C, 0x00, // 5
	0x1C, 0x20, 0x40, 0x7C, 0x42, 0x42, 0x3C, 0x00, // 6
	0x7E, 0x02, 0x04, 0x08, 0x10, 0x20, 0x20, 0x00, // 7
	0x3C, 0x42, 0x42, 0x3C, 0x42, 0x42, 0x3C, 0x00, // 8
	0x3C, 0x42, 0x42, 0x3E, 0x02, 0x04, 0x38, 0x00, // 9
	0x00, 0x18, 0x18, 0x00, 0x18, 0x18, 0x00, 0x00, // :
	0x00, 0x18, 0x18, 0x00, 0x18, 0x18, 0x08, 0x00, // ;
	0x04, 0x08, 0x10, 0x20, 0x10, 0x08, 0x04, 0x00, // <
	0x00, 0x00, 0x7E, 0x00, 0x7E, 0x00, 0x00, 0x00, // =
	0x20, 0x10, 0x08, 0x04, 0x08, 0x10, 0x20, 0x00, // >
	0x3C, 0x42, 0x02, 0x0C, 0x10, 0x00, 0x10, 0x00, // ?
}

func getCharIndex(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0' + 14) // digits start at index 14
	}
	if c >= 'A' && c <= 'Z' {
		return int(c - 'A' + 24)
	}
	if c >= 'a' && c <= 'z' {
		return int(c - 'a' + 24) // same as uppercase for simplicity
	}
	switch c {
	case ':': return 24
	case '.': return 12
	case '/': return 13
	case ' ': return 0
	case 'h': return 32
	case 't': return 46
	case 'p': return 42
	case 's': return 45
	case 'n': return 40
	case 'r': return 44
	case 'u': return 47
	case 'e': return 31
	case 'v': return 48
	case 'i': return 35
	case 'o': return 41
	}
	return 0
}

// Simple bitmap character renderer
func renderChar(glctx gl.Context, x, y float32, scale float32, c byte) {
	idx := getCharIndex(c)
	if idx >= len(fontBitmap)/8 {
		return
	}
	cellSize := scale * 8
	base := idx * 8
	for row := 0; row < 8; row++ {
		b := fontBitmap[base+row]
		for col := 0; col < 8; col++ {
			if b&(1<<(7-col)) != 0 {
				px := x + float32(col)*scale
				py := y + float32(row)*scale
				drawRect(glctx, px, py, scale, scale)
			}
		}
	}
}

func drawRect(glctx gl.Context, x, y, w, h float32) {
	// Using immediate mode-like approach
	vertices := []float32{
		x, y,
		x + w, y,
		x, y + h,
		x + w, y + h,
	}
	glctx.VertexAttrib2f(0, vertices[0], vertices[1])
	glctx.VertexAttrib2f(0, vertices[2], vertices[3])
	glctx.VertexAttrib2f(0, vertices[4], vertices[5])
	glctx.VertexAttrib2f(0, vertices[2], vertices[3])
	glctx.VertexAttrib2f(0, vertices[4], vertices[5])
	glctx.VertexAttrib2f(0, vertices[6], vertices[7])
	glctx.DrawArrays(gl.TRIANGLES, 0, 6)
}

func renderString(glctx gl.Context, x, y float32, scale float32, s string) {
	curX := x
	for i := 0; i < len(s); i++ {
		renderChar(glctx, curX, y, scale, s[i])
		curX += scale * 8
	}
}

func renderScene(glctx gl.Context, w, h float32) {
	// Dark background
	glctx.ClearColor(0.08, 0.08, 0.16, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	// Scale based on screen size
	scale := h / 64.0
	if scale < 2 {
		scale = 2
	}

	// Title
	title := "LanShare"
	if serverOK {
		title = "LanShare [RUNNING]"
	}
	x := w/2 - float32(len(title))*scale*4
	glctx.Uniform4f(1, 0.65, 0.48, 0.98, 1.0) // #a78bfa
	renderString(glctx, x, h*0.15, scale*1.2, title)

	// URL
	if serverOK {
		url := fmt.Sprintf("http://%s:%d", serverIP, portNum)
		x = w/2 - float32(len(url))*scale*4
		glctx.Uniform4f(1, 0.38, 0.65, 0.98, 1.0) // #60a5fa
		renderString(glctx, x, h*0.35, scale, url)
	} else {
		msg := "Starting server..."
		x = w/2 - float32(len(msg))*scale*4
		glctx.Uniform4f(1, 0.6, 0.72, 0.8, 1.0)
		renderString(glctx, x, h*0.35, scale, msg)
	}

	// Instructions
	lines := []string{
		"Open browser on any device",
		"and visit the URL above",
		"",
		"iPhone 4 / iOS 7 compatible",
	}
	for i, line := range lines {
		x := w/2 - float32(len(line))*scale*4
		glctx.Uniform4f(1, 0.6, 0.6, 0.7, 1.0)
		renderString(glctx, x, h*0.5+float32(i)*scale*10, scale*0.75, line)
	}

	// Status line
	uptime := time.Since(startTime).Round(time.Second).String()
	status := fmt.Sprintf("Files: %d  Uptime: %s", fileCount, uptime)
	x = w/2 - float32(len(status))*scale*4
	glctx.Uniform4f(1, 0.4, 0.4, 0.5, 1.0)
	renderString(glctx, x, h*0.75, scale*0.5, status)
}

func main() {
	// Start server in background
	go startServer()

	app.Main(func(a app.App) {
		var glctx gl.Context
		var w, h float32

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
				w = float32(e.WidthPx)
				h = float32(e.HeightPx)
			case paint.Event:
				if glctx == nil || w <= 0 || h <= 0 {
					a.Send(paint.Event{})
					continue
				}
				renderScene(glctx, w, h)
				a.Publish()
				// Keep refreshing
				go func() {
					time.Sleep(100 * time.Millisecond)
					a.Send(paint.Event{})
				}()
			}
		}
	})
}
