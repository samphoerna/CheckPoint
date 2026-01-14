package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"net"
	"path/filepath"

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

// getAppBaseDir returns the directory where the application is running.
// If running inside a macOS .app bundle, it returns the directory CONTAINING the .app bundle.
// This ensures that portable functionality (like saving screenshots to flash drive) works as expected.
func (a *App) getAppBaseDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(exePath)

	if runtime.GOOS == "darwin" {
		// Typical path: /Volumes/USB/CheckPoint.app/Contents/MacOS/CheckPoint
		// We want: /Volumes/USB/
		// Traverse up if we are inside .app/Contents/MacOS
		// 1. .../MacOS
		// 2. .../Contents
		// 3. .../CheckPoint.app
		// 4. .../ (Target)
		if filepath.Base(dir) == "MacOS" {
			parent := filepath.Dir(dir) // Contents
			if filepath.Base(parent) == "Contents" {
				grandParent := filepath.Dir(parent) // CheckPoint.app
				if filepath.Ext(grandParent) == ".app" {
					return filepath.Dir(grandParent), nil
				}
			}
		}
	}

	return dir, nil
}

// captureScreenshot takes a screenshot silently and saves it to a timestamped folder
func (a *App) captureScreenshot(reason string) {
	// Root screenshot directory relative to executable (Portable)
	baseDir, err := a.getAppBaseDir()
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[ERR] Failed to determine app location: %s. Using temp dir.", err))
		baseDir = os.TempDir()
	}

	// Folder format: CP-SS-<Day><Month><Year> (e.g., CP-SS-08012026)
	dateFolder := fmt.Sprintf("CP-SS-%s", time.Now().Format("02012006"))
	screenshotDir := filepath.Join(baseDir, dateFolder)

	// Attempt to create directory. If read-only (e.g. CD-ROM), fallback to temp.
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[WARN] Cannot write to %s (Read-only?). Falling back to Temp.", screenshotDir))
		baseDir = os.TempDir()
		screenshotDir = filepath.Join(baseDir, dateFolder)
		if err := os.MkdirAll(screenshotDir, 0755); err != nil {
			wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[ERR] Failed to create temp screenshot dir: %s", err))
			return
		}
	}

	// Find next sequence number
	files, _ := os.ReadDir(screenshotDir)
	count := len(files) + 1
	filename := fmt.Sprintf("%03d.png", count) // 001.png
	fullPath := filepath.Join(screenshotDir, filename)

	wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[INFO] Capturing screenshot for '%s'...", reason))
	wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[INFO] Saving to: %s", fullPath))

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// macOS: screencapture -x (silent) -m (main monitor)
		cmd = exec.Command("screencapture", "-x", "-m", fullPath)
	} else {
		// Windows: Powershell snippet to capture screen
		// NOTE: This requires .NET (System.Windows.Forms).
		psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$Screen = [System.Windows.Forms.Screen]::PrimaryScreen
$Width = $Screen.Bounds.Width
$Height = $Screen.Bounds.Height
$Left = $Screen.Bounds.Left
$Top = $Screen.Bounds.Top
$Bitmap = New-Object System.Drawing.Bitmap $Width, $Height
$Graphic = [System.Drawing.Graphics]::FromImage($Bitmap)
$Graphic.CopyFromScreen($Left, $Top, 0, 0, $Bitmap.Size)
$Bitmap.Save('%s')
$Graphic.Dispose()
$Bitmap.Dispose()
`, fullPath)
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
		cmd.SysProcAttr = getSysProcAttr()
	}

	if err := cmd.Run(); err != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[ERR] Screenshot failed: %s", err))
	} else {
		wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("[OK] Screenshot saved."))
	}
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

			// SCREENSHOT TRIGGER: If error or specific conditions met
			// The user requested:
			// 1. Specific popup window appears (hard to detect without keeping process open, but we can infer from "Open/Start" commands)
			// 2. Command executes outside terminal (e.g. "open" / "start")
			// 3. Fails to produce output (stderr) - covered by err check above

			triggerScreenshot := false
			if err != nil {
				triggerScreenshot = true
			}

			// Hardcoded list of features that open external windows
			if feature == "System Information" || feature == "Task Manager" || feature == "Registry Editor" || feature == "Device Manager (Bluetooth)" {
				triggerScreenshot = true
			}

			if triggerScreenshot {
				a.captureScreenshot(feature)
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
	case "System Information": // logic moved from Remote System Properties
		if isMac {
			streamCommand("system_profiler", "SPSoftwareDataType")
			streamCommand("open", "/System/Library/PreferencePanes/Dock.prefPane") // Using Dock as generic placeholder or System Settings
		} else {
			runPowerShell("Get-ComputerInfo | Select-Object CsName, OsName, WindowsVersion, OsArchitecture, BiosVersion | Format-List")
			runPowerShell("Start-Process ms-settings:about")
		}

	case "Check installed applications":
		if isMac {
			streamCommand("system_profiler", "SPApplicationsDataType")
		} else {
			runPowerShell("Get-WmiObject -Class Win32_Product | Select-Object Name, Version, Vendor, InstallDate | Format-Table -AutoSize")
		}

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
			runPowerShell("Get-ItemProperty HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion | Select-Object -Property ProgramFilesDir, CommonFilesDir, DevicePath")
		}

	case "Startup Services": // Moved from Remote Services
		if isMac {
			streamCommand("ls", "-la", "/Library/LaunchDaemons")
		} else {
			runPowerShell("Get-CimInstance Win32_Service | Where-Object StartMode -eq 'Auto' | Select-Object Name, State, StartMode, PathName | Format-Table -AutoSize")
		}

	case "Registry Check": // Renamed from Startup Registry Check
		if isMac {
			streamCommand("ls", "-la", "/Library/LaunchAgents")
		} else {
			runPowerShell("Get-CimInstance Win32_StartupCommand | Select-Object Name, Command, Location | Format-Table -AutoSize")
		}

	case "Registry Editor": // Moved
		if isMac {
			streamCommand("open", "/Library/Preferences")
		} else {
			runPowerShell("Write-Output 'Registry Editor cannot be run silently. Please use system tools if GUI access is needed.'")
			runPowerShell("Start-Process regedit") // Actually open it as per request "moved... must display info... also open" logic applied to System Info, but implied for Editor too
		}

	case "Task Manager": // Moved
		if isMac {
			streamCommand("open", "-a", "Activity Monitor")
		} else {
			runPowerShell("Get-Process | Sort-Object CPU -Descending | Select-Object -First 20 | Format-Table -AutoSize")
			runPowerShell("Start-Process taskmgr")
		}

	// --- MALWARE / ANTI VIRUS ---
	case "Microsoft Malware Removal Tool":
		if isMac {
			emitLog(fmt.Sprintf("=====================================\n[ %s ]\nStatus : Checking...\n=====================================", feature))
			emitLog("[INFO] MRT is Windows-only. Checking XProtect status instead...")
			streamCommand("bash", "-c", "system_profiler SPInstallHistoryDataType | grep -A 5 \"XProtect\"")
		} else {
			runPowerShell("Get-MpComputerStatus | Select-Object -Property AntivirusEnabled,AMServiceEnabled,AntispywareEnabled,BehaviorMonitorEnabled,IoavProtectionEnabled,NisEnabled,OnAccessProtectionEnabled | Format-List")
		}

	case "Check Default Antivirus Status":
		if isMac {
			streamCommand("spctl", "--status")
		} else {
			runPowerShell("Get-MpComputerStatus | Format-List")
		}

	// --- REMOTE SERVICES ---
	case "Check active network service ports":
		emitLog(fmt.Sprintf("=====================================\n[ %s ]\nStatus : Checking Ports...\n=====================================", feature))
		ports := map[string]string{
			"21":   "FTP",
			"22":   "SSH",
			"445":  "SMB",
			"3389": "RDP",
		}

		found := false
		for port, name := range ports {
			timeout := 500 * time.Millisecond
			conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, timeout)
			status := "CLOSED"
			if err == nil {
				conn.Close()
				status = "OPEN"
				found = true
			}
			emitLog(fmt.Sprintf("[%s] Port %s (%s)", status, port, name))
		}
		if !found {
			emitLog("[INFO] No active target services found locally.")
		}
		wailsRuntime.EventsEmit(a.ctx, "done", feature)

	case "Check installed browser extensions":
		emitLog(fmt.Sprintf("=====================================\n[ %s ]\nStatus : Checking Extensions...\n=====================================", feature))

		// Basic check for Chrome Extensions folder
		var extPath string
		if isWindows {
			// Typical path: C:\Users\<User>\AppData\Local\Google\Chrome\User Data\Default\Extensions
			home, _ := os.UserHomeDir()
			extPath = filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Extensions")
			runPowerShell(fmt.Sprintf("Get-ChildItem -Path '%s' -Recurse -Depth 2 | Select-Object FullName", extPath))
		} else {
			// Mac: ~/Library/Application Support/Google/Chrome/Default/Extensions
			home, _ := os.UserHomeDir()
			extPath = filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "Extensions")
			streamCommand("ls", "-R", extPath)
		}

	case "Device Manager (Bluetooth)":
		if isMac {
			streamCommand("system_profiler", "SPBluetoothDataType")
		} else {
			runPowerShell("Get-PnpDevice -Class Bluetooth | Select-Object Status, Class, FriendlyName, InstanceId | Format-Table -AutoSize")
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
