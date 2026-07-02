//go:build gui
// +build gui

package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type guiState struct {
	app        fyne.App
	window     fyne.Window

	// Widgets
	statusIcon  *widget.Icon
	statusLabel *widget.Label
	ipLabel     *widget.Label
	portLabel   *widget.Label
	dirLabel    *widget.Label
	countLabel  *widget.Label
	fileList    *widget.List
	logArea     *widget.Entry

	// Server
	serverPort int
	shareDir   string
	logs       []string
}

func main() {
	// Parse args for directory
	shareDir := os.Getenv("LANSHARE_DIR")
	if shareDir == "" {
		shareDir = filepath.Join(os.TempDir(), "lanshare")
	}
	portStr := os.Getenv("LANSHARE_PORT")
	port := 8080
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

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

	shareDir, _ = filepath.Abs(shareDir)
	os.MkdirAll(shareDir, 0755)

	actualPort, err := startServer()
	if err != nil {
		log.Fatalf("Cannot start server: %v", err)
	}

	// Start GUI
	a := app.NewWithID("com.lanshare.app")
	w := a.NewWindow("LanShare")

	state := &guiState{
		app:    a,
		window: w,
		serverPort: actualPort,
		shareDir:   shareDir,
	}

	buildGUI(state)

	// Set icon
	w.SetIcon(theme.ComputerIcon())
	w.Resize(fyne.NewSize(720, 520))
	w.CenterOnScreen()

	// System tray
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("LanShare",
			fyne.NewMenuItem("打開視窗", func() {
				w.Show()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("退出", func() {
				a.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}

	// Close to tray on window close (desktop only)
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	// Handle signals
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		a.Quit()
	}()

	state.addLog(fmt.Sprintf("LanShare 已啟動於 port %d", actualPort))
	state.addLog(fmt.Sprintf("共享目錄: %s", shareDir))
	state.addLog(fmt.Sprintf("本地訪問: http://localhost:%d", actualPort))
	state.addLog(fmt.Sprintf("區域網路: http://%s:%d", getLocalIP(), actualPort))

	w.ShowAndRun()
}

func buildGUI(state *guiState) {
	w := state.window

	// ---- Status Card ----
	statusIcon := widget.NewIcon(theme.ConfirmIcon())
	state.statusIcon = statusIcon

	statusLabel := widget.NewLabelWithStyle("伺服器運行中", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	state.statusLabel = statusLabel

	ipLabel := widget.NewLabel(fmt.Sprintf("IP: %s", getLocalIP()))
	state.ipLabel = ipLabel

	portLabel := widget.NewLabel(fmt.Sprintf("Port: %d", state.serverPort))
	state.portLabel = portLabel

	dirLabel := widget.NewLabel(fmt.Sprintf("目錄: %s", state.shareDir))
	state.dirLabel = dirLabel

	countLabel := widget.NewLabel("檔案: 0")
	state.countLabel = countLabel

	statusCard := container.NewVBox(
		widget.NewCard("", "伺服器狀態", container.NewVBox(
			container.NewCenter(statusIcon),
			container.NewCenter(statusLabel),
			layout.NewSpacer(),
			ipLabel,
			portLabel,
			dirLabel,
			countLabel,
		)),
	)

	// ---- Quick Actions ----
	openBtn := widget.NewButtonWithIcon("在瀏覽器開啟", theme.ViewFullScreenIcon(), func() {
		url := fmt.Sprintf("http://localhost:%d", state.serverPort)
		openBrowser(url)
	})

	copyBtn := widget.NewButtonWithIcon("複製區域網路網址", theme.ContentCopyIcon(), func() {
		url := fmt.Sprintf("http://%s:%d", getLocalIP(), state.serverPort)
		w.Clipboard().SetContent(url)
		dialog.ShowInformation("已複製", "區域網路網址已複製到剪貼板", w)
	})

	openDirBtn := widget.NewButtonWithIcon("開啟共享目錄", theme.FolderOpenIcon(), func() {
		openFileManager(state.shareDir)
	})

	actionsCard := widget.NewCard("", "快捷操作", container.NewVBox(
		openBtn,
		copyBtn,
		openDirBtn,
	))

	// ---- File List ----
	fileList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.FileIcon()),
				widget.NewLabel(""),
				layout.NewSpacer(),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {},
	)
	state.fileList = fileList

	refreshBtn := widget.NewButtonWithIcon("重新整理", theme.ViewRefreshIcon(), func() {
		refreshFileList(state)
	})

	filesCard := widget.NewCard("", "共享檔案", container.NewBorder(
		refreshBtn, nil, nil, nil,
		fileList,
	))

	// ---- Log ----
	logEntry := widget.NewEntry()
	logEntry.MultiLine = true
	logEntry.Disable()
	logEntry.SetMinRowsVisible(6)
	state.logArea = logEntry

	// ---- Layout ----
	leftPanel := container.NewVBox(
		statusCard,
		layout.NewSpacer(),
		actionsCard,
	)

	rightPanel := container.NewBorder(
		filesCard, nil, nil, nil,
		container.NewBorder(
			widget.NewLabelWithStyle("日誌", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			container.NewScroll(logEntry),
		),
	)

	split := container.NewHSplit(
		container.NewScroll(leftPanel),
		rightPanel,
	)
	split.SetOffset(0.35)

	w.SetContent(split)

	// Periodic refresh
	go func() {
		for range time.NewTicker(3 * time.Second).C {
			refreshFileList(state)
		}
	}()
}

func refreshFileList(state *guiState) {
	entries, err := os.ReadDir(state.shareDir)
	if err != nil {
		return
	}

	type fileInfo struct {
		name string
		size string
	}
	var files []fileInfo
	var totalSize int64

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()
		files = append(files, fileInfo{name: e.Name(), size: formatSize(info.Size())})
	}

	if files == nil {
		files = []fileInfo{}
	}

	state.fileList.Length = func() int { return len(files) }
	state.fileList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		container := item.(*fyne.Container)
		icon := container.Objects[0].(*widget.Icon)
		label := container.Objects[1].(*widget.Label)
		sizeLabel := container.Objects[3].(*widget.Label)

		icon.SetResource(theme.FileIcon())
		label.SetText(files[id].name)
		sizeLabel.SetText(files[id].size)
	}
	state.fileList.Refresh()

	state.countLabel.SetText(fmt.Sprintf("檔案: %d (%s)", len(files), formatSize(totalSize)))
}

func (state *guiState) addLog(msg string) {
	state.logs = append(state.logs, msg)
	if len(state.logs) > 100 {
		state.logs = state.logs[len(state.logs)-100:]
	}
	state.logArea.SetText("")
	for _, l := range state.logs {
		state.logArea.SetText(state.logArea.Text + l + "\n")
	}
	state.logArea.CursorRow = len(state.logs) - 1
}

func openFileManager(path string) {
	switch runtime.GOOS {
	case "windows":
		execCommand("explorer", path)
	case "darwin":
		execCommand("open", path)
	default:
		execCommand("xdg-open", path)
	}
}

func execCommand(name string, arg string) {
	proc, err := os.StartProcess(name, []string{name, arg}, &os.ProcAttr{})
	if err == nil {
		proc.Release()
	}
}
