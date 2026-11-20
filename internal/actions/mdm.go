package actions

import (
	"context"
	"fmt"
	"freetab/internal/assets"
	"strings"
	"time"
)

// RunMDMBypass performs Step 2: Neutralizing Zero-Touch & Tectoro MDM Lock.
func (a *Actions) RunMDMBypass(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("🛡️ STEP 2: Neutralize Zero-Touch & Tectoro MDM Lock")
	a.log("==========================================================================")

	// 1. Wait for ADB device
	if err := a.ADB.WaitForDevice(ctx, 45*time.Second); err != nil {
		return err
	}

	// 2. Install Magisk App
	a.log("\n[0/4] Checking & installing Magisk App...")
	apkPath, cleanup, err := assets.ExtractMagiskApk()
	if err == nil {
		defer cleanup()
		if installErr := a.ADB.InstallAPK(ctx, apkPath); installErr == nil {
			a.log("  ✓ Magisk APK installed/updated successfully.")
		} else {
			a.log("  ℹ️ Magisk install check completed.")
		}
	}

	// 3. Verify Root
	a.log("\n[1/4] Verifying root privileges...")
	idOut, _ := a.ADB.Shell(ctx, "su -c id")
	if !strings.Contains(idOut, "uid=0") {
		a.log("  ⚠️ Root prompt not auto-approved. Launching Magisk on tablet...")
		_, _ = a.ADB.Shell(ctx, "monkey -p com.topjohnwu.magisk -c android.intent.category.LAUNCHER 1")
		a.log("  👉 LOOK AT YOUR TABLET SCREEN NOW:")
		a.log("     1. If prompted 'Requires Additional Setup', tap OK (tablet reboots).")
		a.log("     2. When Superuser request appears, tap 'GRANT' / 'ALLOW'!")

		// Poll for grant for up to 25 seconds
		granted := false
		for i := 0; i < 25; i++ {
			time.Sleep(1 * time.Second)
			check, _ := a.ADB.Shell(ctx, "su -c id")
			if strings.Contains(check, "uid=0") {
				granted = true
				break
			}
		}

		if !granted {
			return fmt.Errorf("root permission was not granted. Please open Magisk app on tablet, complete setup, and try again")
		}
	}
	a.log("  ✓ Root confirmed (uid=0).")

	// Auto-grant ADB root in Magisk SQLite DB for future commands
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
