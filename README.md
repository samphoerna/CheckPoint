# CheckPoint – Secure Device Checker

![Version](https://img.shields.io/badge/version-v0.1.34-blue.svg) ![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS-lightgrey.svg)

> **Secure Device Checker & Network Diagnostic Tool**

CheckPoint is a secure, offline diagnostic utility designed for IT administrators, SOC analysts, and security engineers. It allows users to run essential system, network, and security checks through a unified, clean interface, eliminating the need to memorize complex terminal commands or manage multiple command prompt windows.

---

## Table of Contents
1. [Project Title](#checkpoint--secure-device-checker)
2. [Short Description](#-short-description)
3. [Key Features](#-key-features)
4. [Supported Platforms](#-supported-platforms)
5. [Installation Guide](#-installation-guide)
6. [Feature Categories & Functions](#-feature-categories--functions)
7. [Workflow](#-workflow)
8. [Disclaimer](#-disclaimer)
9. [License / Ownership](#-license--ownership)

---

## Short Description
**CheckPoint** is a lightweight, cross-platform desktop application designed for local device security auditing, network diagnostics, and compliance tracking. It safely executes system commands via a GUI and displays real-time output in an integrated terminal pane. Additionally, CheckPoint generates global compliance scores.

---

## Key Features
- **Unified Dashboard:** Access network, system, and security tools from a single interface.
- **Cross-Platform:** Works on Windows 10/11 and macOS (Intel & Apple Silicon).
- **Integrated Terminal:** View real-time command output directly within the application.
- **Portable:** No installation required; runs directly from a USB stick or local folder.

### Compliance Score
Score Formula (%) = (Total Poin / 22) * 100

Score Criteria =

| Total Poin | Percent | Status | Colour  |
| :--------- | :--------- | :----- | :----- |
| 17 - 22    | ≥ 80%      | Pass   | Green  |
| 14 - 16    | 65 - 79%   | OK     | Yellow |
| 10 - 13    | 50 - 64%   | Bad    | Orange |
| 0 - 9      | < 50%      | Worse  | Red    |

---

## Supported Platforms
- **Windows:** Windows 10, Windows 11 (x64)
    - *Format:* Portable Executable (`.exe`)
- **macOS:** macOS Big Sur (11.0) and later
    - *Format:* Application Bundle (`.app`), Portable
- **Permissions:** Admin privileges may be required for specific low-level checks (e.g., specific registry keys or system files).

---

## 📥 Installation Guide

### How to Run via Command (Dev)
export PATH=$PATH:$(go env GOPATH)/bin && wails dev

### How to Build via Command (onliner)
export PATH=$PATH:$(go env GOPATH)/bin && \
wails build -platform windows/amd64 -ldflags "-s -w -X main.AppVersion=[Version]" -clean && \
mv "build/bin/CheckPoint.exe" "build/bin/CheckPoint_[Version].exe" && \
wails build -platform darwin/universal -ldflags "-s -w -X main.AppVersion=[Version]" && \
mv "build/bin/CheckPoint.app" "build/bin/CheckPoint_[Version].app"

### How to Build .dmg via Command (onliner)
./package_mac.sh [Version] && mv CheckPoint_[Version].dmg build/bin/

---

## Feature Categories & Functions

CheckPoint organizes tools into five logical categories:

### A. Network
*Diagnostics and compliance for connectivity and networking profiles.*
- **Internet Connectivity:** Verifies if the device has active internet access.
- **Public IP Check:** Polls api.ipify.org to track external IP addresses.
- **VPN Connection:** Inspects whether the system is connected to a Virtual Private Network.
- **Firewall:** Checks the global firewall profile (Windows Firewall or macOS Application Firewall).
- **Proxy State:** Identifies enabled system-wide internet proxies.

### B. Application / System
*Hardware, OS, and software management inspection.*
- **System Information:** Displays Hostname, OS version, Architecture, and BIOS details. 
- **IP Check:** Queries `ipconfig` / `ifconfig` for local interface properties.
- **Network Adapter:** Lists detailed hardware network interfaces.
- **Task Manager:** Opens native Task Manager / Activity Monitor and lists active processes (triggers Screenshot on Windows).
- **Installed App:** Lists installed software via WMI (Windows) or System Profiler (macOS).
- **Startup Services:** Queries startup configuration commands and tools.
- **Remote Access Setting:** Inspects RDP/Remote Login configuration.
- **Browser Extension:** Enumerates extensions for Google Chrome and Edge.

### C. Remote Services
*Detection of risky open ports and system sharing.*
- **Port 21:** Checks if FTP services are active and listening.
- **Port 22:** Checks if SSH services are active and listening.
- **Port 23:** Checks if Telnet services are active and listening.
- **Port 445:** Checks if SMB/CIFS shares are active and listening.
- **Port 3389:** Checks if RDP services are active and listening.
- **Bluetooth:** Checks the status of the local Bluetooth adapter.
- **File Sharing:** Verifies local file sharing or near-share channels are securely configured.

### D. Security & Antivirus
*Status of built-in protection engines and configurations.*
- **Antivirus:** Detects Antivirus product installations (WMI namespace or XProtect).
- **Sys_Security_Status:** Validates core OS security systems like Defender RealTimeProtection, Gatekeeper, and SIP status.

### E. Clean Files
*System cleanup utilities.*
- **Run Full Cleanup:** Executes a pre-defined sequence of cleanup tasks (Temp, Trash, Office Temps).
- **Open Temp Folder:** Calculates size and opens the system temporary directory (triggers Screenshot on Windows).
- **Open Trash / Recycle Bin:** Calculates size and opens the Trash/Recycle Bin (triggers Screenshot on Windows).
- **Open Office Temp Files:** Locates and opens the AutoRecovery folder for Microsoft Word (triggers Screenshot on Windows).

---

## Workflow

in Offline/Online PC
   ↓
Run CheckPoint Tool
   ↓
Output Audit JSON
   ↓
Rerun app in Operator PC (Online)
   ↓
Trigger HTTP POST
   ↓
Google Spreadsheet (Database)

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

## Disclaimer

This tool is provided "as is" without warranty of any kind. It is intended for use by authorized security personnel for diagnostic purposes. The developers are not responsible for any data loss or system instability resulting from the use of the "Clean Files" or system modification features.

---

## License / Ownership

**© 2025 CheckPoint – Secure Device Checker.**
Developed by **Tim Proteksi**.

All trademarks, product names, and company names or logos cited herein are the property of their respective owners.
