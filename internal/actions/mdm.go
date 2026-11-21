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

// AuthorizeShellRoot performs Phase 2B: Prompts and verifies root permission for Shell (uid 2000).
func (a *Actions) AuthorizeShellRoot(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("🔑 STAGE 2B: Authorize Shell Root Access")
	a.log("==========================================================================")

	// 1. Wait for ADB device after reboot
	a.log("Waiting for tablet ADB connection...")
	if err := a.ADB.WaitForDevice(ctx, 60*time.Second); err != nil {
		return err
	}

	// 2. Check if already granted
	a.log("[1/2] Checking if Shell already has Superuser privileges...")
	idOut, _ := a.ADB.Shell(ctx, "su -c id")

	if strings.Contains(idOut, "uid=0") {
		a.log("  ✓ Shell root is ALREADY authorized! (uid=0)")
		a.log("  ✓ No prompt was shown because Magisk already granted permanent root to Shell.")
		// Ensure SQLite policy is set
		_, _ = a.ADB.Shell(ctx, `su -c 'magisk --sqlite "REPLACE INTO policies (uid,policy,until,logging,notification) VALUES (2000,2,0,1,1);"'`)
		a.log("\n==========================================================================")
		a.log("✓ ROOT VERIFIED! You can now proceed directly to [Stage 2C: Neutralize MDM]!")
		a.log("==========================================================================")
		return nil
	}

	// 3. Trigger Superuser Request Popup on Tablet Screen
	a.log("\n[2/2] Triggering Superuser Request on tablet screen...")
	// Launch Magisk first to ensure UI is ready
	_, _ = a.ADB.Shell(ctx, "monkey -p com.topjohnwu.magisk -c android.intent.category.LAUNCHER 1")

	a.log("\n--------------------------------------------------------------------------")
	a.log("👉 ACTION ON TABLET SCREEN:")
	a.log("   If a 'Superuser Request' popup appears: Tap 'GRANT' or 'ALLOW'.")
	a.log("")
	a.log("   ⚠️ IF NO PROMPT APPEARS (Magisk default setting blocks ADB by default):")
	a.log("   1. Open Magisk on your tablet.")
	a.log("   2. Tap the ⚙️ Settings gear icon in the top-right corner.")
	a.log("   3. Scroll down to 'Superuser' section.")
	a.log("   4. Tap 'Superuser Access' and change it from 'Apps only' to 'Apps and ADB'.")
	a.log("--------------------------------------------------------------------------")
	a.log("  [..] Waiting for root authorization...\n")

	// Trigger `su` asynchronously or in a loop
	granted := false
	for i := 0; i < 60; i++ {
		time.Sleep(1 * time.Second)
		check, _ := a.ADB.Shell(ctx, "su -c id")
		if strings.Contains(check, "uid=0") {
			granted = true
			break
		}
		// Every 10 seconds re-trigger monkey / su and show status
		if i > 0 && i%10 == 0 {
			a.log(fmt.Sprintf("  [..] Still waiting (%ds/60s)... Last response: %s", 60-i, strings.TrimSpace(check)))
			a.log("       Remember: In Magisk Settings (⚙️), set 'Superuser Access' to 'Apps and ADB'")
			_, _ = a.ADB.Shell(ctx, "monkey -p com.topjohnwu.magisk -c android.intent.category.LAUNCHER 1")
		}
	}

	if !granted {
		return fmt.Errorf("root access was not granted within 60 seconds. On your tablet: Open Magisk > Settings (⚙️) > Superuser Access > Select 'Apps and ADB', then click 2B again")
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
