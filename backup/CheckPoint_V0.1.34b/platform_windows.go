package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}

func cleanupWindows(logFunc func(string)) {
	logFunc("[Clean Files] Starting Windows cleanup...")

	// Helper to run PS command
	runPS := func(desc, cmd string) {
		logFunc(fmt.Sprintf("  - %s...", desc))
		// We use -ErrorAction SilentlyContinue to avoid spamming errors for locked files
		fullCmd := fmt.Sprintf("powershell -NoProfile -NonInteractive -Command \"%s\"", cmd)
		c := exec.Command("cmd", "/C", fullCmd)
		c.SysProcAttr = getSysProcAttr()
		out, err := c.CombinedOutput()
		if err != nil {
			logFunc(fmt.Sprintf("    [INFO] %s (Partial/Locks)", desc))
		} else if len(out) > 0 {
			logFunc(fmt.Sprintf("    [OK] %s", string(out)))
		} else {
			logFunc("    [OK] Completed.")
		}
	}

	// 1. Recycle Bin
	runPS("Emptying Recycle Bin", "Clear-RecycleBin -Force -ErrorAction SilentlyContinue")

	// 2. Temp Folders (%temp%, temp)
	runPS("Cleaning User Temp (%TEMP%)", "Get-ChildItem -Path $env:TEMP -Recurse -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue")
	runPS("Cleaning System Temp (C:\\Windows\\Temp)", "Get-ChildItem -Path 'C:\\Windows\\Temp' -Recurse -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue")

	// 3. Browser Cache (Chrome, Edge, Firefox)
	// Chrome
	runPS("Cleaning Chrome Cache", "Get-ChildItem -Path \"$env:LOCALAPPDATA\\Google\\Chrome\\User Data\\Default\\Cache\\*\" -Recurse -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue")
	// Edge
	runPS("Cleaning Edge Cache", "Get-ChildItem -Path \"$env:LOCALAPPDATA\\Microsoft\\Edge\\User Data\\Default\\Cache\\*\" -Recurse -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue")

	// 4. Autosave / Recent
	runPS("Cleaning Recent Items", "Get-ChildItem -Path \"$env:APPDATA\\Microsoft\\Windows\\Recent\\*\" -Recurse -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force -ErrorAction SilentlyContinue")

	// 5. Thumbnails (Explorer) - Requires stopping explorer usually, skipping to avoid UI crash.
	logFunc("  - Cleaning Thumbnails (Skipping to prevent explorer restart)")

	// 6. Exam/Documents Check
	logFunc("[Clean Files] Checking Documents/Desktop for 'exam' files...")
	// Just looking for keyword 'exam' or 'ujian' to be safe/helpful
	runPS("Scanning Desktop", "Get-ChildItem -Path $env:USERPROFILE\\Desktop -Filter *exam* -Recurse -ErrorAction SilentlyContinue | Select-Object Name")
	runPS("Scanning Documents", "Get-ChildItem -Path $env:USERPROFILE\\Documents -Filter *exam* -Recurse -ErrorAction SilentlyContinue | Select-Object Name")

	// 7. Removable Drives
	logFunc("[Clean Files] Checking Removable Drives...")
	runPS("Listing Removable Drives", "Get-CimInstance -ClassName Win32_LogicalDisk -Filter \"DriveType = 2\" | Select-Object DeviceID, VolumeName")

	logFunc("[Clean Files] Windows Cleanup Summary Complete.")
}

func cleanupMac(logFunc func(string)) {}
