package device

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Fastboot handles executing fastboot commands with streaming logs.
type Fastboot struct {
	ExePath string
	OnLog   func(string)
}

func NewFastboot(exePath string, logFunc func(string)) *Fastboot {
	return &Fastboot{
		ExePath: exePath,
		OnLog:   logFunc,
	}
}

func (f *Fastboot) log(msg string) {
	if f.OnLog != nil {
		f.OnLog(msg)
	}
}

// Run executes a fastboot command with streaming output.
func (f *Fastboot) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, f.ExePath, args...)
	cmd.SysProcAttr = getSysProcAttr()

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
			f.log(line)
		}
	}

	go readPipe(stdoutPipe)
	go readPipe(stderrPipe)

	err = cmd.Wait()
	return strings.TrimSpace(output.String()), err
}

// Detect checks if a device is connected in Fastboot mode.
func (f *Fastboot) Detect(ctx context.Context) (bool, string) {
	if f.ExePath == "" {
		return false, ""
	}

	cmd := exec.CommandContext(ctx, f.ExePath, "devices")
	cmd.SysProcAttr = getSysProcAttr()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && (parts[1] == "fastboot" || parts[1] == "fastbootd") {
			return true, parts[0]
		}
	}
	return false, ""
}

// WaitForDevice polls until a fastboot device is detected or timeout occurs.
func (f *Fastboot) WaitForDevice(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	f.log("  [..] Scanning for tablet in FASTBOOT mode...")

	prompted := false
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		connected, serial := f.Detect(ctx)
		if connected {
			f.log("  ✓ Tablet detected in Fastboot mode (" + serial + ")")
			return nil
		}

		if !prompted {
			f.log("\n  👉 IF TABLET IS NOT DETECTED YET:")
			f.log("     1. Power OFF the tablet completely.")
			f.log("     2. Hold POWER + VOLUME UP until the boot menu appears.")
			f.log("     3. Use Volume Up to point to [Fastboot Mode], press Volume Down.")
			f.log("     4. Connect the USB cable firmly to your PC.\n")
			prompted = true
		}

		time.Sleep(1500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for Fastboot connection. Please check USB cable or drivers")
}

// FlashingUnlock sends the flashing unlock command.
func (f *Fastboot) FlashingUnlock(ctx context.Context) error {
	f.log("\n[1/3] Sending Bootloader Unlock Command...")
	f.log("  ⚠️  LOOK AT YOUR TABLET SCREEN NOW!")
	f.log("  👉 Press VOLUME UP on the tablet once to confirm unlocking!\n")

	_, err := f.Run(ctx, "flashing", "unlock")
	return err
}

// Flash flashes an image to a partition.
func (f *Fastboot) Flash(ctx context.Context, partition, imgPath string) error {
	f.log(fmt.Sprintf("Flashing %s with %s ...", partition, imgPath))
	_, err := f.Run(ctx, "flash", partition, imgPath)
	return err
}

// Reboot reboots the device into system.
func (f *Fastboot) Reboot(ctx context.Context) error {
	f.log("Rebooting tablet into system...")
	_, err := f.Run(ctx, "reboot")
	return err
}

// RebootFastbootd reboots from fastboot into userspace fastbootd.
func (f *Fastboot) RebootFastbootd(ctx context.Context) error {
	f.log("Rebooting into userspace Fastbootd...")
	_, err := f.Run(ctx, "reboot", "fastboot")
	return err
}
