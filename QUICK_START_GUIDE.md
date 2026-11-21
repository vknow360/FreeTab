# 🚀 FreeTab v1.2 — Acer One 8 (T4-82L Gen 2) Toolkit
### All-in-One Fastboot Root, Zero-Touch MDM Neutralizer & Performance Suite

This enhanced toolkit provides an interactive launcher and individual 1-click scripts to unlock, root, permanently defeat the Tectoro / Google Zero-Touch MDM lock, and optimize any Acer One 8 tablet without technical hassle.

---

## ⚠️ THE GOLDEN RULE
> ### 🛑 DO NOT CONNECT TO WI-FI UNTIL STEP 2 (MDM DEFENSE) IS COMPLETE!
> If the tablet connects to the internet before running the MDM defense, Google Zero-Touch will immediately download the enterprise profile and lock the tablet with a wipe countdown.
> - **During setup wizard:** Choose **"Set up offline"** or **Skip**.
> - **Keep Wi-Fi OFF** until the anti-MDM step confirms `🎉 SUCCESS`.

---

## 🎮 How to Use (Interactive Launcher or 1-Click Scripts)

You can simply double-click **`Run_Toolkit.bat`** to get the full interactive menu, or use the numbered standalone scripts:

```text
Run_Toolkit.bat
  [1] ⚡ 1-Click Fastboot Unlock & Magisk Root
  [2] 🛡️  Neutralize Zero-Touch / Tectoro MDM Lock
  [3] 🚀 Apply Extreme Performance Boost (0.5x Scale + Debloat)
  [4] 🔄 Full Guided Express Flow (Steps 1, 2 & 3 in sequence)
  [5] 🛟 Check Device Status & Diagnostics
  [0] ❌ Exit Toolkit
```

---

## 📋 The 10-Minute Workflow for Any New Tablet:

### Step 1: Unlock Bootloader & Flash Root (2 Minutes)
1. Power OFF the tablet completely.
2. Hold **Power + Volume UP** until the small boot menu appears.
3. Select **`[Fastboot Mode]`** with Volume Down.
4. Plug in the USB cable to your PC.
5. Double-click **`Run_Toolkit.bat`** and select **`[1]`** (or double-click **`1_Fastboot_Root.bat`**).
6. ⚠️ **LOOK AT TABLET SCREEN:** Press **Volume UP** once to confirm bootloader unlock.  
   *(The script flashes root and reboots the tablet automatically).*

---

### Step 2: Offline Setup & Enable USB Debugging (1 Minute)
1. When the tablet reboots, complete the Android setup wizard:
   * **DO NOT CONNECT TO WI-FI!** Select **"Set up offline"** / **Skip**.
2. Enable Developer Options:
   * Open **Settings > About tablet** -> Tap **Build number** 7 times.
3. Enable USB Debugging:
   * Open **Settings > System > Developer options** -> Toggle **USB debugging** to **ON**.
4. Plug into PC and tap **"Always allow from this computer"** on the tablet screen.

---

### Step 3: Defeat Zero-Touch MDM Lock (2 Minutes)
1. In FreeTab (or Toolkit menu), run Stage **2A** to install Magisk app.
2. Open the **Magisk** app on your tablet screen:
   * If it prompts *"Requires Additional Setup"*, tap **OK** (it reboots).
   * After reboot, open **Magisk** again -> tap **Superuser** (shield icon in bottom navigation bar) -> toggle **Shell** to **ON**.
3. In FreeTab, click **2B: Confirm Shell Root Permission** to verify.
4. Click **2C: Execute Anti-MDM Defense** to permanently neutralize Zero-Touch and apply kernel firewall protection.

---

### Step 4: Apply 1-Click Performance Boost (30 Seconds)
In the Toolkit menu, select **`[3]`** (or double-click **`3_Performance_Boost.bat`**).
* Cuts animation scale to **0.5x** (instant 2x UI speed).
* Disables useless background telemetry & bloatware (~800 MB RAM freed).
* Finalizes and dismisses lingering setup wizard notifications.

---

### Step 5: You Are Done!
* **Turn Wi-Fi ON.**
* Add your Google account.
* Use your tablet as a fast, clean, personal device!

---

## 🔌 Windows USB Driver Diagnostics
FreeTab includes automatic, zero-permission Windows USB driver diagnostics in the left sidebar and main dashboard:
- **ADB Interface (`WinUSB` / `androidusb`):** Required during Android runtime for MDM defense and performance optimization.
- **Fastboot Interface (`WinUSB`):** Crucial for Step 1 bootloader unlocking and kernel flashing.
- **MediaTek Preloader:** Fallback port for MTK factory recovery.
- If FreeTab flags missing drivers or a **Code 28** error in Device Manager, click **"Download Google USB Driver"** or **"Open Device Manager"** right from the FreeTab UI.

---

## 🛟 Emergency Recovery Cheat Sheet

* **Hard Power Off (If ever stuck in a reboot loop):**  
  Unplug USB, then hold **Power + Volume Up + Volume Down** simultaneously for 15 seconds.
* **Boot Selection Menu:**  
  Hold **Power + Volume Up** while powered off.
* **Factory Reset (Format Data):**  
  Hold **Power + Volume Down** while powered off -> select **`Clear Storage`**.
* **Restore 100% Stock Backup:**  
  In fastbootd (`fastboot reboot fastboot`):
  ```powershell
  fastboot flash system tablet_backup\stock_system.img
  fastboot flash product tablet_backup\stock_product.img
  fastboot flash system_ext tablet_backup\stock_system_ext.img
  fastboot reboot
  ```
