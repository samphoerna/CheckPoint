// Categories Configuration
const categories = [
    {
        id: "network",
        title: "NETWORK",
        icon: "📡",
        tools: [
            "Cek IP",
            "Cek Routing",
            "Netstat",
            "ARP Table",
            "Ping Connectivity"
        ]
    },
    {
        id: "application_system",
        title: "APPLICATION / SYSTEM",
        icon: "💻",
        tools: [
            "List PS Drives",
            "Access HKLM Registry",
            "Startup Registry Check"
        ]
    },
    {
        id: "malware_antivirus",
        title: "MALWARE / ANTI VIRUS",
        icon: "🛡️",
        tools: [
            "Microsoft Malware Removal Tool",
            "Check Default Antivirus Status"
        ]
    },
    {
        id: "remote_services",
        title: "REMOTE SERVICES",
        icon: "🔗",
        tools: [
            "Windows Services",
            "Remote System Properties",
            "Device Manager (Bluetooth)",
            "Registry Editor",
            "Task Manager",
            "Startup Services"
        ]
    },
    {
        id: "clean_files",
        title: "CLEAN FILES",
        icon: "🧹",
        tools: [
            "Open Temp Folder",
            "Open Trash / Recycle Bin",
            "Open Microsoft Office Temp Files"
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
                <span>${cat.icon}</span>
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
                <div class="tool-info">
                    <div class="tool-name">${toolName}</div>
                </div>
                <button class="run-btn" id="btn-${toolId}">Run</button>
            `;

            body.appendChild(row);

            // Bind State
            const btn = row.querySelector(`#btn-${toolId}`);
            btn.onclick = () => runTool(toolName, btn);

            toolControls[toolName] = { button: btn, state: 'idle' };
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
