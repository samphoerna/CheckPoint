// Categories Configuration
const categories = [
    {
        id: "application_system",
        title: "Application / System",
        icon: "💻",
        tools: [
            "System Information",
            "IP Check",
            "Network Adapter",
            "Task Manager",
            "Installed App",
            "Startup Services",
            "Remote Access Setting",
            "Browser Extension"
        ]
    },
    {
        id: "network",
        title: "Network",
        icon: "📡",
        tools: [
            "Internet Connectivity",
            "Public IP Check",
            "VPN Connection",
            "Firewall",
            "Proxy State"
        ]
    },
    {
        id: "security_antivirus",
        title: "Security & Antivirus",
        icon: "🔐",
        tools: [
            "Antivirus",
            "Security Status"
        ]
    },
    {
        id: "malware_antivirus",
        title: "Malware / Anti Virus",
        icon: "🛡️",
        tools: [
            "Security Status",
            "Protection Health",
            "Run Quick Scan"
        ]
    },
    {
        id: "remote_services",
        title: "Remote Services",
        icon: "🔗",
        tools: [
            "PORT 21 (FTP)",
            "PORT 22 (SSH)",
            "PORT 23 (TELNET)",
            "PORT 445 (SMB)",
            "PORT 3389 (RDP)",
            "BLUETOOTH",
            "FILE SHARING"
        ]
    },
    {
        id: "clean_files",
        title: "Clean Files",
        icon: "🧹",
        tools: [
            "Run Full Cleanup",
            "Open Temp Folder",
            "Open Trash / Recycle Bin",
            "Open Office Temp Files"
        ]
    }
];

// References
const runtime = window.runtime;
const toolsContainer = document.getElementById('toolsContainer');
const logContainer = document.getElementById('logContainer');
const clearBtn = document.getElementById('clearLogsBtn');
const exportBtn = document.getElementById('exportLogsBtn');
const globalResetBtn = document.getElementById('globalResetBtn');

// State Tracking
// Format: { "toolName": { button: HTMLElement, state: 'idle'|'running'|'done' } }
const toolControls = {};
let logBuffer = "";

// --- Logging ---

function appendLog(message, isRaw = false) {
    if (!message) return; // Ignore empty strings if any
    logBuffer += message + "\n";
    const entry = document.createElement('div');
    entry.className = 'log-line';
    if (!isRaw) {
        if (message.startsWith('===')) {
            entry.style.color = '#86868B'; // Secondary color for headers
        }
        else if (message.startsWith('[ERROR]')) {
            entry.style.color = '#FF453A'; // Red for errors
        }
        else if (message.startsWith('[OK]')) {
            entry.style.color = '#32D74B'; // Green for success
        }
    }
    entry.textContent = message;
    logContainer.appendChild(entry);
    logContainer.scrollTop = logContainer.scrollHeight;
}

// --- Initialization ---

function init() {
    categories.forEach(cat => {
        // Create Category Block
        const block = document.createElement('div');
        block.className = 'category-block expanded'; // Default expanded

        // Header
        const header = document.createElement('div');
        header.className = 'category-header';
        header.innerHTML = `
            <div class="cat-title">
                <span>${cat.title}</span>
            </div>
            <span class="chevron">▼</span>
        `;
        // Toggle Accordion
        header.onclick = () => {
            block.classList.toggle('expanded');
        };
        block.appendChild(header);

        // Body with Tools
        const body = document.createElement('div');
        body.className = 'category-body';

        cat.tools.forEach(toolName => {
            const row = document.createElement('div');
            row.className = 'tool-row';

            // Clean ID for tool
            const toolId = toolName.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase();

            row.innerHTML = `
                <div class="tool-info" style="flex: 1;">
                    <div class="tool-name">${toolName}</div>
                </div>
                <div class="btn-group" style="display: flex; gap: 4px;">
                    <button class="run-btn reset-tool-btn" id="reset-${toolId}" style="display: none; background: #FF453A; color: white;">Reset</button>
                    <button class="run-btn" id="btn-${toolId}">Run</button>
                </div>
                <div class="repair-text" id="repair-${toolId}" style="display: none; width: 100%; color: #FF453A; font-size: 0.85em; margin-top: 8px; padding-left: 8px; border-left: 2px solid #FF453A;"></div>
            `;
            
            // To allow the repair text to sit on a new line within the flex row
            row.style.flexWrap = 'wrap';

            body.appendChild(row);

            // Backend alias for duplicate feature names
            const backendFeatureName = (cat.id === "security_antivirus" && toolName === "Security Status") ? "Sys_Security_Status" : toolName;

            // Bind State
            const btn = row.querySelector(`#btn-${toolId}`);
            btn.onclick = () => runTool(backendFeatureName, btn);
            
            const resetBtn = row.querySelector(`#reset-${toolId}`);
            resetBtn.onclick = () => {
                const ctrl = toolControls[backendFeatureName];
                ctrl.state = 'idle';
                ctrl.button.textContent = 'Run';
                ctrl.button.classList.remove('running', 'done');
                ctrl.button.disabled = false;
                
                resetBtn.style.display = 'none';
                const repairDiv = document.getElementById(`repair-${toolId}`);
                if (repairDiv) {
                    repairDiv.style.display = 'none';
                    repairDiv.textContent = '';
                }
            };

            toolControls[backendFeatureName] = { button: btn, state: 'idle' };
        });

        block.appendChild(body);
        toolsContainer.appendChild(block);
    });

    // Global Events
    clearBtn.onclick = () => {
        logContainer.innerHTML = '';
        logBuffer = "";
        appendLog('--- Console Cleared ---', true);
    };

    exportBtn.onclick = () => {
        if (window.go && window.go.main && window.go.main.App && window.go.main.App.ExportLogs) {
            window.go.main.App.ExportLogs(logBuffer).catch(err => {
                appendLog(`[ERROR] Save failed: ${err}`);
            });
        }
    };

    globalResetBtn.onclick = resetAll;

    // Wails Events
    if (runtime) {
        runtime.EventsOn("log", (msg) => appendLog(msg));
        runtime.EventsOn("done", (featureName) => handleDone(featureName));

        // Fetch Version
        if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetAppVersion) {
            window.go.main.App.GetAppVersion().then(ver => {
                const verEl = document.getElementById('app-version');
                if (verEl) verEl.textContent = ver;
                appendLog(`System Ready (Version: ${ver})`);
            });
        }
    }
}

// --- Logic ---

function runTool(toolName, btn) {
    if (btn.disabled) return;

    // UI Update
    btn.textContent = 'Running...';
    btn.classList.add('running');
    btn.disabled = true;

    // Execute
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.ExecuteCommand) {
        window.go.main.App.ExecuteCommand(toolName);
    } else {
        // Mock (Fallback if backend unreachable in dev)
        appendLog(`[MOCK] Starting ${toolName}...`);
        setTimeout(() => handleDone(toolName), 1000);
    }
}

function handleDone(featureName) {
    // Note: featureName coming from backend must match the tool name exactly
    const ctrl = toolControls[featureName];
    if (ctrl) {
        ctrl.state = 'done';
        ctrl.button.textContent = 'Done';
        ctrl.button.classList.remove('running');
        ctrl.button.classList.add('done');
        // Keep disabled as per "persistent done state" requirement
        ctrl.button.disabled = true;

        // Run compliance check for specific Network tools if applicable
        const complianceFeatures = [
            "Internet Connectivity",
            "Public IP Check",
            "VPN Connection",
            "Firewall",
            "Proxy State",
            "Antivirus",
            "Security Status",
            "PORT 21 (FTP)",
            "PORT 22 (SSH)",
            "PORT 23 (TELNET)",
            "PORT 445 (SMB)",
            "PORT 3389 (RDP)",
            "BLUETOOTH",
            "FILE SHARING"
        ];
        
        if (complianceFeatures.includes(featureName) && window.go && window.go.main && window.go.main.App && window.go.main.App.CheckNetworkCompliance) {
            window.go.main.App.CheckNetworkCompliance(featureName).then(result => {
                const toolId = featureName.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase();
                const resetBtn = document.getElementById(`reset-${toolId}`);
                const repairDiv = document.getElementById(`repair-${toolId}`);
                
                if (!result.compliant) {
                    if (repairDiv && result.repairText) {
                        repairDiv.textContent = `Fix: ${result.repairText}`;
                        repairDiv.style.display = 'block';
                    }
                    if (resetBtn) {
                        resetBtn.style.display = 'inline-block';
                    }
                }
            }).catch(err => {
                console.error(`Error checking compliance for ${featureName}:`, err);
            });
        }
        
    } else {
        console.warn(`[Frontend] Received done event for unknown feature: ${featureName}`);
    }
}

function resetAll() {
    // Reset all tool buttons
    Object.keys(toolControls).forEach(key => {
        const ctrl = toolControls[key];
        ctrl.state = 'idle';
        ctrl.button.textContent = 'Run';
        ctrl.button.classList.remove('running', 'done');
        ctrl.button.classList.remove('error'); // If we had error state
        ctrl.button.disabled = false;
        
        const toolId = key.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase();
        const repairDiv = document.getElementById(`repair-${toolId}`);
        if (repairDiv) {
            repairDiv.style.display = 'none';
            repairDiv.textContent = '';
        }
        
        const resetBtn = document.getElementById(`reset-${toolId}`);
        if (resetBtn) {
            resetBtn.style.display = 'none';
        }
    });

    // Clear logs
    logContainer.innerHTML = '';
    logBuffer = "";
    const msg = document.createElement('div');
    msg.className = 'log-line system-msg';
    msg.textContent = 'System Reset Complete. Ready.';
    logContainer.appendChild(msg);
}

init();
