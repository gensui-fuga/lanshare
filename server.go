package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:embed static/* static/**
var staticFS embed.FS

var (
	shareDir  string
	port      int
	appSecret string
	startTime = time.Now()
	hostname  string
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Generate random secret for upload tokens
	b := make([]byte, 8)
	rand.Read(b)
	appSecret = hex.EncodeToString(b)
}

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

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	size := float64(bytes)
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f %s", size, units[i])
}

func sanitizePath(path string) string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, "..", "")
	return path
}

// ---------- API Handlers ----------

type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	SizeStr  string `json:"size_str"`
	Modified int64  `json:"modified"`
	IsDir    bool   `json:"is_dir"`
}

type ServerInfo struct {
	Version   string `json:"version"`
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	ShareDir  string `json:"share_dir"`
	TotalSize int64  `json:"total_size"`
	FileCount int    `json:"file_count"`
	Uptime    string `json:"uptime"`
	Platform  string `json:"platform"`
}

func handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var totalSize int64
	var fileCount int
	filepath.Walk(shareDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	info := ServerInfo{
		Version:   "1.0.0",
		Hostname:  hostname,
		IP:        getLocalIP(),
		Port:      port,
		ShareDir:  shareDir,
		TotalSize: totalSize,
		FileCount: fileCount,
		Uptime:    time.Since(startTime).Round(time.Second).String(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	json.NewEncoder(w).Encode(info)
}

func handleAPIFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	entries, err := os.ReadDir(shareDir)
	if err != nil {
		http.Error(w, `{"error":"cannot read directory"}`, 500)
		return
	}

	var files []FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:     e.Name(),
			Size:     info.Size(),
			SizeStr:  formatSize(info.Size()),
			Modified: info.ModTime().UnixMilli(),
			IsDir:    false,
		})
	}

	if files == nil {
		files = []FileInfo{}
	}
	json.NewEncoder(w).Encode(files)
}

func handleAPIUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	// Limit to 4GB
	r.Body = http.MaxBytesReader(w, r.Body, 4<<30)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse error: %v", err), 400)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "no file uploaded", 400)
		return
	}

	var uploaded []string
	for _, fh := range files {
		fname := filepath.Base(fh.Filename)
		if fname == "" || fname == "." || fname == ".." {
			continue
		}

		src, err := fh.Open()
		if err != nil {
			continue
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(shareDir, fname))
		if err != nil {
			continue
		}
		defer dst.Close()

		written, err := io.Copy(dst, src)
		if err != nil {
			os.Remove(filepath.Join(shareDir, fname))
			continue
		}
		_ = written
		uploaded = append(uploaded, fname)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"uploaded": uploaded,
		"count":    len(uploaded),
	})
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/f/")
	path = sanitizePath(path)
	fullPath := filepath.Join(shareDir, path)

	if !strings.HasPrefix(fullPath, shareDir) {
		http.Error(w, "invalid path", 403)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "file not found", 404)
		return
	}

	fname := filepath.Base(fullPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fname))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, fullPath)
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/s/")
	path = sanitizePath(path)
	fullPath := filepath.Join(shareDir, path)

	if !strings.HasPrefix(fullPath, shareDir) {
		http.Error(w, "invalid path", 403)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "file not found", 404)
		return
	}

	// Detect MIME type by extension
	ext := strings.ToLower(filepath.Ext(fullPath))
	mime := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".png":
		mime = "image/png"
	case ".gif":
		mime = "image/gif"
	case ".webp":
		mime = "image/webp"
	case ".mp4":
		mime = "video/mp4"
	case ".webm":
		mime = "video/webm"
	case ".mp3":
		mime = "audio/mpeg"
	case ".wav":
		mime = "audio/wav"
	case ".ogg":
		mime = "audio/ogg"
	case ".flac":
		mime = "audio/flac"
	case ".pdf":
		mime = "application/pdf"
	case ".txt":
		mime = "text/plain; charset=utf-8"
	case ".html", ".htm":
		mime = "text/html; charset=utf-8"
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, fullPath)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}

	req.Name = sanitizePath(req.Name)
	fullPath := filepath.Join(shareDir, req.Name)
	if !strings.HasPrefix(fullPath, shareDir) {
		http.Error(w, "invalid path", 403)
		return
	}

	if err := os.Remove(fullPath); err != nil {
		http.Error(w, fmt.Sprintf("delete failed: %v", err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Fallback: just return info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "not_available"})
}

// ---------- Middleware ----------

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(lrw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, time.Since(start).Round(time.Millisecond))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// ---------- Server ----------

func startServer() (int, error) {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/info", handleAPIInfo)
	mux.HandleFunc("/api/files", handleAPIFiles)
	mux.HandleFunc("/api/upload", handleAPIUpload)
	mux.HandleFunc("/api/delete", handleDelete)
	mux.HandleFunc("/api/ws", handleWebSocket)
	mux.HandleFunc("/api/devices", handleAPIDevices)

	// File routes
	mux.HandleFunc("/f/", handleDownload)
	mux.HandleFunc("/s/", handleStream)

	// Web UI
	subFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		return 0, fmt.Errorf("embed fs error: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	// Serve index.html for root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			data, err := staticFS.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "not found", 404)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	handler := logMiddleware(corsMiddleware(mux))

	// Try the specified port, fallback to random
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Port unavailable, try random
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return 0, fmt.Errorf("cannot bind: %v", err)
		}
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	port = actualPort

	go func() {
		log.Printf("LanShare server starting on port %d", actualPort)
		log.Printf("Share directory: %s", shareDir)
		log.Printf("Local URL: http://%s:%d", getLocalIP(), actualPort)
		log.Printf("iPhone 4 compatible web UI at: http://%s:%d", getLocalIP(), actualPort)

		if err := http.Serve(listener, handler); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	return actualPort, nil
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux
		cmd = "xdg-open"
		args = []string{url}
	}

	if cmd != "" {
		proc, err := os.StartProcess(cmd, append([]string{cmd}, args...), &os.ProcAttr{})
		if err == nil {
			proc.Release()
		}
	}
}

// Device discovery handler
func handleAPIDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	devices := []map[string]string{
		{
			"name": hostname,
			"ip":   getLocalIP(),
			"port": strconv.Itoa(port),
		},
	}
	json.NewEncoder(w).Encode(devices)
}
