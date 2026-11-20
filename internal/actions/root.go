package actions

import (
	"context"
	"fmt"
	"freetab/internal/assets"
	"strings"
	"time"
)

// RunFastbootRoot performs Step 1: Bootloader unlock and Magisk root flashing.
func (a *Actions) RunFastbootRoot(ctx context.Context) error {
	a.log("==========================================================================")
	a.log("⚡ STEP 1: Fastboot Bootloader Unlock & Magisk Root")
	a.log("==========================================================================")

	// 1. Extract embedded image
	a.log("Extracting embedded root image (freetab_root_init_boot.img)...")
	imgPath, cleanup, err := assets.ExtractInitBoot()
	if err != nil {
		return fmt.Errorf("failed to extract freetab_root_init_boot.img: %w", err)
	}
	defer cleanup()

	// 2. Check if device is currently connected via ADB; if so, auto-reboot to Fastboot
	if a.ADB != nil && a.ADB.ExePath != "" {
		adbInfo := a.ADB.Detect(ctx)
		if adbInfo.Connected {
			a.log("Detected tablet connected via ADB (Android). Sending 'adb reboot bootloader'...")
			_, _ = a.ADB.Run(ctx, "reboot", "bootloader")
			a.log("  [..] Tablet is restarting into Fastboot mode...")
			time.Sleep(4 * time.Second)
		}
	}

	// 3. Wait for Fastboot connection
	if err := a.Fastboot.WaitForDevice(ctx, 60*time.Second); err != nil {
		return err
	}

	// 3. Send Unlock Command
	unlockOut, unlockErr := a.Fastboot.Run(ctx, "flashing", "unlock")
	if unlockErr != nil {
		if strings.Contains(unlockOut, "Unlock operation is not allowed") {
			a.log("\n❌ ERROR: BOOTLOADER UNLOCK IS BLOCKED BY OEM RESTRICTION!")
			a.log("--------------------------------------------------------------------------")
			a.log("👉 HOW TO RESOLVE (MANDATORY STEPS ON TABLET):")
			a.log(" 1. Power OFF tablet. Hold POWER + VOLUME DOWN to enter Factory Mode.")
			a.log(" 2. Select 'Clear Storage' to format userdata clean.")
			a.log(" 3. Boot into Android OFFLINE (Do NOT connect to Wi-Fi!).")
			a.log(" 4. Settings > About tablet -> Tap 'Build number' 7 times.")
			a.log(" 5. Settings > System > Developer options:")
			a.log("    - Toggle 'OEM unlocking' to ON (Tap Enable on prompt)")
			a.log("    - Toggle 'USB debugging' to ON")
			a.log(" 6. Power OFF, hold POWER + VOLUME UP -> choose [Fastboot Mode], and retry.\n")
			return fmt.Errorf("bootloader unlock blocked: OEM unlocking must be enabled in Android Developer Options")
		} else {
			a.log("  ⚠️ Note: " + strings.TrimSpace(unlockOut))
		}
	} else {
		a.log("  ✓ Bootloader unlock command processed.")
	}

	time.Sleep(3 * time.Second)

	// 4. Flash to slots A & B
	a.log("\n[2/3] Flashing FreeTab GKI Root Image to both slots...")
	flashAOut, flashAErr := a.Fastboot.Run(ctx, "flash", "init_boot_a", imgPath)
	if flashAErr != nil {
		if strings.Contains(flashAOut, "not allowed in locked state") {
			a.log("\n❌ ERROR: Cannot flash because bootloader is still in LOCKED state!")
			a.log("👉 Please make sure OEM unlocking is enabled, and press VOLUME UP on tablet screen to confirm unlock.\n")
		}
		return fmt.Errorf("failed to flash init_boot_a: %w", flashAErr)
	}
	a.log("  ✓ init_boot_a flashed successfully.")

	_, flashBErr := a.Fastboot.Run(ctx, "flash", "init_boot_b", imgPath)
	if flashBErr != nil {
		return fmt.Errorf("failed to flash init_boot_b: %w", flashBErr)
	}
	a.log("  ✓ init_boot_b flashed successfully.")

	// 5. Reboot into Android
	a.log("\n[3/3] Rebooting tablet into Android...")
	if err := a.Fastboot.Reboot(ctx); err != nil {
		a.log("  ⚠️ Warning: reboot command returned: " + err.Error())
	} else {
		a.log("  ✓ Tablet is rebooting!")
	}

	a.log("\n==========================================================================")
	a.log("🎉 ROOT FLASH COMPLETED SUCCESSFULLY!")
	a.log("==========================================================================")
	a.log("Next Steps:")
	a.log(" 1. Let tablet boot and complete initial setup wizard OFFLINE (Wi-Fi OFF!).")
	a.log(" 2. Open Settings > About tablet -> Tap 'Build number' 7 times.")
	a.log(" 3. Open Settings > System > Developer options -> Enable 'USB debugging'.")
	a.log(" 4. Connect tablet to PC and proceed to Step 2 (MDM Neutralizer).")
	a.log("==========================================================================")

	return nil
}
