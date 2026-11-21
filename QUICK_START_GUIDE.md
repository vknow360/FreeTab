# FreeTab v1.2 — Acer One 8 (T4-82L Gen 2) Toolkit
### All-in-One Fastboot Root, Zero-Touch MDM Neutralizer & Performance Suite

FreeTab is a standalone, all-in-one graphical desktop application engineered to unlock, root, permanently defeat Google Zero-Touch / Tectoro MDM locks, and optimize the Acer One 8 (T4-82L Gen 2) tablet. All actions are handled directly within the FreeTab graphical interface.

---

## Mandatory Requirement: Keep Wi-Fi Off
> **Do not connect to Wi-Fi until Step 2 (MDM Defense) is complete.**  
> If the tablet connects to the internet before running the MDM defense, Google Zero-Touch will immediately download the enterprise profile and lock the tablet.
> - **During setup wizard:** Choose **"Set up offline"** or **Skip**.
> - **Keep Wi-Fi OFF** until the anti-MDM step confirms completion in FreeTab.

---

## Launching FreeTab

Double-click **`FreeTab.exe`** on Windows. No command line, terminal scripts, or external dependencies are required. All flashing, device telemetry, rooting, and defense steps are fully automated inside the user interface.

---

## Guided Workflow for New Tablets

### Step 1: Unlock Bootloader & Flash Root
1. Power off the tablet completely.
2. Hold **Power + Volume Up** until the boot selection menu appears:
   ```text
   [Recovery Mode]
   [Fastboot Mode]  <<==
   [Normal Boot]
   ```
3. Press **Volume Up** to point the arrow at `[Fastboot Mode]`, then press **Volume Down** to select it.  
   Verify the screen displays `=> FASTBOOT mode...`.
4. Connect the USB-C cable to your computer.  
   *(On desktop PCs or AMD Ryzen systems, use a black USB 2.0 port or a USB 2.0 hub to ensure stable packet transfer).*
5. In FreeTab, navigate to **Step 1 (Unlock & Root)** and click **Flash Unlock & Root Now**.
6. **Watch the tablet screen:** When prompted, press **Volume Up** once to confirm bootloader unlock.  
   *(FreeTab automatically flashes root to slots A & B and reboots the tablet).*

---

### Step 2: Offline Setup & Enable USB Debugging
1. When the tablet reboots, complete the Android setup wizard:
   * **Do not connect to Wi-Fi.** Choose **"Set up offline"** or **Skip**.
2. Enable Developer Options:
   * Open **Settings > About tablet** -> Tap **Build number** 7 times.
3. Enable USB Debugging:
   * Open **Settings > System > Developer options** -> Toggle **USB debugging** to **ON**.
4. Connect to PC and tap **"Always allow from this computer"** on the tablet screen.

---

### Step 3: Defeat Zero-Touch MDM Lock
1. In FreeTab, navigate to **Step 2 (Defeat MDM Lock)** and click **2A: Install Magisk App**.
2. Open the **Magisk** app on the tablet screen:
   * If prompted *"Requires Additional Setup"*, tap **OK** (the tablet reboots).
   * After reboot, open **Magisk** again -> tap **Superuser** (shield icon in the bottom navigation bar) -> toggle **Shell** to **ON**.
3. In FreeTab, click **2B: Confirm Shell Root Permission** to verify access.
4. Click **2C: Execute Anti-MDM Defense** to permanently neutralize Zero-Touch and apply kernel firewall protection.

---

### Step 4: Apply Performance Boost (Optional)
In FreeTab, navigate to **Step 3 (Performance Boost)** and click **Apply Performance Boost**:
* Sets animation scale to **0.5x** for immediate UI responsiveness.
* Disables background telemetry and unneeded system packages.
* Dismisses remaining setup wizard notifications.

---

### Step 5: Setup Complete
* Turn Wi-Fi ON.
* Add your personal Google account.
* Use your tablet freely.

---

## Windows USB Driver & Fastboot Troubleshooting

If your tablet is not detected in Fastboot mode or ADB mode, follow these verified driver configurations:

### 1. Fastboot Mode Recognition (Resolving `< waiting for any device >`)
When the tablet enters Fastboot mode, its hardware ID changes to `USB\VID_18D1&PID_4EE0` (Google Android Bootloader Interface).

* **Check Device Manager (`devmgmt.msc`):**
  When connected in Fastboot mode, look for an entry named **`Android`** under **Other devices** with a warning icon.
* **Install the Correct Driver:**
  1. Right-click **`Android`** under Other devices -> select **Update driver**.
  2. Select **Browse my computer for drivers** -> **Let me pick from a list of available drivers on my computer**.
  3. Select **Android Device** (or Show All Devices) -> click Next.
  4. Under Model, select **Android Bootloader Interface** (do **not** select "Android ADB Interface").
  5. Click Next and confirm the installation warning.
* **Do Not Use "Add Legacy Hardware":**
  Using *Action -> Add legacy hardware* creates a virtual phantom device disconnected from the physical USB bus. Windows will fail to start it with **Code 10** (`This device cannot start`). Always update the real hardware device under Other devices.
* **USB 3.0 / AMD Ryzen Motherboard Issue:**
  MediaTek Helio G85 bootloaders frequently fail USB packet handshakes on blue or red USB 3.0/3.1 ports. Connect the tablet to a black **USB 2.0 port** or use an external USB 2.0 hub.

### 2. Android OS Mode (ADB USB Debugging)
* In normal booted Android OS, the tablet presents a MediaTek hardware ID (`USB\VID_0E8D&PID_201C`).
* The standard Google USB Driver inf file does not include `VID_0E8D` by default and will report no compatible drivers found.
* **Recommended Driver:** Install the **[Universal ADB Driver](https://adb.clockworkmod.com/)** (by Koushik Dutta) or the MediaTek USB Driver package. These packages support both MediaTek ADB (`VID_0E8D`) and Google Fastboot (`VID_18D1`).

---

## Credits & Acknowledgments
Special credit to **[codeturbulent/Acerone8_](https://github.com/codeturbulent/Acerone8_)** for the original foundation scripts, partition analysis, and fastboot rooting workflow for the Acer One 8 (T4-82L Gen 2).

---

## Hardware Recovery Cheat Sheet

* **Hard Power Off (Stuck in Bootloop):**  
  Unplug USB, then hold **Power + Volume Up + Volume Down** simultaneously for 15 seconds.
* **Boot Selection Menu:**  
  Hold **Power + Volume Up** while powered off.
* **Factory Reset (Format Data):**  
  Hold **Power + Volume Down** while powered off -> select **`Clear Storage`**.
* **Restore Stock Backup:**  
  In fastbootd (`fastboot reboot fastboot`):
  ```powershell
  fastboot flash system tablet_backup\stock_system.img
  fastboot flash product tablet_backup\stock_product.img
  fastboot flash system_ext tablet_backup\stock_system_ext.img
  fastboot reboot
  ```
