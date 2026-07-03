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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed static/* static/**
var staticFS embed.FS

var (
	uploadDir  string
	downloadDir string
	bgDir      string
	port       int
	appSecret  string
	startTime  = time.Now()
	hostname   string
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
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

func countFiles(dir string) (int, int64) {
	var count int
	var totalSize int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		count++
		totalSize += info.Size()
		return nil
	})
	return count, totalSize
}

func handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uc, us := countFiles(uploadDir)
	dc, ds := countFiles(downloadDir)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":        "1.0.0",
		"hostname":       hostname,
		"ip":             getLocalIP(),
		"port":           port,
		"upload_dir":     uploadDir,
		"download_dir":   downloadDir,
		"upload_count":   uc,
		"upload_size":    us,
		"upload_size_str": formatSize(us),
		"download_count": dc,
		"download_size":  ds,
		"download_size_str": formatSize(ds),
		"total_count":    uc + dc,
		"total_size":     us + ds,
		"total_size_str": formatSize(us + ds),
		"uptime":         time.Since(startTime).Round(time.Second).String(),
		"platform":       runtime.GOOS + "/" + runtime.GOARCH,
	})
}

func handleAPIFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dirType := r.URL.Query().Get("type")
	dir := downloadDir
	if dirType == "upload" {
		dir = uploadDir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	var files []map[string]interface{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		if info == nil {
			continue
		}
		files = append(files, map[string]interface{}{
			"name":     e.Name(),
			"size":     info.Size(),
			"size_str": formatSize(info.Size()),
			"modified": info.ModTime().UnixMilli(),
		})
	}
	if files == nil {
		files = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(files)
}

func handleAPIUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<30)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("parse error: %v", err), 400)
		return
	}

	files := r.MultipartForm.File["file"]
	var uploaded []string
	for _, fh := range files {
		fname := filepath.Base(fh.Filename)
		if fname == "" || fname == "." || fname == ".." {
			continue
		}
		src, _ := fh.Open()
		if src == nil {
			continue
		}
		defer src.Close()

		dst, _ := os.Create(filepath.Join(uploadDir, fname))
		if dst == nil {
			continue
		}
		defer dst.Close()
		io.Copy(dst, src)
		uploaded = append(uploaded, fname)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"uploaded": uploaded, "count": len(uploaded),
	})
}

func handleAPIDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	req.Name = sanitizePath(req.Name)

	dir := downloadDir
	if req.Type == "upload" {
		dir = uploadDir
	}
	fullPath := filepath.Join(dir, req.Name)
	if !strings.HasPrefix(fullPath, dir) {
		http.Error(w, "invalid path", 403)
		return
	}
	os.Remove(fullPath)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	path := sanitizePath(strings.TrimPrefix(r.URL.Path, "/f/"))

	// Try download dir first, then upload dir
	var fullPath string
	if _, err := os.Stat(filepath.Join(downloadDir, path)); err == nil {
		fullPath = filepath.Join(downloadDir, path)
	} else {
		fullPath = filepath.Join(uploadDir, path)
	}

	if !strings.HasPrefix(fullPath, downloadDir) && !strings.HasPrefix(fullPath, uploadDir) {
		http.Error(w, "invalid path", 403)
		return
	}
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "file not found", 404)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(fullPath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, fullPath)
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	path := sanitizePath(strings.TrimPrefix(r.URL.Path, "/s/"))
	var fullPath string
	if _, err := os.Stat(filepath.Join(downloadDir, path)); err == nil {
		fullPath = filepath.Join(downloadDir, path)
	} else {
		fullPath = filepath.Join(uploadDir, path)
	}
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "file not found", 404)
		return
	}
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

func handleAPISettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "POST" {
		var req struct {
			UploadDir   string `json:"upload_dir"`
			DownloadDir string `json:"download_dir"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.UploadDir != "" {
			uploadDir = req.UploadDir
			os.MkdirAll(uploadDir, 0755)
		}
		if req.DownloadDir != "" {
			downloadDir = req.DownloadDir
			os.MkdirAll(downloadDir, 0755)
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"upload_dir":   uploadDir,
		"download_dir": downloadDir,
	})
}

func handleUploadBg(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50MB max
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "parse error", 400)
		return
	}
	files := r.MultipartForm.File["bg"]
	if len(files) == 0 {
		http.Error(w, "no file", 400)
		return
	}
	fh := files[0]
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" && ext != ".bmp" {
		http.Error(w, "unsupported format: "+ext, 400)
		return
	}
	src, _ := fh.Open()
	defer src.Close()

	// Clean old backgrounds
	old, _ := filepath.Glob(filepath.Join(bgDir, "bg-*"))
	for _, o := range old {
		os.Remove(o)
	}

	bgName := "bg-" + hex.EncodeToString([]byte(time.Now().String()))[:8] + ext
	dst, _ := os.Create(filepath.Join(bgDir, bgName))
	defer dst.Close()
	io.Copy(dst, src)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"url":    "/bg/" + bgName,
	})
}

func handleServeBg(w http.ResponseWriter, r *http.Request) {
	path := sanitizePath(strings.TrimPrefix(r.URL.Path, "/bg/"))
	fullPath := filepath.Join(bgDir, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, fullPath)
}

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

func startServer() (int, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/info", handleAPIInfo)
	mux.HandleFunc("/api/files", handleAPIFiles)
	mux.HandleFunc("/api/upload", handleAPIUpload)
	mux.HandleFunc("/api/delete", handleAPIDelete)
	mux.HandleFunc("/api/settings", handleAPISettings)
	mux.HandleFunc("/api/upload-bg", handleUploadBg)
	mux.HandleFunc("/f/", handleDownload)
	mux.HandleFunc("/s/", handleStream)
	mux.HandleFunc("/bg/", handleServeBg)

	// Web UI with cache control
	subFS, _ := fs.Sub(staticFS, "static")
	fileServer := http.FileServer(http.FS(subFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			data, err := staticFS.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "not found", 404)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fileServer.ServeHTTP(w, r)
	})

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &lrw{rw: w, code: 200}
		mux.ServeHTTP(lrw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.code, time.Since(start).Round(time.Millisecond))
	}))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return 0, fmt.Errorf("cannot bind: %v", err)
		}
	}
	port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		log.Printf("LanShare v1.0.0  http://%s:%d", getLocalIP(), port)
		http.Serve(listener, handler)
	}()
	return port, nil
}

type lrw struct {
	rw   http.ResponseWriter
	code int
}

func (l *lrw) Header() http.Header         { return l.rw.Header() }
func (l *lrw) Write(b []byte) (int, error) { return l.rw.Write(b) }
func (l *lrw) WriteHeader(c int)           { l.code = c; l.rw.WriteHeader(c) }

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
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	exec.Command(cmd, args...).Start()
}
