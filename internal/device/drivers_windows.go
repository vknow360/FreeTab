//go:build windows

package device

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const GoogleUSBDriverURL = "https://dl.google.com/android/repository/usb_driver_r13-windows.zip"

// CheckWindowsDrivers scans the Windows Registry and Driver Store to verify ADB, Fastboot, and MTK drivers.
func CheckWindowsDrivers() *DriverReport {
	report := &DriverReport{
		DownloadURL: GoogleUSBDriverURL,
		AdbDriver: DriverItem{
			Name:   "Android ADB Interface",
			Status: "Missing",
		},
		FastbootDriver: DriverItem{
			Name:   "Android Bootloader Interface (Fastboot)",
			Status: "Missing",
		},
		MtkDriver: DriverItem{
			Name:   "MediaTek USB VCOM / Preloader",
			Status: "Not Installed",
		},
	}

	// 1. Scan Windows Driver Store INF files in C:\Windows\INF
	infDir := filepath.Join(os.Getenv("SystemRoot"), "INF")
	if files, err := filepath.Glob(filepath.Join(infDir, "oem*.inf")); err == nil {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			content := string(data)

			if !report.AdbDriver.Installed && (strings.Contains(content, "SingleAdbInterface") || strings.Contains(content, "CompositeAdbInterface") || strings.Contains(content, "Android ADB Interface")) {
				report.AdbDriver.Installed = true
				report.AdbDriver.Details = filepath.Base(file)
			}
			if !report.FastbootDriver.Installed && (strings.Contains(content, "SingleBootLoaderInterface") || strings.Contains(content, "Android Bootloader Interface")) {
				report.FastbootDriver.Installed = true
				report.FastbootDriver.Details = filepath.Base(file)
			}
			if !report.MtkDriver.Installed && (strings.Contains(content, "MediaTek USB") || strings.Contains(content, "vcom_bootrom") || strings.Contains(content, "MT65xx Preloader")) {
				report.MtkDriver.Installed = true
				report.MtkDriver.Details = filepath.Base(file)
			}

			if report.AdbDriver.Installed && report.FastbootDriver.Installed && report.MtkDriver.Installed {
				break
			}
		}
	}

	// 2. Scan Windows Registry for active/registered USB Devices
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Enum\USB`, registry.READ)
	if err == nil {
		defer k.Close()
		subkeys, _ := k.ReadSubKeyNames(-1)

		for _, sk := range subkeys {
			upperSk := strings.ToUpper(sk)

			// Match Fastboot devices: Google (18D1:4EE0) or generic Fastboot PID
			isFastbootKey := strings.Contains(upperSk, "VID_18D1&PID_4EE0") || (strings.Contains(upperSk, "VID_0BB4&PID_0FFF"))
			// Match ADB devices: MTK (0E8D:201C, 0E8D:200E, 0E8D:201D) or Google (18D1:D001, 18D1:4EE7)
			isAdbKey := strings.Contains(upperSk, "VID_0E8D&PID_201C") || strings.Contains(upperSk, "VID_0E8D&PID_200E") ||
				strings.Contains(upperSk, "VID_0E8D&PID_201D") || strings.Contains(upperSk, "VID_18D1&PID_D001") ||
				strings.Contains(upperSk, "VID_18D1&PID_4EE7") || strings.Contains(upperSk, "VID_0BB4&PID_0C01")
			// Match MTK Preloader/BROM (0E8D:0003, 0E8D:2000)
			isMtkPreloaderKey := strings.Contains(upperSk, "VID_0E8D&PID_0003") || strings.Contains(upperSk, "VID_0E8D&PID_2000")

			if !isFastbootKey && !isAdbKey && !isMtkPreloaderKey {
				continue
			}

			devKey, err := registry.OpenKey(k, sk, registry.READ)
			if err != nil {
				continue
			}

			instances, _ := devKey.ReadSubKeyNames(-1)
			for _, inst := range instances {
				instKey, err := registry.OpenKey(devKey, inst, registry.READ)
				if err != nil {
					continue
				}

				desc, _, _ := instKey.GetStringValue("DeviceDesc")
				service, _, _ := instKey.GetStringValue("Service")
				cfgFlags, _, _ := instKey.GetIntegerValue("ConfigFlags")
				instKey.Close()

				// ConfigFlags bit 6 (0x40) represents CONFIGFLAG_FAILEDINSTALL (Code 28: Driver Missing)
				hasProblem := (cfgFlags&0x40 != 0) || (service == "" && strings.Contains(strings.ToLower(desc), "android"))

				if isFastbootKey {
					report.FastbootDriver.DeviceSeen = true
					if service != "" {
						report.FastbootDriver.Service = service
					}
					if hasProblem {
						report.FastbootDriver.Status = "Driver Error (Code 28)"
						report.ProblemFound = true
					} else if service != "" {
						report.FastbootDriver.Status = "OK"
					}
				}

				if isAdbKey {
					report.AdbDriver.DeviceSeen = true
					if service != "" {
						report.AdbDriver.Service = service
					}
					if hasProblem {
						report.AdbDriver.Status = "Driver Error (Code 28)"
						report.ProblemFound = true
					} else if service != "" {
						report.AdbDriver.Status = "OK"
					}
				}

				if isMtkPreloaderKey {
					report.MtkDriver.DeviceSeen = true
					if service != "" {
						report.MtkDriver.Service = service
					}
					if hasProblem {
						report.MtkDriver.Status = "Driver Error (Code 28)"
					} else if service != "" {
						report.MtkDriver.Status = "OK"
					}
				}
			}
			devKey.Close()
		}
	}

	// 3. Synthesize Status
	if report.AdbDriver.Status == "Missing" && report.AdbDriver.Installed {
		report.AdbDriver.Status = "Installed (Ready)"
	}
	if report.FastbootDriver.Status == "Missing" && report.FastbootDriver.Installed {
		report.FastbootDriver.Status = "Installed (Ready)"
	}
	if report.MtkDriver.Status == "Not Installed" && report.MtkDriver.Installed {
		report.MtkDriver.Status = "Installed (Ready)"
	}

	// Ready if ADB and Fastboot are either running (OK) or installed in Driver Store
	report.Ready = (report.AdbDriver.Status == "OK" || report.AdbDriver.Installed) &&
		(report.FastbootDriver.Status == "OK" || report.FastbootDriver.Installed) &&
		!report.ProblemFound

	if report.Ready {
		report.Summary = "✓ Windows USB Drivers are installed and healthy (ADB & Fastboot ready)."
	} else if report.ProblemFound {
		report.Summary = "⚠️ Driver installation error detected (Code 28 - Yellow exclamation in Device Manager)."
		report.ActionNeeded = "Fastboot/ADB driver is missing. Download Google USB Driver and update driver in Device Manager."
	} else if !report.FastbootDriver.Installed {
		report.Summary = "⚠️ Fastboot Driver (Android Bootloader Interface) not detected in Windows Driver Store."
		report.ActionNeeded = "Install the official Google USB Driver so Windows recognizes the tablet in Fastboot mode."
	} else {
		report.Summary = "⚠️ ADB or Fastboot drivers could not be verified in Windows Driver Store."
		report.ActionNeeded = "Please install the Google USB Driver before flashing."
	}

	return report
}
