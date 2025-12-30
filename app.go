package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	Version string
}

// NewApp creates a new App application struct
func NewApp(version string) *App {
	return &App{
		Version: version,
	}
}

// GetAppVersion returns the current application version
func (a *App) GetAppVersion() string {
	return a.Version
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ExportLogs opens a save dialog and saves the provided content to a file
func (a *App) ExportLogs(content string) error {
	defaultName := fmt.Sprintf("checkpoint-log-%s.txt", time.Now().Format("20060102-150405"))

	options := wailsRuntime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Export Logs",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
		},
	}

	path, err := wailsRuntime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return err
	}

	if path == "" {
		return nil // User cancelled
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// ExecuteCommand runs a system command based on the feature name
func (a *App) ExecuteCommand(feature string) string {

	// Helper to emit log lines
	emitLog := func(msg string) {
		wailsRuntime.EventsEmit(a.ctx, "log", msg)
	}

	// Helper to run command and stream output
	streamCommand := func(command string, args ...string) {
		go func() {
			startTime := time.Now()

			// Formatted Header
			separator := "====================================="
			header := fmt.Sprintf("%s\n[ %s ]\nTime : %s\nStatus : Running...\n%s",
				separator, feature, startTime.Format("2006-01-02 15:04:05"), separator)
			emitLog(header)

			// Force usage of /bin/bash for shell commands if needed,
			// but exec.Command is generally safer with direct args.
			// For complex pipes, we might wrap in bash -c.

			cmd := exec.Command(command, args...)

			// Setup pipes
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				emitLog(fmt.Sprintf("[ERROR] Failed to get stdout pipeline: %s", err))
				wailsRuntime.EventsEmit(a.ctx, "done", feature)
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				emitLog(fmt.Sprintf("[ERROR] Failed to get stderr pipeline: %s", err))
				wailsRuntime.EventsEmit(a.ctx, "done", feature)
				return
			}

			if err := cmd.Start(); err != nil {
				emitLog(fmt.Sprintf("[ERROR] Failed to start command: %s", err))
				wailsRuntime.EventsEmit(a.ctx, "done", feature)
				return
			}

			// Read stdout
			scannerOut := bufio.NewScanner(stdout)
			go func() {
				for scannerOut.Scan() {
					emitLog(scannerOut.Text())
				}
			}()

			// Read stderr
			scannerErr := bufio.NewScanner(stderr)
			go func() {
				for scannerErr.Scan() {
					emitLog(fmt.Sprintf("[ERR] %s", scannerErr.Text()))
				}
			}()

			// We need to wait for scanners? scan blocks.
			// Actually better to run them sync or use a WaitGroup if strictly needed.
			// But for simplicity in this constrained env, allow them to race slightly
			// or just do them sequentially? No, sequentially matches order better if one finishes.
			// But they are pipes.
			// Let's stick to the previous simple loop:
			// The previous loop did sequential read which might block if buffer full on one side.
			// IMPORTANT: Correct way is separate goroutines for reading pipes.

			// Re-implemented simple wait since we are in a single `go func` for the whole command:
			// We can't block on Wait() until reading is done.
			// Simple approach: just use CombinedOutput? No, we need streaming.
			// Let's spin up two goroutines for reading.

			// (Simplification for this specific request to ensure reliability without complex sync code)
			// A trick is to use cmd.Stdout = writerWrapper, cmd.Stderr = ...
			// But sticking to the pattern:

			// We will trust the OS buffers for now or use the previous sequential read if buffers are large enough.
			// However, to satisfy "CRITICAL ISSUE: TOOLS NOT RUNNING", let's be robust.
			// Go routines for reading IS the robust way.

			// Redo reading:
			doneReading := make(chan bool)
			go func() {
				for scannerOut.Scan() {
					emitLog(scannerOut.Text())
				}
				doneReading <- true
			}()
			go func() {
				for scannerErr.Scan() {
					emitLog(fmt.Sprintf("[ERR] %s", scannerErr.Text()))
				}
				doneReading <- true
			}()

			// Wait for process
			err = cmd.Wait()

			// Wait for readers to drain
			<-doneReading
			<-doneReading

			if err != nil {
				emitLog(fmt.Sprintf("\n[STOP] Process finished with error: %v", err))
			} else {
				emitLog("\n[OK] Process completed successfully.")
			}

			// Minimum delay for visible 0.7s
			elapsed := time.Since(startTime)
			if elapsed < 700*time.Millisecond {
				time.Sleep(700*time.Millisecond - elapsed)
			}

			// Signal done
			wailsRuntime.EventsEmit(a.ctx, "done", feature)
		}()
	}

	// Determine Platform (Though we assume macOS based on context,
	// good to be safe or explicit mapping)
	isMac := runtime.GOOS == "darwin"

	switch feature {
	// --- NETWORK ---
	case "Cek IP": // ipconfig /all
		if isMac {
			streamCommand("ifconfig")
		} else {
			streamCommand("ipconfig", "/all")
		}

	case "Cek Routing": // tracert
		if isMac {
			streamCommand("netstat", "-nr")
		} else {
			streamCommand("tracert", "8.8.8.8")
		}

	case "Netstat": // netstat -a
		streamCommand("netstat", "-a")

	case "ARP Table": // arp -a -v
		streamCommand("arp", "-a")

	case "Ping Connectivity": // ping 8.8.8.8
		if isMac {
			streamCommand("ping", "-c", "4", "8.8.8.8")
		} else {
			streamCommand("ping", "8.8.8.8")
		}

	// --- APPLICATION / SYSTEM ---
	case "List PS Drives": // Get-PSDrive -> df -h
		streamCommand("df", "-h")

	case "Access HKLM Registry": // HKLM: -> Defaults
		streamCommand("defaults", "read", "NSGlobalDomain")

	case "Startup Registry Check": // HKLM Run -> LaunchAgents
		streamCommand("ls", "-la", "/Library/LaunchAgents")

	// --- MALWARE / ANTI VIRUS ---
	case "Microsoft Malware Removal Tool": // mrt
		emitLog(fmt.Sprintf("=====================================\n[ %s ]\nStatus : Checking...\n=====================================", feature))
		emitLog("[INFO] MRT is Windows-only. Checking XProtect status instead...")
		streamCommand("bash", "-c", "system_profiler SPInstallHistoryDataType | grep -A 5 \"XProtect\"")

	case "Check Default Antivirus Status": // Get-MpComputerStatus
		// Using spctl or similar as proxy
		streamCommand("spctl", "--status")

	// --- REMOTE SERVICES ---
	case "Windows Services": // services.msc -> launchctl
		streamCommand("launchctl", "list")

	case "Remote System Properties": // systempropertiesadvanced
		streamCommand("system_profiler", "SPSoftwareDataType")

	case "Device Manager (Bluetooth)": // devmgmt.msc -> bluetooth
		streamCommand("system_profiler", "SPBluetoothDataType")

	case "Registry Editor": // regedit -> open Library
		streamCommand("open", "/Library/Preferences")

	case "Task Manager": // taskmgr -> Activity Monitor
		streamCommand("open", "-a", "Activity Monitor")

	case "Startup Services": // launchctl services
		streamCommand("ls", "-la", "/Library/LaunchDaemons")

	// --- CLEAN FILES ---
	case "Open Temp Folder": // %temp%
		streamCommand("open", os.Getenv("TMPDIR"))

	case "Open Trash / Recycle Bin":
		streamCommand("open", os.Getenv("HOME")+"/.Trash")

	case "Open Microsoft Office Temp Files":
		// Attempting standard path
		path := os.Getenv("HOME") + "/Library/Containers/com.microsoft.Word/Data/Library/Preferences/AutoRecovery"
		streamCommand("open", path)

	default:
		return fmt.Sprintf("Unknown feature: %s", feature)
	}

	return "Request received..."
}
