# CheckPoint – Secure Device Checker

![Version](https://img.shields.io/badge/version-v0.1.33a-blue.svg) ![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS-lightgrey.svg)

> **Secure Device Checker & Network Diagnostic Tool**

CheckPoint is a secure, offline diagnostic utility designed for IT administrators, SOC analysts, and security engineers. It allows users to run essential system, network, and security checks through a unified, clean interface, eliminating the need to memorize complex terminal commands or manage multiple command prompt windows.

---

## 📖 Table of Contents
1. [Project Title](#checkpoint--secure-device-checker)
2. [Short Description](#-short-description)
3. [Key Features](#-key-features)
4. [Supported Platforms](#-supported-platforms)
5. [System Requirements](#-system-requirements)
6. [Installation Guide](#-installation-guide)
7. [How to Run](#-how-to-run-cli--gui)
8. [Feature Categories & Functions](#-feature-categories--functions)
9. [Screenshot & Logging Behavior](#-screenshot--logging-behavior)
10. [Versioning & Backup Strategy](#-versioning--backup-strategy)
11. [Build & Release Notes](#-build--release-notes)
12. [Security & Permission Notes](#-security--permission-notes)
13. [Troubleshooting](#-troubleshooting)
14. [Disclaimer](#-disclaimer)
15. [License / Ownership](#-license--ownership)

---

## 📝 Short Description
**CheckPoint** is a lightweight, cross-platform desktop application for performing local device security and network diagnostics. It allows users to execute standard OS commands (like `ping`, `netstat`, `ipconfig`) via a GUI, displaying output in an integrated terminal window. It is designed to be **safe**, **transparent**, and **privacy-focused**, operating entirely offline without sending data to external servers.

---

## 🛠 Key Features
- **Unified Dashboard:** Access network, system, and security tools from a single interface.
- **Cross-Platform:** Works on Windows 10/11 and macOS (Intel & Apple Silicon).
- **Silent Execution:** Runs diagnostic commands in the background without popping up annoying terminal windows.
- **Integrated Terminal:** View real-time command output directly within the application.
- **Privacy First:** No telemetry, no cloud connections. All data stays local.
- **Portable:** No installation required; runs directly from a USB stick or local folder.

---

## 💻 Supported Platforms
- **Windows:** Windows 10, Windows 11 (x64)
    - *Format:* Portable Executable (`.exe`)
- **macOS:** macOS Big Sur (11.0) and later
    - *Format:* Application Bundle (`.app`), Portable

---

## ⚙️ System Requirements
- **Processor:** Intel Core i3 / AMD Ryzen 3 or better; Apple M1/M2/M3.
- **RAM:** 4GB minimum.
- **Disk Space:** ~50MB.
- **Permissions:** Admin privileges may be required for specific low-level checks (e.g., specific registry keys or system files).

---

## 📥 Installation Guide

### Windows (.exe)
1.  **Download** the `CheckPoint.exe` file.
2.  **Place** it in a folder of your choice (e.g., `C:\Tools\CheckPoint`).
3.  **No Installation Needed:** Just double-click to run.

### macOS (.app)
1.  **Download** the `CheckPoint_v0.1.33a.dmg` image (or `.zip`).
2.  **Open** the DMG file.
3.  **Drag-and-Drop** `CheckPoint.app` into your `Applications` folder or a portable drive.
4.  **First Launch:**
    - Right-click `CheckPoint.app`.
    - Select **Open**.
    - Click **Open** again to bypass Gatekeeper warning (first time only).

---

## 🚀 How to Run (CLI & GUI)

### GUI Mode (Standard)
simply double-click the application icon. The dashboard will load with the "Network" category selected by default.

### CLI Mode (Debug)
For advanced users, you can launch CheckPoint from a terminal to see standard output/error streams useful for debugging.

**Windows:**
```powershell
.\CheckPoint.exe
```

**macOS:**
```bash
./CheckPoint.app/Contents/MacOS/CheckPoint
```

---

## 🧰 Feature Categories & Functions

CheckPoint organizes tools into five logical categories:

### A. Network
*Diagnostics for connectivity and local network configuration.*
- **Cek IP:** Displays IP address, subnet mask, and gateway (runs `ipconfig /all` or `ifconfig`).
- **Cek Routing:** Traces the route to a common public DNS (8.8.8.8) to verify internet path.
- **Netstat:** Lists all active TCP/UDP connections and listening ports.
- **ARP Table:** Shows the mapping of IP addresses to physical MAC addresses on the local network.
- **Ping Connectivity:** Checks basic reachability to the internet (8.8.8.8).

### B. Application / System
*Hardware, OS, and installed software inspection.*
- **System Information:** Displays Hostame, OS version, Architecture, and BIOS details. Opens System Settings/About.
- **Check installed applications:** Lists installed software via WMI (Windows) or System Profiler (macOS).
- **List PS Drives:** Shows all mounted drives and volume usage.
- **Access HKLM Registry:** (Windows) Checks critical registry paths. (macOS) Reads global defaults.
- **Startup Services:** Lists services configured to start automatically.
- **Registry Check (Startup):** Inspects startup commands defined in the Registry/LaunchAgents.
- **Registry Editor:** Opens `regedit` (Windows) or `Preferences` (macOS) for manual inspection (triggers Screenshot).
- **Task Manager:** Lists top CPU-consuming processes and opens the native Task Manager / Activity Monitor (triggers Screenshot).

### C. Malware / Anti Virus
*Status of built-in protection engines.*
- **Security Status:** Checks Windows Defender / Gatekeeper status.
- **Protection Health:** Verifies last update time and real-time protection status.
- **Run Quick Scan:** Initiates a **Windows Defender Quick Scan** (Windows) or inspects persistence folders (macOS).

### D. Remote Services
*Detection of risky open ports and browser extensions.*
- **Check active network service ports:** Scans localhost for open ports: FTP (21), SSH (22), SMB (445), RDP (3389).
- **Check installed browser extensions:** Enumerates extensions for:
    - Google Chrome
    - Microsoft Edge
    - Mozilla Firefox
    - Safari (macOS only)
    - *Behavior:* Sequential processing with headers in the output terminal.

### E. Clean Files
*System cleanup utilities.*
- **Run Full Cleanup:** Executes a pre-defined sequence of cleanup tasks (Temp, Trash, Office Temps).
- **Open Temp Folder:** Calculates size and opens the system temporary directory.
- **Open Trash / Recycle Bin:** Calculates size and opens the Trash/Recycle Bin.
- **Open Office Temp Files:** Locates and opens the AutoRecovery folder for Microsoft Word.

---

## 📸 Screenshot & Logging Behavior

### Screenshot (Windows Only)
- **Status:** **ACTIVE on Windows**, DISABLED on macOS.
- **Trigger:** Screenshots are automatically taken when specific features are run that open external windows (e.g., "Task Manager", "Registry Editor", "System Information").
- **Behavior:**
    - **Delayed Capture:** The system waits 2 seconds after the button click to ensure the external window is visible.
    - **Full Screen:** Captures the entire primary display.
- **Storage Location:**
    - Folder: `CP-SS-<DDMMYYYY>` (e.g., `CP-SS-14012026`) inside the **application's directory**.
    - If running from a read-only source (e.g., CD-ROM), it falls back to the system `%TEMP%` directory.

### Logging
- **Console Log:** All command output is visible in the right-hand black terminal pane.
- **Export:** Click the **Export Logs** button (top right) to save the current session to a `.txt` file.
    - Default Location: Sibling directory of the application.
    - Filename: `checkpoint-log-<YYYYMMDD-HHMMSS>.txt`.

---

## 📦 Versioning & Backup Strategy

### Current Version: `V.0.1.33a`

### Upgrade Workflow
1.  **Backup:** Before upgrading, the existing version is backed up.
2.  **Naming Convention:** Backups are stored in a dedicated folder.
    - Format: `/backup/CheckPoint_<Version>`
    - Example: `/backup/CheckPoint_V0.1.33`
3.  **Restore:** To restore, simply delete the current executable and copy the backup file back to the main directory.

---

## 🏗 Build & Release Notes

- **Portable Build:** The application is compiled as a single binary (Windows) or bundle (macOS).
- **Dependencies:**
    - No external runtimes required (Go runtime is embedded).
    - Windows: Depends on standard PowerShell libraries.
    - macOS: Depends on standard Unix utilities.
- **Cross-Compilation:** Can be built for Windows from macOS and vice-versa (using `wails build -platform ...`).

---

## 🔐 Security & Permission Notes

- **Admin Rights:**
    - **Windows:** It is recommended to run as Administrator to allow WMI queries and System Cleanup to function effectively.
    - **macOS:** Standard user permissions suffice for most checks. `sudo` is not invoked by the app; if a command fails due to permission, it logs the error.
- **Safety:**
    - The "Clean Files" module deletes files from **Temp** and **Recycle Bin** only. It does not touch user documents or system files.
    - No external network calls are made except for connectivity checks (Ping/Trace to 8.8.8.8).

---

## ❓ Troubleshooting

| Issue | Platform | Solution |
| :--- | :--- | :--- |
| **"Windows protected your PC"** | Windows | Click "More info" > "Run anyway". App is not code-signed yet. |
| **"App cannot be opened"** | macOS | Right-click the App > Open > Open. Validate inside System Settings > Privacy & Security. |
| **Screenshot not saved** | Windows | Check write permissions in the folder. If running from CD, check `%TEMP%`. |
| **Empty Output in Terminal** | All | Some commands (like Netstat) may take a few seconds. Wait for "Process completed successfully". |

---

## ⚠️ Disclaimer

This tool is provided "as is" without warranty of any kind. It is intended for use by authorized security personnel for diagnostic purposes. The developers are not responsible for any data loss or system instability resulting from the use of the "Clean Files" or system modification features.

---

## 📜 License / Ownership

**© 2025 CheckPoint – Secure Device Checker.**
Developed by **Tim Proteksi, Direktorat Operasi Keamanan Siber**.

All trademarks, product names, and company names or logos cited herein are the property of their respective owners.
