package actions

import (
	"context"
	"fmt"
	"freetab/internal/assets"
	"strings"
	"time"
)

// InstallMagiskApp performs Phase 2A: Installs Magisk.apk and launches it to trigger the Additional Setup reboot.
func (a *Actions) InstallMagiskApp(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("📦 PHASE 2A: Install Magisk Management App")
	a.log("==========================================================================")

	// 1. Wait for ADB device
	if err := a.ADB.WaitForDevice(ctx, 45*time.Second); err != nil {
		return err
	}

	// 2. Extract and install Magisk APK
	a.log("[1/2] Extracting and installing Magisk APK on tablet...")
	apkPath, cleanup, err := assets.ExtractMagiskApk()
	if err != nil {
		return fmt.Errorf("failed to extract Magisk.apk: %w", err)
	}
	defer cleanup()

	if err := a.ADB.InstallAPK(ctx, apkPath); err != nil {
		a.log("  ⚠️ Note during APK install: " + err.Error())
	} else {
		a.log("  ✓ Magisk APK installed successfully.")
	}

	// 3. Launch Magisk App
	a.log("\n[2/2] Launching Magisk on tablet...")
	_, _ = a.ADB.Shell(ctx, "monkey -p com.topjohnwu.magisk -c android.intent.category.LAUNCHER 1")

	a.log("\n==========================================================================")
	a.log("👉 LOOK AT YOUR TABLET SCREEN NOW:")
	a.log(" 1. If Magisk prompts 'Requires Additional Setup', tap OK.")
	a.log(" 2. The tablet will automatically reboot in ~10 seconds.")
	a.log(" 3. Once it finishes rebooting and you unlock the screen:")
	a.log("    👉 Click [2B: Authorize Root & Neutralize MDM] to complete the process!")
	a.log("==========================================================================")

	return nil
}

// NeutralizeMDM performs Phase 2B: Verifies root, requests shell Superuser permission, and neutralizes MDM.
func (a *Actions) NeutralizeMDM(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("🛡️ PHASE 2B: Authorize Shell Root & Neutralize Zero-Touch MDM")
	a.log("==========================================================================")

	// 1. Wait for ADB device after reboot
	a.log("Waiting for tablet ADB connection...")
	if err := a.ADB.WaitForDevice(ctx, 60*time.Second); err != nil {
		return err
	}

	// 2. Verify Root / Prompt Superuser
	a.log("\n[1/4] Verifying root privileges & requesting Superuser access...")
	idOut, _ := a.ADB.Shell(ctx, "su -c id")

	if !strings.Contains(idOut, "uid=0") {
		a.log("  [..] Root prompt not yet approved. Triggering Superuser request...")
		// Launch Magisk to ensure it's running
		_, _ = a.ADB.Shell(ctx, "monkey -p com.topjohnwu.magisk -c android.intent.category.LAUNCHER 1")

		a.log("\n  👉 LOOK AT YOUR TABLET SCREEN NOW:")
		a.log("  👉 A popup 'Superuser Request' from Shell will appear.")
		a.log("  👉 Tap 'GRANT' or 'ALLOW' to grant permanent root access!")
		a.log("  [..] Waiting for authorization...\n")

		// Poll for grant for up to 45 seconds
		granted := false
		for i := 0; i < 45; i++ {
			time.Sleep(1 * time.Second)
			// Trigger su check again
			check, _ := a.ADB.Shell(ctx, "su -c id")
			if strings.Contains(check, "uid=0") {
				granted = true
				break
			}
		}

		if !granted {
			return fmt.Errorf("root access was not granted. Please open Magisk on your tablet, grant Superuser to Shell (or in Magisk Settings set Superuser Access to 'Apps and ADB'), then retry")
		}
	}
	a.log("  ✓ Root confirmed: uid=0 (Superuser active).")

	// 3. Auto-grant ADB root in Magisk SQLite DB for future commands
	a.log("  [..] Configuring automatic ADB root authorization...")
	_, _ = a.ADB.Shell(ctx, `su -c 'magisk --sqlite "REPLACE INTO policies (uid,policy,until,logging,notification) VALUES (2000,2,0,1,1);"'`)

	// 4. Disable Zero-Touch activities & lock configuration
	a.log("\n[2/4] Disabling Zero-Touch activities & locking configuration...")
	cmdDisable := `
pm disable com.google.android.gms/.enterprise.zerotouch.ui.FactoryResetActivity >/dev/null 2>&1
pm disable com.google.android.gms/.enterprise.zerotouch.businesslogic.ZeroTouchGmsTaskService >/dev/null 2>&1
if [ -f /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml ]; then
    chattr -i /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml 2>/dev/null
    echo -n "" > /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml
    chmod 444 /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml
    chattr +i /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml 2>/dev/null
fi
pm disable com.android.managedprovisioning >/dev/null 2>&1
pm disable-user --user 0 com.android.managedprovisioning >/dev/null 2>&1
pm disable com.chipc.update >/dev/null 2>&1
pm disable-user --user 0 com.chipc.update >/dev/null 2>&1
am force-stop com.google.android.gms >/dev/null 2>&1
`
	if err := a.ADB.RootShell(ctx, cmdDisable); err != nil {
		a.log("  ⚠️ Warning executing Zero-Touch neutralization script: " + err.Error())
	} else {
		a.log("  ✓ Zero-Touch activities disabled and configuration locked.")
	}

	// 5. Firewall DeviceLockController at kernel level
	a.log("\n[3/4] Firewalling DeviceLockController at kernel level...")
	cmdFirewall := `
DLC_UID=$(pm list packages -U | grep com.android.devicelockcontroller | sed -n "s/.*uid://p")
if [ -n "$DLC_UID" ]; then
    iptables -C OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP 2>/dev/null || iptables -I OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP
    ip6tables -C OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP 2>/dev/null || ip6tables -I OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP
fi
`
	if err := a.ADB.RootShell(ctx, cmdFirewall); err != nil {
		a.log("  ⚠️ Warning configuring firewall: " + err.Error())
	} else {
		a.log("  ✓ Outgoing network traffic for DeviceLockController blocked.")
	}

	// 6. Install persistent Magisk boot service
	a.log("\n[4/4] Installing persistent Magisk boot service...")
	cmdService := `
mkdir -p /data/adb/service.d
cat << "EOF" > /data/adb/service.d/block_mdm.sh
#!/system/bin/sh
while [ "$(getprop sys.boot_completed)" != "1" ]; do
    sleep 2
done
pm disable com.google.android.gms/.enterprise.zerotouch.ui.FactoryResetActivity >/dev/null 2>&1
pm disable com.google.android.gms/.enterprise.zerotouch.businesslogic.ZeroTouchGmsTaskService >/dev/null 2>&1
if [ -f /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml ]; then
    chattr -i /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml 2>/dev/null
    echo -n "" > /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml
    chmod 444 /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml
    chattr +i /data/data/com.google.android.gms/shared_prefs/zero_touch_prefs.xml 2>/dev/null
fi
pm disable com.android.managedprovisioning >/dev/null 2>&1
pm disable-user --user 0 com.android.managedprovisioning >/dev/null 2>&1
pm disable com.chipc.update >/dev/null 2>&1
pm disable-user --user 0 com.chipc.update >/dev/null 2>&1
DLC_UID=$(pm list packages -U | grep com.android.devicelockcontroller | sed -n "s/.*uid://p")
if [ -n "$DLC_UID" ]; then
    iptables -C OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP 2>/dev/null || iptables -I OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP
    ip6tables -C OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP 2>/dev/null || ip6tables -I OUTPUT -m owner --uid-owner "$DLC_UID" -j DROP
fi
rm -f /data/system/device_owner_2.xml /data/system/device_policies.xml
EOF
chmod 755 /data/adb/service.d/block_mdm.sh
`
	if err := a.ADB.RootShell(ctx, cmdService); err != nil {
		a.log("  ⚠️ Warning installing boot service: " + err.Error())
	} else {
		a.log("  ✓ Persistent boot service installed in /data/adb/service.d/block_mdm.sh")
	}

	a.log("\n==========================================================================")
	a.log("🎉 SUCCESS: All institutional MDM / Zero-Touch locks eliminated!")
	a.log("You can now safely connect your tablet to Wi-Fi and the Internet.")
	a.log("==========================================================================")

	return nil
}

// RunMDMBypass executes the combined routine (installs Magisk if missing, then neutralizes MDM).
func (a *Actions) RunMDMBypass(ctx context.Context) error {
	// Check if root already active
	idOut, _ := a.ADB.Shell(ctx, "su -c id")
	if !strings.Contains(idOut, "uid=0") {
		// Magisk app might not even be installed or setup yet
		_ = a.InstallMagiskApp(ctx)
	}
	return a.NeutralizeMDM(ctx)
}
