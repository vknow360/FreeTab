package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Partitions to backup
type PartitionBackup struct {
	Name       string
	DevicePath string
	FileName   string
}

// RunStockBackup dumps dynamic partitions via root adb and pulls them to outputDir.
func (a *Actions) RunStockBackup(ctx context.Context, outputDir string) error {
	a.log("==========================================================================")
	a.log("💾 STEP 4: Backup Stock Partitions")
	a.log("==========================================================================")

	if err := a.ADB.WaitForDevice(ctx, 30*time.Second); err != nil {
		return err
	}

	// Verify root
	idOut, _ := a.ADB.Shell(ctx, "su -c id")
	if !strings.Contains(idOut, "uid=0") {
		return fmt.Errorf("root access is required to dump stock partitions. Please grant root in Magisk first")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Determine active slot
	slot, _ := a.ADB.Shell(ctx, "getprop ro.boot.slot_suffix")
	slot = strings.TrimSpace(slot)
	if slot == "" {
		slot = "_b" // Default for Acer One 8
	}
	a.log(fmt.Sprintf("Active partition slot: %s", slot))

	partitions := []PartitionBackup{
		{Name: "system" + slot, DevicePath: "/dev/block/mapper/system" + slot, FileName: "stock_system.img"},
		{Name: "product" + slot, DevicePath: "/dev/block/mapper/product" + slot, FileName: "stock_product.img"},
		{Name: "system_ext" + slot, DevicePath: "/dev/block/mapper/system_ext" + slot, FileName: "stock_system_ext.img"},
	}

	for idx, part := range partitions {
		a.log(fmt.Sprintf("\n[%d/%d] Dumping %s partition...", idx+1, len(partitions), part.Name))
		tmpImg := fmt.Sprintf("/data/local/tmp/%s", part.FileName)

		// dd dump
		ddCmd := fmt.Sprintf("su -c 'dd if=%s of=%s bs=4M'", part.DevicePath, tmpImg)
		a.log("  [..] Reading partition to tablet temp storage...")
		if _, err := a.ADB.Shell(ctx, ddCmd); err != nil {
			return fmt.Errorf("failed to dump %s: %w", part.Name, err)
		}

		// adb pull
		localDest := filepath.Join(outputDir, part.FileName)
		a.log(fmt.Sprintf("  [..] Pulling %s to %s ...", part.FileName, localDest))
		if _, err := a.ADB.Run(ctx, "pull", tmpImg, localDest); err != nil {
			_, _ = a.ADB.Shell(ctx, fmt.Sprintf("rm -f %s", tmpImg))
			return fmt.Errorf("failed to pull %s: %w", part.FileName, err)
		}

		// remove tablet temp file
		_, _ = a.ADB.Shell(ctx, fmt.Sprintf("rm -f %s", tmpImg))
		a.log(fmt.Sprintf("  ✓ Successfully saved %s", part.FileName))
	}

	a.log("\n==========================================================================")
	a.log("🎉 ALL STOCK PARTITIONS BACKED UP SUCCESSFULLY!")
	a.log("Saved to: " + outputDir)
	a.log("==========================================================================")

	return nil
}
