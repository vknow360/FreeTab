/* ==========================================================================
   FreeTab Pro — High-Precision Logic & Telemetry Engine
   ========================================================================== */

let isExecuting = false;
let isConsoleExpanded = false;

document.addEventListener("DOMContentLoaded", () => {
    initNavigation();
    initSubscriptions();
    checkTools();
    checkDrivers();
    refreshDeviceStatus();
    checkForUpdates(false);
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
// Driver Diagnostics Check
// -----------------------------------------------------------------------------
let cachedDriverReport = null;

async function checkDrivers() {
    if (!window.go || !window.go.main || !window.go.main.App) return;

    const pill = document.getElementById("driversStatusPill");
    if (pill) {
        pill.textContent = "Scanning...";
        pill.className = "badge-status badge-neutral";
    }

    try {
        const report = await window.go.main.App.CheckDrivers();
        cachedDriverReport = report;
        renderDriverUI(report);
    } catch (err) {
        if (pill) {
            pill.textContent = "Error";
            pill.className = "badge-status badge-danger";
        }
        appendLog("[ERROR] Driver diagnostic failed: " + err);
    }
}

function renderDriverUI(report) {
    const pill = document.getElementById("driversStatusPill");
    if (pill) {
        if (report.ready) {
            pill.textContent = "Drivers OK";
            pill.className = "badge-status badge-success";
        } else if (report.problemFound) {
            pill.textContent = "Driver Error";
            pill.className = "badge-status badge-danger";
        } else {
            pill.textContent = "Missing";
            pill.className = "badge-status badge-fastboot";
        }
    }

    // Update Dashboard Card
    const adbBadge = document.getElementById("driverBadgeAdb");
    const fbBadge = document.getElementById("driverBadgeFastboot");
    const mtkBadge = document.getElementById("driverBadgeMtk");
    const driverSummary = document.getElementById("driverSummaryText");
    const driverAlert = document.getElementById("driverAlertBox");
    const driverActionText = document.getElementById("driverActionText");

    if (adbBadge) {
        adbBadge.textContent = report.adbDriver.status;
        adbBadge.className = report.adbDriver.status === "OK" || report.adbDriver.installed ? "badge-status badge-success" : "badge-status badge-danger";
    }
    if (fbBadge) {
        fbBadge.textContent = report.fastbootDriver.status;
        fbBadge.className = report.fastbootDriver.status === "OK" || report.fastbootDriver.installed ? "badge-status badge-success" : "badge-status badge-danger";
    }
    if (mtkBadge) {
        mtkBadge.textContent = report.mtkDriver.status;
        mtkBadge.className = report.mtkDriver.status === "OK" || report.mtkDriver.installed ? "badge-status badge-success" : "badge-status badge-neutral";
    }
    if (driverSummary) {
        driverSummary.textContent = report.summary;
    }
    if (driverAlert) {
        if (!report.ready) {
            driverAlert.style.display = "flex";
            if (driverActionText) driverActionText.textContent = report.actionNeeded;
        } else {
            driverAlert.style.display = "none";
        }
    }

    // Step 1 Fastboot Driver Warning Strip
    const step1FbWarning = document.getElementById("step1FastbootDriverWarning");
    if (step1FbWarning) {
        if (!report.fastbootDriver.installed && report.fastbootDriver.status !== "OK") {
            step1FbWarning.style.display = "flex";
        } else {
            step1FbWarning.style.display = "none";
        }
    }
}

async function openDeviceManager() {
    try {
        await window.go.main.App.OpenDeviceManager();
    } catch (e) {
        appendLog("[ERROR] Could not launch Device Manager: " + e);
    }
}

async function openDriverDownload() {
    try {
        await window.go.main.App.OpenDriverDownload();
    } catch (e) {
        appendLog("[ERROR] Could not open driver download: " + e);
    }
}

// -----------------------------------------------------------------------------
// GitHub Auto-Update Engine
// -----------------------------------------------------------------------------
let cachedUpdateInfo = null;

async function checkForUpdates(manual = false) {
    if (!window.go || !window.go.main || !window.go.main.App) return;

    const pill = document.getElementById("updateStatusPill");
    if (pill) {
        pill.textContent = "Checking...";
        pill.className = "badge-status badge-neutral";
    }

    try {
        const info = await window.go.main.App.CheckForUpdates();
        cachedUpdateInfo = info;

        if (info && info.updateAvailable) {
            if (pill) {
                pill.textContent = "v" + info.latestVersion + " Available";
                pill.className = "badge-status badge-success";
            }
            showUpdateBanner(info);
            if (manual) {
                appendLog(`[INFO] New update available: v${info.currentVersion} → v${info.latestVersion} (${info.releaseName || "Latest Release"})`);
            }
        } else {
            if (pill) {
                pill.textContent = "v" + (info && info.currentVersion ? info.currentVersion : "1.2");
                pill.className = "badge-status badge-neutral";
            }
            hideUpdateBanner();
            if (manual) {
                appendLog(`[INFO] FreeTab is up to date (v${info && info.currentVersion ? info.currentVersion : "1.2"}).`);
            }
        }
    } catch (err) {
        if (pill) {
            pill.textContent = "v1.2";
            pill.className = "badge-status badge-neutral";
        }
        if (manual) {
            appendLog("[WARN] Could not check for updates: " + err);
        }
    }
}

function showUpdateBanner(info) {
    const banner = document.getElementById("updateBanner");
    const desc = document.getElementById("updateBannerDesc");
    if (!banner) return;

    if (desc) {
        desc.textContent = `Version v${info.latestVersion} is available to install.`;
    }
    banner.style.display = "flex";
}

function hideUpdateBanner() {
    const banner = document.getElementById("updateBanner");
    if (banner) {
        banner.style.display = "none";
    }
}

function dismissUpdateBanner() {
    hideUpdateBanner();
}

async function applyUpdate() {
    if (!cachedUpdateInfo || !cachedUpdateInfo.downloadUrl) {
        appendLog("[ERROR] No valid update download URL found.");
        return;
    }

    const btn = document.getElementById("btnApplyUpdate");
    if (btn) {
        btn.disabled = true;
        btn.textContent = "Downloading & Installing...";
    }

    appendLog(`[INFO] Starting auto-upgrade to v${cachedUpdateInfo.latestVersion}...`);
    try {
        await window.go.main.App.InstallUpdate(cachedUpdateInfo.downloadUrl, cachedUpdateInfo.assetName);
        if (btn) {
            btn.disabled = false;
            btn.textContent = "Restart FreeTab Now";
            btn.onclick = restartApp;
        }
        appendLog("[SUCCESS] Upgrade completed! Click 'Restart FreeTab Now' to launch the new version.");
    } catch (err) {
        appendLog("[ERROR] Auto-update failed: " + err);
        if (btn) {
            btn.disabled = false;
            btn.textContent = "Retry Update";
        }
    }
}

async function restartApp() {
    try {
        await window.go.main.App.RestartApp();
    } catch (err) {
        appendLog("[ERROR] Failed to restart application: " + err);
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

async function rebootToFastboot() {
    if (isExecuting) return;
    setButtonBusy("btnRebootFastboot", true, "Rebooting...", "<span>⚡ 1-Click Reboot to Fastboot (via ADB)</span>");

    try {
        appendLog("[START] Sending 'adb reboot bootloader' to tablet...");
        await window.go.main.App.RebootToFastboot();
        appendLog("[SUCCESS] Tablet is restarting into Fastboot mode. Waiting a few seconds...");
    } catch (err) {
        appendLog("[ERROR] Failed to reboot to Fastboot: " + err);
    } finally {
        setButtonBusy("btnRebootFastboot", false, "", "<span>⚡ 1-Click Reboot to Fastboot (via ADB)</span>");
        refreshDeviceStatus();
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
        appendLog("[START] Initiating Stage 2A: Installing Magisk APK & launching on tablet...");
        await window.go.main.App.RunStep2A_InstallMagisk();
        appendLog("[INFO] Stage 2A complete. If Magisk shows 'Requires Additional Setup', tap OK and let the tablet reboot.");
    } catch (err) {
        appendLog("[ERROR] Stage 2A failed: " + err);
    } finally {
        setButtonBusy("btnRunStep2A", false, "", "<span>2A: Install Magisk App &amp; Trigger Setup</span>");
        refreshDeviceStatus();
    }
}

async function runStep2B() {
    if (isExecuting) return;
    setButtonBusy("btnRunStep2B", true, "Verifying Root Grant...", "<span>2B: Confirm Shell Root Permission</span>");

    try {
        appendLog("[START] Initiating Stage 2B: Verifying Shell root permission in Magisk...");
        await window.go.main.App.RunStep2B_AuthorizeRoot();
        appendLog("[SUCCESS] Stage 2B complete! Shell root access is confirmed. You can now run Stage 2C.");
    } catch (err) {
        appendLog("[ERROR] Stage 2B failed: " + err);
    } finally {
        setButtonBusy("btnRunStep2B", false, "", "<span>2B: Confirm Shell Root Permission</span>");
        refreshDeviceStatus();
    }
}

async function runStep2C() {
    if (isExecuting) return;
    setButtonBusy("btnRunStep2C", true, "Applying MDM Countermeasures...", "<span>2C: Execute Anti-MDM Defense</span>");

    try {
        appendLog("[START] Initiating Stage 2C: Executing full Zero-Touch & MDM defense routine...");
        await window.go.main.App.RunStep2C_ExecuteMDM();
        appendLog("[SUCCESS] Stage 2C complete! All MDM components permanently neutralized.");
    } catch (err) {
        appendLog("[ERROR] Stage 2C failed: " + err);
    } finally {
        setButtonBusy("btnRunStep2C", false, "", "<span>2C: Execute Anti-MDM Defense</span>");
        refreshDeviceStatus();
    }
}

async function runStep2() {
    return runStep2C();
}

async function runStep3() {
    if (isExecuting) return;

    const confirmed = confirm("WARNING: Performance boost modifies system animation scales and disables 7 background services (including Print Spooler, Digital Wellbeing, and MediaTek log tools).\n\nThis step is OPTIONAL. Would you like to proceed?");
    if (!confirmed) return;

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
