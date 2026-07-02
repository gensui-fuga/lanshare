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
	// Parse arguments
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

	// Set defaults
	shareDir = os.Getenv("LANSHARE_DIR")
	if shareDir == "" {
		shareDir = filepath.Join(os.TempDir(), "lanshare")
	}

	portStr := os.Getenv("LANSHARE_PORT")
	port = 8080
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Allow CLI args: lanshare [port] [directory]
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
			if len(os.Args) > 2 {
				shareDir = os.Args[2]
			}
		} else {
			shareDir = os.Args[1]
			if len(os.Args) > 2 {
				if p, err := strconv.Atoi(os.Args[2]); err == nil {
					port = p
				}
			}
		}
	}

	// Resolve share directory
	shareDir, _ = filepath.Abs(shareDir)
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		log.Fatalf("Cannot create share directory: %v", err)
	}

	actualPort, err := startServer()
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	printBanner(actualPort)

	// If LANSHARE_OPEN_BROWSER is set, open browser
	if os.Getenv("LANSHARE_OPEN_BROWSER") == "1" {
		ip := getLocalIP()
		url := fmt.Sprintf("http://%s:%d", ip, actualPort)
		openBrowser(url)
	}

	// Wait for signal
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
	fmt.Println()
	fmt.Printf("  📂 Shared:  %s\n", shareDir)
	fmt.Printf("  🌐 Server:  http://%s:%d\n", ip, port)
	fmt.Println()
	fmt.Println("  📱 Open on any device (including iPhone 4):")
	fmt.Printf("     http://%s:%d\n", ip, port)
	fmt.Println()
	fmt.Println("  ⏹  Ctrl+C to stop")
	fmt.Println()
}

func printHelp() {
	fmt.Println("LanShare v1.0.0 - Cross-Platform LAN File Sharing")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lanshare                  # Port 8080, temp dir")
	fmt.Println("  lanshare 9000             # Custom port")
	fmt.Println("  lanshare ./shared         # Custom directory")
	fmt.Println("  lanshare 9000 ./shared    # Custom port + dir")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  LANSHARE_DIR          Share directory")
	fmt.Println("  LANSHARE_PORT         Port number")
	fmt.Println("  LANSHARE_OPEN_BROWSER Set to 1 to auto-open browser")
	fmt.Println()
}
