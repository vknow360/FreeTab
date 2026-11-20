package assets

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed freetab_root_init_boot.img
var initBootData []byte

//go:embed Magisk.apk
var magiskApkData []byte

// ExtractInitBoot extracts the embedded freetab_root_init_boot.img to a temporary file.
// Returns the file path and a cleanup callback.
func ExtractInitBoot() (string, func(), error) {
	tmpDir := os.TempDir()
	outPath := filepath.Join(tmpDir, "freetab_root_init_boot.img")
	err := os.WriteFile(outPath, initBootData, 0644)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract freetab_root_init_boot.img: %w", err)
	}
	cleanup := func() {
		_ = os.Remove(outPath)
	}
	return outPath, cleanup, nil
}

// ExtractMagiskApk extracts the embedded Magisk.apk to a temporary file.
// Returns the file path and a cleanup callback.
func ExtractMagiskApk() (string, func(), error) {
	tmpDir := os.TempDir()
	outPath := filepath.Join(tmpDir, "freetab_Magisk.apk")
	err := os.WriteFile(outPath, magiskApkData, 0644)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract Magisk.apk: %w", err)
	}
	cleanup := func() {
		_ = os.Remove(outPath)
	}
	return outPath, cleanup, nil
}
