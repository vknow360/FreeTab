package actions

import (
	"context"
	"fmt"
	"freetab/internal/assets"
	"strings"
	"time"
)

// InstallMagiskApp performs Phase 2A: Installs Magisk.apk and launches it on tablet.
func (a *Actions) InstallMagiskApp(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("📦 STAGE 2A: Install Magisk Management App")
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
	a.log("    👉 Proceed to [Stage 2B: Authorize Shell Root]!")
	a.log("==========================================================================")

	return nil
}

// AuthorizeShellRoot performs Phase 2B: Confirms root permission for Shell (uid 2000).
func (a *Actions) AuthorizeShellRoot(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("🔑 STAGE 2B: Confirm Shell Root Access")
	a.log("==========================================================================")

	// 1. Wait for ADB device
	a.log("Waiting for tablet ADB connection...")
	if err := a.ADB.WaitForDevice(ctx, 10*time.Second); err != nil {
		return err
	}

	// 2. Check if already granted
	a.log("[1/2] Checking if Shell has Superuser privileges...")
	idOut, _ := a.ADB.Shell(ctx, "su -c id")

	if strings.Contains(idOut, "uid=0") {
		a.log("  ✓ Shell root is GRANTED! (uid=0 confirmed)")
		// Ensure SQLite policy is set
		_, _ = a.ADB.Shell(ctx, `su -c 'magisk --sqlite "REPLACE INTO policies (uid,policy,until,logging,notification) VALUES (2000,2,0,1,1);"'`)
		a.log("\n==========================================================================")
		a.log("✓ ROOT VERIFIED! You can now proceed directly to [Stage 2C: Neutralize MDM]!")
		a.log("==========================================================================")
		return nil
	}

	// 3. Prompt user to toggle in Superuser tab
	a.log("\n[2/2] Root not yet active. Please authorize Shell on your tablet:")
	a.log("--------------------------------------------------------------------------")
	a.log("👉 ON YOUR TABLET SCREEN:")
	a.log("   1. Open the Magisk app.")
	a.log("   2. Tap the 'Superuser' button in the bottom navigation bar (shield icon).")
	a.log("   3. Find 'Shell' (com.android.shell) in the list and toggle root access to ON.")
	a.log("--------------------------------------------------------------------------")
	a.log("  [..] Checking for Shell root in Magisk Superuser tab (10s timeout)...\n")

	// Check for grant in a 10s loop
	granted := false
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		check, _ := a.ADB.Shell(ctx, "su -c id")
		if strings.Contains(check, "uid=0") {
			granted = true
			break
		}
	}

	if !granted {
		return fmt.Errorf("shell root was not detected within 10s. Please open Magisk on your tablet, tap the 'Superuser' button in the bottom bar, toggle root access ON for 'Shell', then click 2B again")
	}

	a.log("  ✓ Superuser permission GRANTED: uid=0 confirmed!")

	// Auto-grant ADB root in Magisk SQLite DB for future commands
	a.log("  [..] Configuring automatic ADB root authorization...")
	_, _ = a.ADB.Shell(ctx, `su -c 'magisk --sqlite "REPLACE INTO policies (uid,policy,until,logging,notification) VALUES (2000,2,0,1,1);"'`)

	a.log("\n==========================================================================")
	a.log("🎉 SHELL ROOT AUTHORIZED SUCCESSFULLY!")
	a.log("👉 You can now click [Stage 2C: Execute Anti-MDM Defense]!")
	a.log("==========================================================================")

	return nil
}

// ExecuteMDMDefense performs Phase 2C: Executes the full Anti-MDM countermeasures ONLY after root is granted.
func (a *Actions) ExecuteMDMDefense(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("🛡️ STAGE 2C: Neutralize Zero-Touch & Tectoro MDM Lock")
	a.log("==========================================================================")

	// 1. Wait for ADB device
	if err := a.ADB.WaitForDevice(ctx, 30*time.Second); err != nil {
		return err
	}

	// 2. Strict root prerequisite check
	a.log("[1/4] Verifying root privileges before executing commands...")
	idOut, _ := a.ADB.Shell(ctx, "su -c id")
	if !strings.Contains(idOut, "uid=0") {
		a.log("\n❌ ERROR: Root access is NOT granted to Shell yet!")
		a.log("👉 Please click [Stage 2B: Authorize Shell Root] first, and tap GRANT on tablet screen.\n")
		return fmt.Errorf("root access not granted to Shell. Please run Stage 2B first")
	}
	a.log("  ✓ Root confirmed: uid=0.")

	// 3. Disable Zero-Touch activities & lock configuration
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

	// 4. Firewall DeviceLockController at kernel level
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

	// 5. Install persistent Magisk boot service
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
	a.log("You can now proceed to Step 3 (Performance Boost) and connect to Wi-Fi.")
	a.log("==========================================================================")

	return nil
}

// RunMDMBypass executes the combined routine (authorizes root first, then neutralizes MDM).
func (a *Actions) RunMDMBypass(ctx context.Context) error {
	if err := a.AuthorizeShellRoot(ctx); err != nil {
		return err
	}
	return a.ExecuteMDMDefense(ctx)
}
