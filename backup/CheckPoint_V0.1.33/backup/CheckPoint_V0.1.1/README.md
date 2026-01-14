# CheckPoint – Secure Device Checker

![Version](https://img.shields.io/badge/version-v0.1.1-blue.svg) ![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows-lightgrey.svg)

> **A lightweight, cross-platform desktop application for performing local device security and network diagnostics without opening external terminals.**

---

## 📖 Table of Contents
- [Overview](#overview)
- [Key Features](#key-features)
- [System Requirements](#system-requirements)
- [Installation](#installation)
- [Usage](#usage)
- [CLI Mode](#cli-mode)
- [Backup & Versioning](#backup--versioning)
- [Development](#development)
- [Security & Privacy](#security--privacy)
- [License](#license)

---

## 🔎 Overview

**CheckPoint** is a secure, offline diagnostic utility designed for IT administrators, SOC analysts, and security engineers. It allows users to run essential system, network, and security checks through a unified, clean interface, eliminating the need to memorize complex terminal commands or manage multiple command prompt windows.

**Key Design Principles:**
- **Local Execution:** All tools run on the host machine; no data is sent to the cloud.
- **Unified Interface:** Terminal output is integrated directly into the application window.
- **No Pop-ups:** Commands execute silently in the background, keeping the workspace clutter-free.
- **Minimalist UI:** A professional "Command Center" aesthetic designed for efficiency.

---

## 🛠 Key Features

### 📡 Network Diagnostics
- **IP Configuration:** View active IP, subnet, and gateway details.
- **Routing & Connectivity:** Trace routes and check connection latency.
- **Interface Inspection:** Analyze active network interfaces and ARP tables.

### 💻 System & Application Checks
- **Startup Items:** Inspect programs configured to launch at boot.
- **Drive & Storage:** List mounted drives and usage.
- **Service Status:** Overview of running system services and background processes.

### 🛡 Security & Malware Awareness
- **Security Tools:** Check the status of built-in OS security features (e.g., Gatekeeper, SIP on macOS).
- **Antivirus Status:** Verify the presence and status of active antivirus solutions (e.g., XProtect, Windows Defender).

### 📝 Logging & Audit
- **Real-time Output:** View command execution logs instantly in the integrated terminal pane.
- **Export Logs:** Save diagnostic sessions to `.txt` files for reporting and auditing.
- **History:** Timestamped execution history for every session.

### 🖥 UI / UX
- **Categorized Dashboard:** Expandable "Accordion" style menu for organized access to tools.
- **Split View:** Dedicated tool selection pane (left) and terminal output pane (right).
- **Status Indicators:** Dynamic button states (Run -> Running -> Done) for clear feedback.

---

## ⚙️ System Requirements

- **Operating System:**
    - **macOS:** version 11.0 (Big Sur) or later.
    - **Windows:** 10 / 11 (x64).
- **Architecture:** Apple Silicon (M1/M2/M3) or Intel (amd64).
- **Connectivity:** No internet connection required for core diagnostic functions.

---

## 📥 Installation

### macOS (.dmg)
1.  **Download** the `CheckPoint_v0.1.1.dmg` file.
2.  **Open** the disk image.
3.  **Drag-and-Drop** `CheckPoint.app` into the `Applications` folder shortcut provided.
4.  **First Run:**
    - Locate **CheckPoint** in your Applications folder.
    - *Note:* If prompted by Gatekeeper/Security settings, Right-Click the app and select **Open**, then confirm to run.

### Windows (.exe)
1.  **Download** the executable installer.
2.  **Run** the installer or portable `.exe`.
3.  **Permissions:** Allow the application to run via Windows Defender SmartScreen if prompted, as it is a diagnostic tool requiring system access.

---

## 🚀 Usage

1.  **Launch** the application.
2.  **Select a Category** from the left sidebar (e.g., "Network", "Security").
3.  **Run a Tool** by clicking the **Run** button next to the desired diagnosis.
4.  **View Results** in the black terminal pane on the right.
5.  **Export Data** by clicking **Export Logs** in the top toolbar to save the current session output.
6.  **Reset** the session (clears logs and tool states) using the **Reset** button.

---

## 💻 CLI Mode

For advanced users or debugging purposes, CheckPoint can be launched from the command line. This allows you to see standard output streams if necessary, though all application logs are mirrored in the GUI.

**Example:**
```bash
# macOS
./Applications/CheckPoint.app/Contents/MacOS/CheckPoint

# Windows
.\CheckPoint.exe
```

---

## 📦 Backup & Versioning

To maintain a secure development lifecycle and allow for rollbacks:

**Recommended Folder Structure:**
```
CheckPoint_Backups/
├── v0.1.0/
├── v0.1.1/
└── current/
```

**Rollback Strategy:**
- Keep previous `.dmg` or `.zip` releases in a backup directory.
- To roll back, simply delete the current installation from `/Applications` and re-install the previous version.

---

## 👨‍💻 Development

Built with **Go** and **Wails**, combining the performance of Go with a web-based frontend.

**Prerequisites:**
- Go 1.20+
- Node.js 16+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

**Build Command:**
```bash
# Development Mode
wails dev

# Production Build (Mac Universal)
./package_mac.sh v0.1.1
```

---

## 🔐 Security & Privacy

CheckPoint is designed with a "Security First" mindset:
- **Local Only:** No diagnostic data, logs, or telemetry are sent to any external server.
- **Read-Only Diagnostics:** Most tools perform read-only checks (querying status, listing files) and do not modify system settings (unless explicitly designed for cleanup, e.g., opening Temp folders).
- **Transparency:** All commands executed are standard OS utilities (`ping`, `netstat`, `ifconfig`, `launchctl`, etc.).

---

## 📄 License & Disclaimer

**© 2025 CheckPoint – Secure Device Checker.**
Developed by **Tim Proteksi, Direktorat Operasi Keamanan Siber**.

All trademarks, logos, and brand names are the property of their respective owners. All company, product and service names used in this software are for identification purposes only.

> **Disclaimer:** This tool is intended for internal use by authorized administrators. It is a diagnostic aid and does not replace enterprise-grade endpoint protection or forensic tools. Use responsibly.
