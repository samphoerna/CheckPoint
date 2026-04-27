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
        id: "remote_services",
        title: "Remote Services",
        icon: "🔗",
        tools: [
            "FTP Service",
            "SSH Service",
            "Telnet Service",
            "SMB Service",
            "RDP Service",
            "Bluetooth",
            "File Sharing"
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

// Calculate scores state array
let complianceScores = {};

// Display name → backend feature name mapping for Remote Services
const serviceToBackend = {
    "FTP Service": "Port 21",
    "SSH Service": "Port 22",
    "Telnet Service": "Port 23",
    "SMB Service": "Port 445",
    "RDP Service": "Port 3389"
};

// Helper to update the score display on the UI
window.updateGlobalScore = function () {
    const scoreDisplay = document.getElementById('global-score-display');
    if (!scoreDisplay) return;

    // If no audit data has been recorded yet, show muted placeholder
    if (Object.keys(complianceScores).length === 0) {
        scoreDisplay.className = "score-muted";
        scoreDisplay.innerHTML = `Compliance Score: <span style="color:#8E8E93">—</span> <span style="color:#D1D1D6; font-weight: 400;">|</span> <span style="color:#8E8E93">N/A</span>`;
        return;
    }

    let totalPoints = Object.values(complianceScores).reduce((sum, val) => sum + val, 0);
    let percentage = Math.round((totalPoints / 22) * 100);

    let statusText = "WORSE";
    let color = "#FF453A"; // default RED

    if (totalPoints >= 18) {
        statusText = "PASS";
        color = "#34C759"; // GREEN
    } else if (totalPoints >= 15) {
        statusText = "OK";
        color = "#007AFF"; // BLUE
    } else if (totalPoints >= 12) {
        statusText = "BAD";
        color = "#FF9F0A"; // ORANGE
    }

    scoreDisplay.className = "";
    scoreDisplay.innerHTML = `Compliance Score: <span style="color:${color}">${percentage}%</span> <span style="color:#D1D1D6; font-weight: 400;">|</span> <span style="color:${color}">${statusText}</span>`;
};

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

            let manualBtnsHTML = '';
            let btnGroupStyle = "display: flex; gap: 4px; align-items: center;";

            if (cat.id === "application_system") {
                btnGroupStyle = "display: flex; flex-direction: column; gap: 4px; width: 70px;";
                manualBtnsHTML = `
                    <div style="display: flex; gap: 4px; width: 100%;">
                        <button class="eval-btn eval-pass" id="pass-${toolId}" style="flex: 1; min-width: 0;">PASS</button>
                        <button class="eval-btn eval-fail" id="fail-${toolId}" style="flex: 1; min-width: 0;">FAIL</button>
                    </div>
                `;
            }

            row.innerHTML = `
                <div class="tool-info" style="flex: 1;">
                    <div class="tool-name">${toolName}</div>
                </div>
                <div class="btn-group" style="${btnGroupStyle}">
                    <button class="run-btn reset-tool-btn" id="reset-${toolId}" style="display: none; background: #FF453A; color: white; width: 100%;">Reset</button>
                    <button class="run-btn" id="btn-${toolId}" style="${cat.id === 'application_system' ? 'width: 100%;' : ''}">Run</button>
                    ${manualBtnsHTML}
                </div>
                <div class="repair-text" id="repair-${toolId}" style="display: none; width: 100%; color: #FF453A; font-size: 0.85em; margin-top: 8px; padding-left: 8px; border-left: 2px solid #FF453A;"></div>
            `;

            // To allow the repair text to sit on a new line within the flex row
            row.style.flexWrap = 'wrap';

            // Insert Manual Validation Handlers if category is App/System
            if (cat.id === "application_system") {
                const passBtn = row.querySelector(`#pass-${toolId}`);
                const failBtn = row.querySelector(`#fail-${toolId}`);

                passBtn.onclick = () => {
                    complianceScores[toolName] = 1;
                    passBtn.classList.add('active');
                    failBtn.classList.remove('active');
                    window.updateGlobalScore();
                };

                failBtn.onclick = () => {
                    complianceScores[toolName] = 0;
                    failBtn.classList.add('active');
                    passBtn.classList.remove('active');
                    window.updateGlobalScore();
                };
            }

            body.appendChild(row);

            // Backend alias for duplicate feature names and service label mapping
            let backendFeatureName = toolName;
            if (cat.id === "security_antivirus" && toolName === "Security Status") {
                backendFeatureName = "Sys_Security_Status";
            } else if (serviceToBackend[toolName]) {
                backendFeatureName = serviceToBackend[toolName];
            }

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
        let totalPoints = Object.values(complianceScores).reduce((sum, val) => sum + val, 0);
        let percentage = Math.round((totalPoints / 22) * 100);
        let hasData = Object.keys(complianceScores).length > 0;
        let statusText = hasData ? "WORSE" : "N/A";
        if (totalPoints >= 18) statusText = "PASS";
        else if (totalPoints >= 15) statusText = "OK";
        else if (totalPoints >= 12) statusText = "BAD";

        const scoreHeader = `=====================================\n[ Compliance Score: ${hasData ? totalPoints + '/22 - ' + percentage + '% - ' + statusText : 'N/A'} ]\n=====================================\n\n`;
        const exportContent = scoreHeader + logBuffer;

        if (window.go && window.go.main && window.go.main.App && window.go.main.App.ExportLogs) {
            window.go.main.App.ExportLogs(exportContent).catch(err => {
                appendLog(`[ERROR] Save failed: ${err}`);
            });
        }

        // Save audit JSON for Sync Queue
        if (window.go && window.go.main && window.go.main.App && window.go.main.App.SaveAuditResult) {
            const mapKey = (frontendName) => {
                const mapping = {
                    "Internet Connectivity": "internet",
                    "Public IP Check": "public_ip",
                    "Firewall": "firewall",
                    "VPN Connection": "vpn_connection",
                    "Proxy State": "proxy",
                    "Antivirus": "antivirus",
                    "Security Status": "security",
                    "Sys_Security_Status": "security",
                    "Port 21": "port21",
                    "Port 22": "port22",
                    "Port 23": "port23",
                    "Port 445": "port445",
                    "Port 3389": "port3389",
                    "FTP Service": "port21",
                    "SSH Service": "port22",
                    "Telnet Service": "port23",
                    "SMB Service": "port445",
                    "RDP Service": "port3389",
                    "Bluetooth": "bluetooth",
                    "File Sharing": "file_sharing"
                };
                return mapping[frontendName] || frontendName.toLowerCase().replace(/[^a-z0-9]/g, '_');
            };

            let payload = {
                internet: false,
                public_ip: false,
                firewall: false,
                vpn_connection: false,
                proxy: false,
                antivirus: false,
                security: false,
                port21: false,
                port22: false,
                port23: false,
                port445: false,
                port3389: false,
                bluetooth: false,
                file_sharing: false,
                score: totalPoints,
                percent: percentage + "%",
                status: statusText
            };

            Object.keys(complianceScores).forEach(key => {
                let jsonKey = mapKey(key);
                if (jsonKey && payload.hasOwnProperty(jsonKey)) {
                    payload[jsonKey] = complianceScores[key] === 1;
                }
            });

            window.go.main.App.SaveAuditResult(payload).then(() => {
                appendLog(`[OK] Offline Audit data exported to queue.`);
            }).catch(err => {
                appendLog(`[ERROR] Offline Audit save failed: ${err}`);
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
            "Sys_Security_Status",
            "Port 21",
            "Port 22",
            "Port 23",
            "Port 445",
            "Port 3389",
            "Bluetooth",
            "File Sharing"
        ];

        if (complianceFeatures.includes(featureName) && window.go && window.go.main && window.go.main.App && window.go.main.App.CheckNetworkCompliance) {
            window.go.main.App.CheckNetworkCompliance(featureName).then(result => {
                const toolId = featureName.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase();
                const resetBtn = document.getElementById(`reset-${toolId}`);
                const repairDiv = document.getElementById(`repair-${toolId}`);

                // Update scoring matrix
                complianceScores[featureName] = result.compliant ? 1 : 0;
                window.updateGlobalScore();

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

        const passBtn = document.getElementById(`pass-${toolId}`);
        if (passBtn) passBtn.classList.remove('active');
        const failBtn = document.getElementById(`fail-${toolId}`);
        if (failBtn) failBtn.classList.remove('active');
    });

    // Reset compliance scores
    complianceScores = {};
    window.updateGlobalScore();

    // Clear logs
    logContainer.innerHTML = '';
    logBuffer = "";
    const msg = document.createElement('div');
    msg.className = 'log-line system-msg';
    msg.textContent = 'System Reset Complete. Ready.';
    logContainer.appendChild(msg);
}

init();
