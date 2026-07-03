package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			printHelp()
			return
		case "-v", "--version":
			fmt.Println("LanShare v1.0.0")
			return
		}
	}

	// Defaults
	base := filepath.Join(os.TempDir(), "lanshare")
	uploadDir = os.Getenv("LANSHARE_UPLOAD_DIR")
	downloadDir = os.Getenv("LANSHARE_DOWNLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(base, "uploads")
	}
	if downloadDir == "" {
		downloadDir = filepath.Join(base, "downloads")
	}
	bgDir = filepath.Join(base, "bg")

	portStr := os.Getenv("LANSHARE_PORT")
	port = 8080
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// CLI: lanshare [port] [dir]   or   lanshare [port] [upload_dir] [download_dir]
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
			if len(os.Args) > 2 {
				uploadDir = os.Args[2]
				if len(os.Args) > 3 {
					downloadDir = os.Args[3]
				} else {
					downloadDir = uploadDir
				}
			}
		} else {
			uploadDir = os.Args[1]
			downloadDir = os.Args[1]
			if len(os.Args) > 2 {
				if p, err := strconv.Atoi(os.Args[2]); err == nil {
					port = p
				}
			}
		}
	}

	// Resolve and create dirs
	uploadDir, _ = filepath.Abs(uploadDir)
	downloadDir, _ = filepath.Abs(downloadDir)
	bgDir = filepath.Join(filepath.Dir(uploadDir), "bg")
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(downloadDir, 0755)
	os.MkdirAll(bgDir, 0755)

	actualPort, err := startServer()
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	printBanner(actualPort)

	if os.Getenv("LANSHARE_OPEN_BROWSER") == "1" {
		url := fmt.Sprintf("http://%s:%d", getLocalIP(), actualPort)
		openBrowser(url)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\nShutting down...")
}

func printBanner(port int) {
	ip := getLocalIP()
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║          LanShare v1.0               ║")
	fmt.Println("  ║    Cross-Platform LAN File Share     ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Printf("\n  📤 Upload:   %s\n", uploadDir)
	fmt.Printf("  📥 Download: %s\n", downloadDir)
	fmt.Printf("  🌐 Server:   http://%s:%d\n\n", ip, port)
	fmt.Println("  ⏹  Ctrl+C to stop")
	fmt.Println()
}

func printHelp() {
	fmt.Println("LanShare v1.0.0 - Cross-Platform LAN File Sharing")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lanshare                           # Port 8080, temp dirs")
	fmt.Println("  lanshare 9000                      # Custom port")
	fmt.Println("  lanshare ./shared                  # Single dir for both")
	fmt.Println("  lanshare 9000 ./uploads ./downloads # Port + separate dirs")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  LANSHARE_UPLOAD_DIR    Upload directory")
	fmt.Println("  LANSHARE_DOWNLOAD_DIR  Download directory")
	fmt.Println("  LANSHARE_PORT          Port number")
	fmt.Println()
}
