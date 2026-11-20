package device

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetPlatformToolsURL returns the official Google platform-tools download URL for the target OS.
func GetPlatformToolsURL() string {
	switch runtime.GOOS {
	case "darwin":
		return "https://dl.google.com/android/repository/platform-tools-latest-darwin.zip"
	case "linux":
		return "https://dl.google.com/android/repository/platform-tools-latest-linux.zip"
	default:
		return "https://dl.google.com/android/repository/platform-tools-latest-windows.zip"
	}
}

// ToolPaths holds paths to the detected adb and fastboot binaries.
type ToolPaths struct {
	Adb      string `json:"adb"`
	Fastboot string `json:"fastboot"`
}

// GetCacheDir returns the default user cache directory for FreeTab tools.
func GetCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOME")
		}
		if home == "" {
			home = "."
		}
	}
	return filepath.Join(home, ".freetab", "cache")
}

// FindTools searches for adb and fastboot in standard locations.
func FindTools() ToolPaths {
	var paths ToolPaths

	home, _ := os.UserHomeDir()
	appDir, _ := os.Getwd()
	execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	candidates := []string{
		filepath.Join(execDir, "tools", "platform-tools"),
		filepath.Join(appDir, "tools", "platform-tools"),
		filepath.Join(GetCacheDir(), "platform-tools"),
	}
	if home != "" {
		candidates = append(candidates, filepath.Join(home, "platform-tools"))
	}

	adbBin := "adb"
	fastbootBin := "fastboot"
	if runtime.GOOS == "windows" {
		adbBin = "adb.exe"
		fastbootBin = "fastboot.exe"
	}

	for _, dir := range candidates {
		adbCandidate := filepath.Join(dir, adbBin)
		fastbootCandidate := filepath.Join(dir, fastbootBin)

		if paths.Adb == "" {
			if fi, err := os.Stat(adbCandidate); err == nil && !fi.IsDir() {
				paths.Adb = adbCandidate
			}
		}
		if paths.Fastboot == "" {
			if fi, err := os.Stat(fastbootCandidate); err == nil && !fi.IsDir() {
				paths.Fastboot = fastbootCandidate
			}
		}
	}

	// Fallback to system PATH
	if paths.Adb == "" {
		if p, err := exec.LookPath(adbBin); err == nil {
			paths.Adb = p
		}
	}
	if paths.Fastboot == "" {
		if p, err := exec.LookPath(fastbootBin); err == nil {
			paths.Fastboot = p
		}
	}

	return paths
}

// EnsureTools verifies tools exist. If either is missing, it attempts to download and extract them.
func EnsureTools(logFunc func(string)) (ToolPaths, error) {
	tools := FindTools()
	if tools.Adb != "" && tools.Fastboot != "" {
		logFunc(fmt.Sprintf("✓ Platform tools found:\n  ADB: %s\n  Fastboot: %s", tools.Adb, tools.Fastboot))
		return tools, nil
	}

	logFunc("⚠️ Platform tools missing. Downloading official Android platform-tools from Google...")
	cacheDir := GetCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return tools, fmt.Errorf("failed to create cache dir %s: %w", cacheDir, err)
	}

	downloadURL := GetPlatformToolsURL()
	zipPath := filepath.Join(cacheDir, "platform-tools.zip")
	logFunc(fmt.Sprintf("Downloading %s ...", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return tools, fmt.Errorf("download error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tools, fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return tools, fmt.Errorf("failed to save zip: %w", err)
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return tools, fmt.Errorf("failed to write zip file: %w", err)
	}

	logFunc("Extracting platform-tools.zip...")
	if err := unzip(zipPath, cacheDir); err != nil {
		return tools, fmt.Errorf("failed to extract zip: %w", err)
	}
	_ = os.Remove(zipPath)

	tools = FindTools()
	if tools.Adb == "" || tools.Fastboot == "" {
		return tools, fmt.Errorf("platform-tools downloaded but adb/fastboot could not be resolved")
	}

	logFunc("✓ Platform tools successfully downloaded and cached.")
	return tools, nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, 0755)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}

		// Ensure executables have permission to run on Unix platforms
		if runtime.GOOS != "windows" {
			baseName := filepath.Base(fpath)
			if baseName == "adb" || baseName == "fastboot" || f.Mode()&0111 != 0 {
				_ = os.Chmod(fpath, 0755)
			}
		}
	}
	return nil
}
