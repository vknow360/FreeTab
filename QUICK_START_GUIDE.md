# 🚀 FreeTab v1.0 — Acer One 8 (T4-82L Gen 2) Toolkit
### Standalone Windows GUI Utility for Bootloader Unlock, Magisk Root & MDM Defense

FreeTab is an all-in-one GUI application specifically engineered to root, permanently neutralize enterprise MDM locks (Google Zero-Touch / Tectoro), and optimize the **Acer One 8 T4-82L Gen 2** (MediaTek Helio G85, Android 16 / SDK 35).

---

## ⚠️ CRITICAL RULES BEFORE STARTING

> ### 🛑 1. DO NOT CONNECT TO WI-FI UNTIL STEP 2 IS COMPLETE
> If the tablet connects to the internet before running the Anti-MDM Defense, Google Zero-Touch will immediately download the enterprise profile and lock the tablet with a wipe countdown.
> - **During Android setup:** Choose **"Set up offline"** or tap **Skip**.
> - Keep Wi-Fi **OFF** until FreeTab confirms `🎉 MDM DEFENSE DEPLOYED SUCCESSFULLY`.

> ### 🛑 2. OEM UNLOCKING MUST BE TOGGLED ON
> If the tablet is already locked by MDM:
> 1. Power OFF the tablet.
> 2. Hold **Power + Volume Down** until Factory Mode appears.
> 3. Use Vol Down to highlight **`Clear Storage`** and press Power to wipe.
> 4. In Android Setup, set up **offline** (no Wi-Fi).
> 5. Go to **Settings > About tablet** ➔ tap **Build number** 7 times.
> 6. Go to **Settings > System > Developer options**:
>    - Toggle **OEM unlocking** to **ON**.
>    - Toggle **USB debugging** to **ON**.

---

## 📋 The 3-Step FreeTab Workflow

### Step 1: Unlock Bootloader & Flash Root
1. Connect the tablet to your PC via USB.
2. Launch **`FreeTab.exe`**.
3. In Step 1, click **"1-Click Reboot to Fastboot"** (or boot manually by holding `Power + Volume Up` while powered off).
4. Click **"Unlock Bootloader & Root"**.
5. ⚠️ **LOOK AT TABLET SCREEN:** Press **Volume UP** once to confirm bootloader unlock.  
   *(FreeTab will flash the patched init_boot image and reboot the tablet automatically).*

---

### Step 2: Deploy Anti-MDM Defense (3 Sub-Steps)
Complete the offline Android setup wizard (**NO WI-FI!**) and ensure USB debugging is enabled:

1. **Stage 2A: Install Magisk Management App**
   * Click **"Install Magisk App"** in FreeTab.
   * Open the Magisk app on your tablet screen.
   * If it prompts *"Requires Additional Setup"*, tap **OK** (the tablet will reboot automatically).
2. **Stage 2B: Authorize Shell Root Permission**
   * Click **"Authorize Shell Root"** in FreeTab.
   * Look at the tablet screen and tap **GRANT** on the superuser request.
   * FreeTab will verify root status (`uid=0 (root)` confirmed).
3. **Stage 2C: Execute Anti-MDM Defense**
   * Click **"Execute Anti-MDM Defense"**.
   * FreeTab permanently disables Zero-Touch components, locks configuration attributes (`chattr +i`), activates the DLC firewall, and installs the persistent `/data/adb/service.d/block_mdm.sh` boot script.

---

### Step 3: Performance Boost (Optional)
Click **"Apply Performance Boost"** to:
* Set animation scale to 0.5x (instant UI responsiveness).
* Disable background telemetry and bloatware packages.
* Dismiss lingering setup wizard notifications.

---

### Step 4: Complete!
* Turn Wi-Fi **ON**.
* Log in with your Google account.
* Enjoy your fast, unlocked, personal tablet!

---

## 🛟 Hardware Keys Reference

* **Hard Power Off (Stuck in loop):** Unplug USB ➔ hold **Power + Volume Up + Volume Down** for 15 seconds.
* **Boot Menu (Fastboot / Recovery):** Power off ➔ hold **Power + Volume Up**.
* **Factory Mode (Clear Storage wipe):** Power off ➔ hold **Power + Volume Down**.
