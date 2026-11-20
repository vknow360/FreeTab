package device

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// DeviceInfo contains live information queried from an ADB-connected device.
type DeviceInfo struct {
	Connected      bool   `json:"connected"`
	State          string `json:"state"`
	Serial         string `json:"serial"`
	Model          string `json:"model"`
	AndroidRelease string `json:"androidRelease"`
	VendorSDK      string `json:"vendorSdk"`
	RootGranted    bool   `json:"rootGranted"`
	Error          string `json:"error,omitempty"`
}

// ADB handles executing adb commands with streaming logs.
type ADB struct {
	ExePath string
	OnLog   func(string)
}

func NewADB(exePath string, logFunc func(string)) *ADB {
	return &ADB{
		ExePath: exePath,
		OnLog:   logFunc,
	}
}

func (a *ADB) log(msg string) {
	if a.OnLog != nil {
		a.OnLog(msg)
	}
}

// Run executes an adb command with streaming output.
func (a *ADB) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, a.ExePath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var output strings.Builder
	readPipe := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line + "\n")
			a.log(line)
		}
	}

	go readPipe(stdoutPipe)
	go readPipe(stderrPipe)

	err = cmd.Wait()
	return strings.TrimSpace(output.String()), err
}

// Shell runs a simple adb shell command and returns output.
func (a *ADB) Shell(ctx context.Context, shellCmd string) (string, error) {
	cmd := exec.CommandContext(ctx, a.ExePath, "shell", shellCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Detect checks device status and retrieves properties.
func (a *ADB) Detect(ctx context.Context) *DeviceInfo {
	info := &DeviceInfo{}
	if a.ExePath == "" {
		info.Error = "adb.exe not found"
		return info
	}

	cmd := exec.CommandContext(ctx, a.ExePath, "devices")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		info.Error = err.Error()
		return info
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") || strings.HasPrefix(line, "*") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			info.Serial = parts[0]
			info.State = parts[1]
			if parts[1] == "device" {
				info.Connected = true
			}
			break
		}
	}

	if !info.Connected {
		return info
	}

	// Query device properties
	if model, err := a.Shell(ctx, "getprop ro.product.model"); err == nil && model != "" {
		info.Model = model
	}
	if rel, err := a.Shell(ctx, "getprop ro.build.version.release"); err == nil && rel != "" {
		info.AndroidRelease = rel
	}
	if sdk, err := a.Shell(ctx, "getprop ro.vendor.build.version.sdk"); err == nil && sdk != "" {
		info.VendorSDK = sdk
	}

	// Check if root is granted via su
	suCheck, _ := a.Shell(ctx, "su -c id")
	if strings.Contains(suCheck, "uid=0") {
		info.RootGranted = true
	}

	return info
}

// WaitForDevice polls until an ADB device is connected and authorized or timeout occurs.
func (a *ADB) WaitForDevice(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	a.log("  [..] Scanning for tablet via ADB (USB Debugging)...")

	warned := false
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info := a.Detect(ctx)
		if info.Connected {
			a.log("  ✓ Tablet connected via ADB (" + info.Serial + ")")
			return nil
		}

		if info.State == "unauthorized" && !warned {
			a.log("  ⚠️ Device unauthorized! Look at tablet screen and tap 'Always allow from this computer'.")
			warned = true
		}

		time.Sleep(1500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for ADB device")
}

// RootShell runs a multi-line shell script with root privileges via su -c
func (a *ADB) RootShell(ctx context.Context, script string) error {
	tmpFile, err := os.CreateTemp("", "freetab_cmd_*.sh")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	cleanScript := strings.ReplaceAll(script, "\r\n", "\n")
	if _, err := tmpFile.WriteString(cleanScript); err != nil {
		return err
	}
	tmpFile.Close()

	// Push to /data/local/tmp/
	pushCmd := exec.CommandContext(ctx, a.ExePath, "push", tmpFile.Name(), "/data/local/tmp/_freetab.sh")
	pushCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push command script: %w", err)
	}

	// Execute with su and remove
	execCmd := exec.CommandContext(ctx, a.ExePath, "shell", "su -c 'sh /data/local/tmp/_freetab.sh; rm -f /data/local/tmp/_freetab.sh'")
	execCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	out, err := execCmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if outStr != "" {
		a.log(outStr)
	}
	return err
}

// InstallAPK installs an APK to the connected device.
func (a *ADB) InstallAPK(ctx context.Context, apkPath string) error {
	_, err := a.Run(ctx, "install", "-r", apkPath)
	return err
}
