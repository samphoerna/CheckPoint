package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"path/filepath"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ComplianceResult is returned by the compliance check
type ComplianceResult struct {
	Compliant  bool   `json:"compliant"`
	RepairText string `json:"repairText"`
}

// App struct
type App struct {
	ctx     context.Context
	Version string
}

// NewApp creates a new App application struct
func NewApp(version string) *App {
	return &App{
		Version: "V.0.1.34c",
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
	a.initFolders()
	go a.SyncAuditResults()
}

// ExportLogs opens a save dialog and saves the provided content to a file
func (a *App) ExportLogs(content string) error {
	defaultName := fmt.Sprintf("checkpoint-log-%s.txt", time.Now().Format("20060102-150405"))

	// Resolve default directory to app base dir
	baseDir, _ := a.getAppBaseDir()

	options := wailsRuntime.SaveDialogOptions{
		DefaultFilename:  defaultName,
		DefaultDirectory: baseDir,
		Title:            "Export Logs",
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
	// 1. Disable entirely on macOS
	if runtime.GOOS == "darwin" {
		return
	}

	// 2. Delayed Capture (Windows) - Wait for window to be visible
	time.Sleep(2 * time.Second)

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
		wailsRuntime.EventsEmit(a.ctx, "log", "[OK] Screenshot saved.")
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

			// Trigger screenshot for Application/System category tools and errors
			triggerScreenshot := false
			if err != nil {
				triggerScreenshot = true
			}

			// Application / System category features
			appSystemFeatures := map[string]bool{
				"System Information":    true,
				"IP Check":              true,
				"Network Adapter":       true,
				"Task Manager":          true,
				"Installed App":         true,
				"Startup Services":      true,
				"Remote Access Setting": true,
				"Browser Extension":     true,
				"Antivirus":             true,
				"Sys_Security_Status":   true,
				"PORT 21 (FTP)":         true,
				"PORT 22 (SSH)":         true,
				"PORT 23 (TELNET)":      true,
				"PORT 445 (SMB)":        true,
				"PORT 3389 (RDP)":       true,
				"BLUETOOTH":             true,
				"FILE SHARING":          true,
			}

			if appSystemFeatures[feature] {
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
	case "Internet Connectivity":
		if isMac {
			streamCommand("curl", "-I", "https://1.1.1.1")
		} else {
			streamCommand("curl", "https://1.1.1.1")
		}

	case "Public IP Check":
		if isMac {
			streamCommand("curl", "-s", "--max-time", "5", "https://api.ipify.org")
		} else {
			runPowerShell("Invoke-RestMethod -Uri \"https://api.ipify.org\" -TimeoutSec 5")
		}

	case "VPN Connection":
		if isMac {
			streamCommand("scutil", "--nc", "list")
		} else {
			runPowerShell("Get-VpnConnection")
		}

	case "Firewall":
		if isMac {
			streamCommand("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
		} else {
			runPowerShell("Get-NetFirewallProfile")
		}

	case "Proxy State":
		if isMac {
			streamCommand("scutil", "--proxy")
		} else {
			runPowerShell("Get-ItemProperty -Path \"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings\" | Select-Object ProxyEnable,ProxyServer")
		}

	// Removed old networking functions since user requested removal.

	// --- APPLICATION / SYSTEM ---
	case "System Information":
		if isMac {
			streamCommand("system_profiler", "SPHardwareDataType", "SPSoftwareDataType")
		} else {
			streamCommand("systeminfo")
		}

	case "IP Check":
		if isMac {
			streamCommand("ifconfig")
		} else {
			streamCommand("ipconfig", "/all")
		}

	case "Network Adapter":
		if isMac {
			streamCommand("system_profiler", "SPHardwareDataType", "SPSoftwareDataType")
		} else {
			runPowerShell("Get-NetAdapter | Out-GridView -Title 'Network Interfaces'")
		}

	case "Task Manager":
		if isMac {
			streamCommand("open", "-a", "Activity Monitor")
		} else {
			runPowerShell("Start-Process taskmgr")
		}

	case "Installed App":
		if isMac {
			streamCommand("open", "/Applications")
		} else {
			runPowerShell("Get-StartApps | Out-GridView -Title 'Installed Applications'")
		}

	case "Startup Services":
		if isMac {
			streamCommand("open", "/Library/LaunchAgents")
		} else {
			runPowerShell("Get-CimInstance Win32_StartupCommand | Select Name, Command, Location | Out-GridView -Title 'Startup Applications'")
		}

	case "Remote Access Setting":
		if isMac {
			streamCommand("open", "/System/Library/PreferencePanes/SharingPref.prefPane")
		} else {
			runPowerShell("Start systempropertiesremote")
		}

	case "Browser Extension":
		emitLog(fmt.Sprintf("=====================================\n[ %s ]\nStatus : Checking Extensions...\n=====================================", feature))
		if isWindows {
			runPowerShell("Start-Process chrome 'chrome://extensions'")
			runPowerShell("Start-Process firefox 'about:addons'")
			runPowerShell("Start-Process brave 'brave://extensions'")
		} else {
			streamCommand("bash", "-c", "open -a 'Google Chrome' chrome://extensions & open -a 'Firefox' about:addons & open -a 'Brave Browser' brave://extensions")
		}

	// --- REMOTE SERVICES ---
	case "PORT 21 (FTP)":
		if isMac {
			streamCommand("bash", "-c", "netstat -an | grep \".21 \" | grep LISTEN")
		} else {
			streamCommand("cmd", "/c", "netstat -ano | findstr :21")
		}

	case "PORT 22 (SSH)":
		if isMac {
			streamCommand("bash", "-c", "netstat -an | grep \".22 \" | grep LISTEN")
		} else {
			streamCommand("cmd", "/c", "netstat -ano | findstr :22")
		}

	case "PORT 23 (TELNET)":
		if isMac {
			streamCommand("bash", "-c", "netstat -an | grep \".23 \" | grep LISTEN")
		} else {
			streamCommand("cmd", "/c", "netstat -ano | findstr :23")
		}

	case "PORT 445 (SMB)":
		if isMac {
			streamCommand("bash", "-c", "netstat -an | grep \".445 \" | grep LISTEN")
		} else {
			streamCommand("cmd", "/c", "netstat -ano | findstr :445")
		}

	case "PORT 3389 (RDP)":
		if isMac {
			streamCommand("bash", "-c", "netstat -an | grep \".3389 \" | grep LISTEN")
		} else {
			streamCommand("cmd", "/c", "netstat -ano | findstr :3389")
		}

	case "BLUETOOTH":
		if isMac {
			streamCommand("system_profiler", "SPBluetoothDataType")
		} else {
			runPowerShell("Get-PnpDevice -Class Bluetooth")
		}

	case "FILE SHARING":
		if isMac {
			streamCommand("ifconfig", "awdl0")
		} else {
			runPowerShell("Get-ItemProperty -Path \"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\CDP\" -Name NearShareChannelUserAuthzPolicy")
		}

	// --- SECURITY & ANTIVIRUS ---
	case "Antivirus":
		if isMac {
			streamCommand("bash", "-c", "system_profiler SPInstallHistoryDataType | grep -i xprotect")
		} else {
			runPowerShell("Get-CimInstance -Namespace root\\SecurityCenter2 -ClassName AntivirusProduct | Select-Object displayName, productState")
		}

	case "Sys_Security_Status":
		if isMac {
			streamCommand("bash", "-c", "csrutil status && spctl --status")
		} else {
			runPowerShell("Get-MpComputerStatus | Select-Object AntivirusEnabled,AMServiceEnabled,AntispywareEnabled,RealTimeProtectionEnabled,BehaviorMonitorEnabled,IoavProtectionEnabled,NISEnabled | Format-List")
		}

	// --- CLEAN FILES ---
	case "Run Full Cleanup":
		a.performFullCleanup()

	case "Open Temp Folder":
		if isMac {
			streamCommand("open", os.Getenv("TMPDIR"))
		} else {
			// Calculate size AND open folder
			runPowerShell("Get-ChildItem -Path $env:TEMP -Recurse -Force -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum | Select-Object Count, @{Name='Total Size(MB)';Expression={[math]::round($_.Sum/1MB,2)}} | Format-List")
			runPowerShell("Start-Process explorer $env:TEMP")
		}

	case "Open Trash / Recycle Bin":
		if isMac {
			streamCommand("open", os.Getenv("HOME")+"/.Trash")
		} else {
			// Calculate Recycle Bin size AND open
			runPowerShell("Get-ChildItem 'C:\\$Recycle.Bin' -Recurse -Force -ErrorAction SilentlyContinue | Measure-Object -Property Length -Sum | Select-Object Count, @{Name='Total Size(MB)';Expression={[math]::round($_.Sum/1MB,2)}} | Format-List")
			runPowerShell("Start-Process explorer shell:RecycleBinFolder")
		}

	case "Open Office Temp Files":
		if isMac {
			path := os.Getenv("HOME") + "/Library/Containers/com.microsoft.Word/Data/Library/Preferences/AutoRecovery"
			streamCommand("open", path)
		} else {
			// Check common autorecover path AND open
			wordPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Word")
			runPowerShell(fmt.Sprintf("Get-ChildItem -Path '%s' -Filter *.asd -Recurse -ErrorAction SilentlyContinue | Select-Object Name, Length, LastWriteTime | Format-Table", wordPath))
			runPowerShell(fmt.Sprintf("Start-Process explorer '%s'", wordPath))
		}

	default:
		return fmt.Sprintf("Unknown feature: %s", feature)
	}

	return "Request received..."
}

// CheckNetworkCompliance executes silent boolean checks for Network Category features
func (a *App) CheckNetworkCompliance(feature string) ComplianceResult {
	isMac := runtime.GOOS == "darwin"
	isWindows := runtime.GOOS == "windows"

	runSilentCommand := func(command string, args ...string) string {
		cmd := exec.Command(command, args...)
		if isWindows {
			cmd.SysProcAttr = getSysProcAttr()
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			return strings.TrimSpace(string(out)) + " " + err.Error()
		}
		return strings.TrimSpace(string(out))
	}

	runSilentPowerShell := func(psCommand string) string {
		return runSilentCommand("powershell", "-NoProfile", "-NonInteractive", "-Command", psCommand)
	}

	switch feature {
	case "Internet Connectivity":
		if isMac {
			// bash -c because we pipe or use conditionals easily
			out := runSilentCommand("bash", "-c", "curl -I --max-time 5 -s https://1.1.1.1 >/dev/null 2>&1; if [ $? -eq 0 ]; then echo true; else echo false; fi")
			// We want NO internet. So if "true", it is NOT compliant.
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "networksetup -setairportpower en0 off",
			}
		} else {
			script := `
curl -I --max-time 5 -s https://1.1.1.1 >$null 2>&1
if ($LASTEXITCODE -eq 0) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Disable-NetAdapter -Name \"*\" -Confirm:$false",
			}
		}

	case "Public IP Check":
		if isMac {
			out := runSilentCommand("bash", "-c", "publicIP=$(curl -s --max-time 5 https://api.ipify.org); if [[ -n \"$publicIP\" ]]; then echo true; else echo false; fi")
			// want NO public IP
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "networksetup -setairportpower en0 off",
			}
		} else {
			script := `
try {
	$publicIP = Invoke-RestMethod -Uri "https://api.ipify.org" -TimeoutSec 5
	if ($publicIP) { $true } else { $false }
}
catch {
	$false
}`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Disable-NetAdapter -Name \"*\" -Confirm:$false",
			}
		}

	case "VPN Connection":
		if isMac {
			// scutil --nc list returns "[Disconnected]" or "(Connected)".
			// We must ensure "Connected" matches without "Disconnected".
			out := runSilentCommand("bash", "-c", "if scutil --nc list | grep -i \"connected\" | grep -v -i \"disconnected\" > /dev/null; then echo true; else echo false; fi")
			// disconnected -> compliant (no connected VPNs)
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "scutil --nc stop <VPN_NAME>",
			}
		} else {
			script := `
$vpn = Get-VpnConnection -ErrorAction SilentlyContinue | Where-Object {$_.ConnectionStatus -eq "Connected"}
if ($vpn) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "rasdial /disconnect",
			}
		}

	case "Firewall":
		if isMac {
			out := runSilentCommand("bash", "-c", "if /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate | grep -iq \"enabled\"; then echo true; else echo false; fi")
			// User requested: Firewal Disabled (state = 0) -> COMPLIANT
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate on",
			}
		} else {
			script := `
if ((Get-NetFirewallProfile | Where-Object {$_.Enabled -eq $true -or $_.Enabled -eq 1 -or $_.Enabled -match "True"}).Count -ge 1) { $true } else { $false }`
			out := runSilentPowerShell(script)
			// User requested: Firewall Disabled -> COMPLIANT
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled True",
			}
		}

	case "Proxy State":
		if isMac {
			out := runSilentCommand("bash", "-c", "proxy=$(scutil --proxy | grep HTTPEnable | awk '{print $3}'); if [[ \"$proxy\" == \"1\" ]]; then echo true; else echo false; fi")
			// Proxy OFF (0) -> Compliant
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "networksetup -setwebproxystate Wi-Fi off",
			}
		} else {
			script := `
$proxy = Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings" -ErrorAction SilentlyContinue
if ($proxy -and $proxy.ProxyEnable -eq 1) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Set-ItemProperty -Path \"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings\" -Name ProxyEnable -Value 0",
			}
		}

	case "Antivirus":
		if isMac {
			out := runSilentCommand("bash", "-c", "if [ -f \"/System/Library/CoreServices/XProtect.bundle/Contents/Resources/XProtect.meta.plist\" ]; then echo true; else echo false; fi")
			compliant := out == "true"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo spctl --master-enable",
			}
		} else {
			script := `
$av = Get-CimInstance -Namespace root\SecurityCenter2 -ClassName AntivirusProduct -ErrorAction SilentlyContinue

if ($av) {
    $true
} else {
    $false
}`
			out := runSilentPowerShell(script)
			compliant := out == "True"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Set-MpPreference -DisableRealtimeMonitoring $false",
			}
		}

	case "Sys_Security_Status":
		if isMac {
			out := runSilentCommand("bash", "-c", "sip=$(csrutil status | grep -i enabled); gate=$(spctl --status | grep -i enabled); if [[ -n \"$sip\" && -n \"$gate\" ]]; then echo true; else echo false; fi")
			compliant := out == "true"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo spctl --master-enable && csrutil enable",
			}
		} else {
			script := `
$status = Get-MpComputerStatus

if ($status.RealTimeProtectionEnabled) {
    $true
} else {
    $false
}`
			out := runSilentPowerShell(script)
			compliant := out == "True"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Open Windows Security > Virus & Threat Protection > Manage Settings",
			}
		}

	// --- REMOTE SERVICES ---
	case "PORT 21 (FTP)":
		if isMac {
			out := runSilentCommand("bash", "-c", "if lsof -iTCP:21 -sTCP:LISTEN -nP >/dev/null 2>&1; then echo true; else echo false; fi")
			// Active -> Non-compliant
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo launchctl unload -w /System/Library/LaunchDaemons/ftp.plist",
			}
		} else {
			script := `if (Get-NetTCPConnection -State Listen -LocalPort 21 -ErrorAction SilentlyContinue) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Stop-Service ftpsvc -Force\nSet-Service ftpsvc -StartupType Disabled",
			}
		}

	case "PORT 22 (SSH)":
		if isMac {
			out := runSilentCommand("bash", "-c", "if lsof -iTCP:22 -sTCP:LISTEN -nP >/dev/null 2>&1; then echo true; else echo false; fi")
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo systemsetup -setremotelogin off",
			}
		} else {
			script := `if (Get-NetTCPConnection -State Listen -LocalPort 22 -ErrorAction SilentlyContinue) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Stop-Service sshd\nSet-Service sshd -StartupType Disabled",
			}
		}

	case "PORT 23 (TELNET)":
		if isMac {
			out := runSilentCommand("bash", "-c", "if lsof -iTCP:23 -sTCP:LISTEN -nP >/dev/null 2>&1; then echo true; else echo false; fi")
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo launchctl unload -w /System/Library/LaunchDaemons/telnet.plist",
			}
		} else {
			script := `if (Get-NetTCPConnection -State Listen -LocalPort 23 -ErrorAction SilentlyContinue) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Stop-Service TlntSvr\nSet-Service TlntSvr -StartupType Disabled",
			}
		}

	case "PORT 445 (SMB)":
		if isMac {
			out := runSilentCommand("bash", "-c", "if lsof -iTCP:445 -sTCP:LISTEN -nP >/dev/null 2>&1; then echo true; else echo false; fi")
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo launchctl unload -w /System/Library/LaunchDaemons/com.apple.smbd.plist",
			}
		} else {
			script := `if (Get-NetTCPConnection -State Listen -LocalPort 445 -ErrorAction SilentlyContinue) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Stop-Service lanmanserver\nSet-Service lanmanserver -StartupType Disabled",
			}
		}

	case "PORT 3389 (RDP)":
		if isMac {
			out := runSilentCommand("bash", "-c", "if lsof -iTCP:3389 -sTCP:LISTEN -nP >/dev/null 2>&1; then echo true; else echo false; fi")
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo /System/Library/CoreServices/RemoteManagement/ARDAgent.app/Contents/Resources/kickstart -deactivate",
			}
		} else {
			script := `if (Get-NetTCPConnection -State Listen -LocalPort 3389 -ErrorAction SilentlyContinue) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Stop-Service TermService",
			}
		}

	case "BLUETOOTH":
		if isMac {
			out := runSilentCommand("bash", "-c", "if system_profiler SPBluetoothDataType | grep -q \"State: On\"; then echo true; else echo false; fi")
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo defaults write /Library/Preferences/com.apple.Bluetooth ControllerPowerState -int 0\nsudo killall -HUP bluetoothd",
			}
		} else {
			script := `
$bt = Get-PnpDevice -Class Bluetooth | Where-Object {$_.Status -eq "OK"}
if ($bt) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Disable-PnpDevice -Class Bluetooth -Confirm:$false",
			}
		}

	case "FILE SHARING":
		if isMac {
			out := runSilentCommand("bash", "-c", "if ifconfig awdl0 | grep -q \"status: active\"; then echo true; else echo false; fi")
			compliant := out == "false"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "sudo ifconfig awdl0 down",
			}
		} else {
			script := `
$share = Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\CDP" -Name NearShareChannelUserAuthzPolicy -ErrorAction SilentlyContinue
if ($share.NearShareChannelUserAuthzPolicy -gt 0) { $true } else { $false }`
			out := runSilentPowerShell(script)
			compliant := out == "False"
			return ComplianceResult{
				Compliant:  compliant,
				RepairText: "Set-ItemProperty -Path \"HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\CDP\" -Name NearShareChannelUserAuthzPolicy -Value 0",
			}
		}
	}

	return ComplianceResult{Compliant: true} // Default positive if unknown
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

// initFolders ensures required directories exist
func (a *App) initFolders() {
	baseDir, err := a.getAppBaseDir()
	if err != nil {
		fmt.Printf("Error getting base dir: %v\n", err)
		return
	}

	dirs := []string{
		filepath.Join(baseDir, "data"),
		filepath.Join(baseDir, "data", "audit_queue"),
		filepath.Join(baseDir, "data", "audit_synced"),
		filepath.Join(baseDir, "logs"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Printf("Failed to create dir %s: %v\n", d, err)
		}
	}
}

// CheckInternet checks if internet is available
func (a *App) CheckInternet() bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://google.com")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// SyncAuditResults processes the offline queue
func (a *App) SyncAuditResults() {
	if !a.CheckInternet() {
		return
	}

	baseDir, err := a.getAppBaseDir()
	if err != nil {
		return
	}

	queueDir := filepath.Join(baseDir, "data", "audit_queue")
	syncedDir := filepath.Join(baseDir, "data", "audit_synced")

	files, err := os.ReadDir(queueDir)
	if err != nil {
		return
	}

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(queueDir, f.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Perform HTTP POST
		url := "https://script.google.com/macros/s/AKfycbymX8A88tzF7NgERlL6eoZ97uO03A_lLhTGBQHPc4fipwHwBw7At9aahBkVpYOMTxgw-w/exec"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
				// Move to synced
				newPath := filepath.Join(syncedDir, f.Name())
				os.Rename(filePath, newPath)
			}
		}
	}
}

// SaveAuditResult saves the audit data as JSON
func (a *App) SaveAuditResult(frontendData map[string]interface{}) error {
	baseDir, err := a.getAppBaseDir()
	if err != nil {
		return err
	}

	queueDir := filepath.Join(baseDir, "data", "audit_queue")
	
	// Ensure directory exists just in case
	os.MkdirAll(queueDir, 0755)

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Unknown"
	}

	// Add backend info
	frontendData["hostname"] = hostname
	frontendData["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	frontendData["os"] = runtime.GOOS
	
	// Create JSON
	jsonData, err := json.MarshalIndent(frontendData, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("audit_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(queueDir, filename)

	return os.WriteFile(filepath, jsonData, 0644)
}
