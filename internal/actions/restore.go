package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunStockRestore flashes stock_system.img, stock_product.img, stock_system_ext.img via fastbootd.
func (a *Actions) RunStockRestore(ctx context.Context, backupDir string) error {
	a.log("==========================================================================")
	a.log("♻️ STEP 5: Restore Stock Partitions")
	a.log("==========================================================================")

	sysImg := filepath.Join(backupDir, "stock_system.img")
	prodImg := filepath.Join(backupDir, "stock_product.img")
	sysExtImg := filepath.Join(backupDir, "stock_system_ext.img")

	for _, p := range []string{sysImg, prodImg, sysExtImg} {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("missing partition image: %s", p)
		}
	}

	a.log("Checking Fastboot / Fastbootd connection...")
	if err := a.Fastboot.WaitForDevice(ctx, 45*time.Second); err != nil {
		return err
	}

	// Reboot to fastbootd userspace
	a.log("\n[1/4] Switching to Fastbootd (userspace dynamic partition mode)...")
	_ = a.Fastboot.RebootFastbootd(ctx)
	time.Sleep(10 * time.Second)

	// Wait for reconnection in fastbootd
	if err := a.Fastboot.WaitForDevice(ctx, 45*time.Second); err != nil {
		a.log("  ⚠️ Note: Continuing flash attempt...")
	}

	// Flash system
	a.log("\n[2/4] Flashing stock system partition...")
	if err := a.Fastboot.Flash(ctx, "system", sysImg); err != nil {
		return fmt.Errorf("failed to flash system: %w", err)
	}
	a.log("  ✓ System partition restored.")

	// Flash product
	a.log("\n[3/4] Flashing stock product partition...")
	if err := a.Fastboot.Flash(ctx, "product", prodImg); err != nil {
		return fmt.Errorf("failed to flash product: %w", err)
	}
	a.log("  ✓ Product partition restored.")

	// Flash system_ext
	a.log("\n[4/4] Flashing stock system_ext partition...")
	if err := a.Fastboot.Flash(ctx, "system_ext", sysExtImg); err != nil {
		return fmt.Errorf("failed to flash system_ext: %w", err)
	}
	a.log("  ✓ System_ext partition restored.")

	// Reboot
	a.log("\nRebooting tablet into restored stock Android...")
	_ = a.Fastboot.Reboot(ctx)

	a.log("\n==========================================================================")
	a.log("🎉 STOCK ROM RESTORE COMPLETED!")
	a.log("==========================================================================")

	return nil
}
