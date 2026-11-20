/* ==========================================================================
   FreeTab Pro — High-Precision Logic & Telemetry Engine
   ========================================================================== */

let isExecuting = false;
let isConsoleExpanded = false;

document.addEventListener("DOMContentLoaded", () => {
    initNavigation();
    initSubscriptions();
    checkTools();
    refreshDeviceStatus();
});

// -----------------------------------------------------------------------------
// Navigation & Views
// -----------------------------------------------------------------------------
function initNavigation() {
    document.querySelectorAll(".nav-link").forEach(btn => {
        btn.addEventListener("click", () => {
            const step = btn.getAttribute("data-step");
            goToStep(step);
        });
    });
}

function goToStep(stepId) {
    document.querySelectorAll(".view-panel").forEach(p => p.classList.remove("active"));
    document.querySelectorAll(".nav-link").forEach(l => l.classList.remove("active"));

    const targetPanel = document.getElementById("view-" + stepId);
    if (targetPanel) {
        targetPanel.classList.add("active");
    }

    const targetNav = document.querySelector(`.nav-link[data-step="${stepId}"]`);
    if (targetNav) {
        targetNav.classList.add("active");
    }
}

// -----------------------------------------------------------------------------
// Subscriptions & Wails IPC
// -----------------------------------------------------------------------------
function initSubscriptions() {
    if (window.runtime && window.runtime.EventsOn) {
        window.runtime.EventsOn("log", line => {
            appendLog(line);
        });

        window.runtime.EventsOn("device_state", state => {
            renderDeviceTelemetry(state);
        });
    }

    const btnRefresh = document.getElementById("btnRefreshDevice");
    if (btnRefresh) {
        btnRefresh.addEventListener("click", refreshDeviceStatus);
    }

    const btnTools = document.getElementById("btnEnsureTools");
    if (btnTools) {
        btnTools.addEventListener("click", ensureTools);
    }
}

// -----------------------------------------------------------------------------
// Device Telemetry
// -----------------------------------------------------------------------------
async function refreshDeviceStatus() {
    if (!window.go || !window.go.main || !window.go.main.App) return;

    try {
        const state = await window.go.main.App.DetectDevice();
        renderDeviceTelemetry(state);
    } catch (e) {
        console.error("DetectDevice failed:", e);
    }
}

function renderDeviceTelemetry(state) {
    const ring = document.getElementById("deviceRing");
    const statusText = document.getElementById("deviceStatusText");
    const detailText = document.getElementById("deviceDetailText");
    const tagModel = document.getElementById("tagModel");
    const tagRelease = document.getElementById("tagRelease");
    const tagRoot = document.getElementById("tagRoot");

    if (!state || state.mode === "disconnected") {
        ring.className = "connection-ring disconnected";
        statusText.textContent = "No Device Detected";
        detailText.textContent = "Connect Acer tablet via USB cable";
        tagModel.textContent = "Model: --";
        tagModel.className = "telemetry-tag";
        tagRelease.textContent = "Android: --";
        tagRelease.className = "telemetry-tag";
        tagRoot.textContent = "Root: --";
        tagRoot.className = "telemetry-tag";
        return;
    }

    if (state.mode === "adb") {
        ring.className = "connection-ring adb";
        const adb = state.adbInfo;
        const model = adb.model ? adb.model : "Acer T4-82L-G2";
        statusText.textContent = `${model} (ADB Online)`;
        detailText.textContent = `Serial: ${adb.serial || "Connected"} | Kernel GKI v4`;

        tagModel.textContent = model;
        tagModel.className = "telemetry-tag tag-active";

        tagRelease.textContent = `Android ${adb.androidRelease || "16"} (SDK ${adb.vendorSdk || "35"})`;
        tagRelease.className = "telemetry-tag tag-active";

        if (adb.rootGranted) {
            tagRoot.textContent = "Root: Magisk Active";
            tagRoot.className = "telemetry-tag tag-active";
        } else {
            tagRoot.textContent = "Root: Pending Grant";
            tagRoot.className = "telemetry-tag";
        }
        return;
    }

    if (state.mode === "fastboot") {
        ring.className = "connection-ring fastboot";
        statusText.textContent = "Fastboot Bootloader Mode";
        detailText.textContent = `Serial: ${state.fastbootSerial || "Device Ready"} | Partitions Unlocked`;

        tagModel.textContent = "Acer T4-82L-G2";
        tagModel.className = "telemetry-tag tag-active";
        tagRelease.textContent = "Fastboot Target";
        tagRelease.className = "telemetry-tag tag-active";
        tagRoot.textContent = "GKI Ready";
        tagRoot.className = "telemetry-tag";
        return;
    }
}

// -----------------------------------------------------------------------------
// Tools Check
// -----------------------------------------------------------------------------
async function checkTools() {
    if (!window.go || !window.go.main || !window.go.main.App) return;

    const pill = document.getElementById("toolsStatusPill");
    try {
        const tools = await window.go.main.App.CheckTools();
        if (tools.adb && tools.fastboot) {
            pill.textContent = "Engine Ready";
            pill.className = "badge-status badge-success";
        } else {
            pill.textContent = "Setup Required";
            pill.className = "badge-status badge-fastboot";
        }
    } catch (e) {
        pill.textContent = "Error";
        pill.className = "badge-status badge-danger";
    }
}

async function ensureTools() {
    if (!window.go || !window.go.main || !window.go.main.App) return;

    const pill = document.getElementById("toolsStatusPill");
    pill.textContent = "Downloading...";
    pill.className = "badge-status badge-neutral";

    try {
        const tools = await window.go.main.App.EnsureTools();
        if (tools.adb && tools.fastboot) {
            pill.textContent = "Engine Ready";
            pill.className = "badge-status badge-success";
            appendLog("[INFO] Platform tools successfully cached and verified.");
        }
    } catch (err) {
        pill.textContent = "Failed";
        pill.className = "badge-status badge-danger";
        appendLog("[ERROR] Failed to download platform-tools: " + err);
    }
}

// -----------------------------------------------------------------------------
// Precision Console Engine
// -----------------------------------------------------------------------------
function getTimestamp() {
    const d = new Date();
    return d.toTimeString().split(" ")[0];
}

function appendLog(rawLine) {
    const viewport = document.getElementById("logOutput");
    if (!viewport) return;

    const entry = document.createElement("div");
    entry.className = "log-entry";

    const ts = getTimestamp();
    const line = String(rawLine);

    if (line.includes("✓") || line.includes("SUCCESS") || line.includes("COMPLETED")) {
        entry.classList.add("log-entry-ok");
    } else if (line.includes("⚠️") || line.includes("Warning") || line.includes("Scanning") || line.includes("LOOK AT")) {
        entry.classList.add("log-entry-warn");
    } else if (line.includes("❌") || line.includes("Error:") || line.includes("failed") || line.includes("Timeout")) {
        entry.classList.add("log-entry-err");
    } else if (line.startsWith("==") || line.startsWith("⚡") || line.startsWith("🛡️") || line.startsWith("🚀")) {
        entry.classList.add("log-entry-highlight");
    }

    entry.textContent = `[${ts}] ${line}`;
    viewport.appendChild(entry);
    viewport.scrollTop = viewport.scrollHeight;
}

function clearLogs() {
    const viewport = document.getElementById("logOutput");
    if (viewport) {
        viewport.innerHTML = `<div class="log-entry log-entry-dim">[${getTimestamp()}] Console buffer cleared.</div>`;
    }
}

function copyLogs() {
    const viewport = document.getElementById("logOutput");
    if (viewport) {
        navigator.clipboard.writeText(viewport.innerText);
        appendLog("[OK] Console log copied to clipboard.");
    }
}

function toggleConsole() {
    const dock = document.getElementById("consoleDock");
    const btn = document.getElementById("btnToggleConsole");
    if (!dock) return;

    isConsoleExpanded = !isConsoleExpanded;
    if (isConsoleExpanded) {
        dock.classList.add("expanded");
        btn.textContent = "Collapse";
    } else {
        dock.classList.remove("expanded");
        btn.textContent = "Expand";
    }
}

// -----------------------------------------------------------------------------
// Step Handlers
// -----------------------------------------------------------------------------
function setButtonBusy(btnId, isBusy, busyText, normalText) {
    isExecuting = isBusy;
    const btn = document.getElementById(btnId);
    if (!btn) return;

    btn.disabled = isBusy;
    if (isBusy) {
        btn.dataset.original = normalText;
        btn.innerHTML = `<span>⏳ ${busyText}</span>`;
    } else {
        btn.innerHTML = btn.dataset.original || normalText;
    }
}

async function runStep1() {
    if (isExecuting) return;
    setButtonBusy("btnRunStep1", true, "Executing Flash Routine...", "<span>Flash Unlock & Root Now</span>");

    try {
        appendLog("[START] Initiating Step 1: Bootloader Unlock & Magisk Root Flash...");
        await window.go.main.App.RunStep1_Root();
        appendLog("[SUCCESS] Step 1 finished. Tablet is rebooting into Android.");
    } catch (err) {
        appendLog("[ERROR] Step 1 failed: " + err);
    } finally {
        setButtonBusy("btnRunStep1", false, "", "<span>Flash Unlock & Root Now</span>");
        refreshDeviceStatus();
    }
}

async function runStep2A() {
    if (isExecuting) return;
    setButtonBusy("btnRunStep2A", true, "Installing Magisk APK...", "<span>2A: Install Magisk App &amp; Trigger Setup</span>");

    try {
        appendLog("[START] Initiating Phase 2A: Installing Magisk APK & launching on tablet...");
        await window.go.main.App.RunStep2A_InstallMagisk();
        appendLog("[INFO] Phase 2A complete. If Magisk shows 'Requires Additional Setup', tap OK and let the tablet reboot.");
    } catch (err) {
        appendLog("[ERROR] Phase 2A failed: " + err);
    } finally {
        setButtonBusy("btnRunStep2A", false, "", "<span>2A: Install Magisk App &amp; Trigger Setup</span>");
        refreshDeviceStatus();
    }
}

async function runStep2B() {
    if (isExecuting) return;
    setButtonBusy("btnRunStep2B", true, "Authorizing & Neutralizing...", "<span>2B: Authorize Root &amp; Neutralize MDM</span>");

    try {
        appendLog("[START] Initiating Phase 2B: Verifying Shell root grant & applying MDM countermeasures...");
        await window.go.main.App.RunStep2B_NeutralizeMDM();
        appendLog("[SUCCESS] Phase 2B complete! All MDM components permanently neutralized.");
    } catch (err) {
        appendLog("[ERROR] Phase 2B failed: " + err);
    } finally {
        setButtonBusy("btnRunStep2B", false, "", "<span>2B: Authorize Root &amp; Neutralize MDM</span>");
        refreshDeviceStatus();
    }
}

async function runStep2() {
    return runStep2B();
}

async function runStep3() {
    if (isExecuting) return;
    setButtonBusy("btnRunStep3", true, "Applying Performance Tweaks...", "<span>Apply Performance Optimizations</span>");

    try {
        appendLog("[START] Initiating Step 3: Performance Boost & Telemetry Debloat...");
        await window.go.main.App.RunStep3_Performance();
        appendLog("[SUCCESS] Step 3 complete. System responsiveness boosted.");
    } catch (err) {
        appendLog("[ERROR] Step 3 failed: " + err);
    } finally {
        setButtonBusy("btnRunStep3", false, "", "<span>Apply Performance Optimizations</span>");
        refreshDeviceStatus();
    }
}

async function browseBackupDir() {
    try {
        const path = await window.go.main.App.SelectDirectory("Select Stock Backup Destination Directory");
        if (path) {
            document.getElementById("backupDirPath").value = path;
        }
    } catch (e) {
        console.error("Browse directory failed:", e);
    }
}

async function runBackup() {
    if (isExecuting) return;
    const path = document.getElementById("backupDirPath").value;
    if (!path) {
        alert("Please select a target backup destination directory first.");
        return;
    }

    setButtonBusy("btnRunBackup", true, "Dumping Raw Partitions...", "<span>Start Partition Backup</span>");

    try {
        appendLog(`[START] Dumping dynamic partitions to: ${path}`);
        await window.go.main.App.RunStep4_Backup(path);
        appendLog("[SUCCESS] All partitions dumped and verified successfully.");
    } catch (err) {
        appendLog("[ERROR] Backup routine failed: " + err);
    } finally {
        setButtonBusy("btnRunBackup", false, "", "<span>Start Partition Backup</span>");
    }
}

async function browseRestoreDir() {
    try {
        const path = await window.go.main.App.SelectDirectory("Select Directory Containing Stock Images");
        if (path) {
            document.getElementById("restoreDirPath").value = path;
        }
    } catch (e) {
        console.error("Browse directory failed:", e);
    }
}

async function runRestore() {
    if (isExecuting) return;
    const path = document.getElementById("restoreDirPath").value;
    if (!path) {
        alert("Please select the source backup directory containing stock partition images.");
        return;
    }

    const confirmed = confirm("WARNING: This will flash factory images back into the dynamic partitions. Proceed?");
    if (!confirmed) return;

    setButtonBusy("btnRunRestore", true, "Restoring Images...", "<span>Restore Stock Images</span>");

    try {
        appendLog(`[START] Restoring stock partition images from: ${path}`);
        await window.go.main.App.RunStockRestore(path);
        appendLog("[SUCCESS] Stock restore routine finished successfully.");
    } catch (err) {
        appendLog("[ERROR] Restore routine failed: " + err);
    } finally {
        setButtonBusy("btnRunRestore", false, "", "<span>Restore Stock Images</span>");
    }
}
