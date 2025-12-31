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

	// Determine Platform
	isMac := runtime.GOOS == "darwin"
	isWindows := runtime.GOOS == "windows"

	// Helper to run command and stream output
	streamCommand := func(command string, args ...string) {
		go func() {
			startTime := time.Now()

			// Formatted Header
			separator := "====================================="
			header := fmt.Sprintf("%s\n[ %s ]\nTime : %s\nStatus : Running...\n%s",
				separator, feature, startTime.Format("2006-01-02 15:04:05"), separator)
			emitLog(header)

			cmd := exec.Command(command, args...)

			// SILENT EXECUTION FOR WINDOWS
			// SILENT EXECUTION FOR WINDOWS
			if isWindows {
				cmd.SysProcAttr = getSysProcAttr()
			}

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

			// Read logs concurrently
			doneReading := make(chan bool)

			// Stdout reader
			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					emitLog(scanner.Text())
				}
				doneReading <- true
			}()

			// Stderr reader
			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					emitLog(fmt.Sprintf("[ERR] %s", scanner.Text()))
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

			// Minimum delay for visible UX
			elapsed := time.Since(startTime)
			if elapsed < 700*time.Millisecond {
				time.Sleep(700*time.Millisecond - elapsed)
			}

			// Signal done
			wailsRuntime.EventsEmit(a.ctx, "done", feature)
		}()
	}

	// Helper to run PowerShell command securely and silently
	runPowerShell := func(psCommand string) {
		// -NoProfile: No user profile loaded
		// -NonInteractive: No prompt
		// -NoLogo: Hides version banner (though we are capturing output anyway)
		// -Command: The actual code
		streamCommand("powershell", "-NoProfile", "-NonInteractive", "-NoLogo", "-Command", psCommand)
	}

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
	case "List PS Drives":
		if isMac {
			streamCommand("df", "-h")
		} else {
			runPowerShell("Get-PSDrive | Format-Table -AutoSize")
		}

	case "Access HKLM Registry":
		if isMac {
			streamCommand("defaults", "read", "NSGlobalDomain")
		} else {
			// Read keys instead of opening regedit
			runPowerShell("Get-ItemProperty HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion | Select-Object -Property ProgramFilesDir, CommonFilesDir, DevicePath")
		}

	case "Startup Registry Check":
		if isMac {
			streamCommand("ls", "-la", "/Library/LaunchAgents")
		} else {
			runPowerShell("Get-CimInstance Win32_StartupCommand | Select-Object Name, Command, Location | Format-Table -AutoSize")
		}

	// --- MALWARE / ANTI VIRUS ---
	case "Microsoft Malware Removal Tool":
		if isMac {
			emitLog(fmt.Sprintf("=====================================\n[ %s ]\nStatus : Checking...\n=====================================", feature))
			emitLog("[INFO] MRT is Windows-only. Checking XProtect status instead...")
			streamCommand("bash", "-c", "system_profiler SPInstallHistoryDataType | grep -A 5 \"XProtect\"")
		} else {
			// MRT is GUI. Use Get-MpComputerStatus for status instead.
			runPowerShell("Get-MpComputerStatus | Select-Object -Property AntivirusEnabled,AMServiceEnabled,AntispywareEnabled,BehaviorMonitorEnabled,IoavProtectionEnabled,NisEnabled,OnAccessProtectionEnabled | Format-List")
		}

	case "Check Default Antivirus Status":
		if isMac {
			streamCommand("spctl", "--status")
		} else {
			runPowerShell("Get-MpComputerStatus | Format-List")
		}

	// --- REMOTE SERVICES ---
	case "Windows Services":
		if isMac {
			streamCommand("launchctl", "list")
		} else {
			// Replaces services.msc
			runPowerShell("Get-Service | Where-Object {$_.Status -eq 'Running'} | Format-Table -AutoSize")
		}

	case "Remote System Properties":
		if isMac {
			streamCommand("system_profiler", "SPSoftwareDataType")
		} else {
			// Replaces systempropertiesadvanced
			runPowerShell("Get-ComputerInfo | Select-Object CsName, OsName, WindowsVersion, OsArchitecture, BiosVersion | Format-List")
		}

	case "Device Manager (Bluetooth)":
		if isMac {
			streamCommand("system_profiler", "SPBluetoothDataType")
		} else {
			// Replaces devmgmt.msc for Bluetooth
			runPowerShell("Get-PnpDevice -Class Bluetooth | Select-Object Status, Class, FriendlyName, InstanceId | Format-Table -AutoSize")
		}

	case "Registry Editor":
		if isMac {
			streamCommand("open", "/Library/Preferences")
		} else {
			// Cannot run silent regedit, providing info instead
			runPowerShell("Write-Output 'Registry Editor cannot be run silently. Please use system tools if GUI access is needed.'")
		}

	case "Task Manager":
		if isMac {
			streamCommand("open", "-a", "Activity Monitor")
		} else {
			// Replaces taskmgr
			runPowerShell("Get-Process | Sort-Object CPU -Descending | Select-Object -First 20 | Format-Table -AutoSize")
		}

	case "Startup Services":
		if isMac {
			streamCommand("ls", "-la", "/Library/LaunchDaemons")
		} else {
			// Check Auto start services
			runPowerShell("Get-CimInstance Win32_Service | Where-Object StartMode -eq 'Auto' | Select-Object Name, State, StartMode, PathName | Format-Table -AutoSize")
		}

	// --- CLEAN FILES ---
	case "Run Full Cleanup":
		a.performFullCleanup()

	case "Open Temp Folder":
		if isMac {
			streamCommand("open", os.Getenv("TMPDIR"))
		} else {
			// Calculate size instead of opening explorer
			runPowerShell("Get-ChildItem -Path $env:TEMP -Recurse -Force -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum | Select-Object Count, @{Name='Total Size(MB)';Expression={[math]::round($_.Sum/1MB,2)}} | Format-List")
		}

	case "Open Trash / Recycle Bin":
		if isMac {
			streamCommand("open", os.Getenv("HOME")+"/.Trash")
		} else {
			// Calculate Recycle Bin size
			runPowerShell("Get-ChildItem 'C:\\$Recycle.Bin' -Recurse -Force -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum | Select-Object Count, @{Name='Total Size(MB)';Expression={[math]::round($_.Sum/1MB,2)}} | Format-List")
		}

	case "Open Microsoft Office Temp Files":
		if isMac {
			path := os.Getenv("HOME") + "/Library/Containers/com.microsoft.Word/Data/Library/Preferences/AutoRecovery"
			streamCommand("open", path)
		} else {
			// Check common autorecover path if it exists
			runPowerShell("Get-ChildItem -Path \"$env:APPDATA\\Microsoft\\Word\\\" -Filter *.asd -Recurse -ErrorAction SilentlyContinue | Select-Object Name, Length, LastWriteTime | Format-Table")
		}

	default:
		return fmt.Sprintf("Unknown feature: %s", feature)
	}

	return "Request received..."
}

func (a *App) performFullCleanup() {
	// Helper to emit log lines
	emitLog := func(msg string) {
		wailsRuntime.EventsEmit(a.ctx, "log", msg)
	}

	startTime := time.Now()
	separator := "====================================="
	header := fmt.Sprintf("%s\n[ Run Full Cleanup ]\nTime : %s\nStatus : Running...\n%s",
		separator, startTime.Format("2006-01-02 15:04:05"), separator)
	emitLog(header)

	isWindows := runtime.GOOS == "windows"
	isMac := runtime.GOOS == "darwin"

	if isWindows {
		cleanupWindows(emitLog)
	} else if isMac {
		cleanupMac(emitLog)
	} else {
		emitLog("[ERROR] Unsupported platform for cleanup.")
	}

	elapsed := time.Since(startTime)
	if elapsed < 700*time.Millisecond {
		time.Sleep(700*time.Millisecond - elapsed)
	}

	emitLog("\n[OK] Cleanup process completed.")
	wailsRuntime.EventsEmit(a.ctx, "done", "Run Full Cleanup")
}
