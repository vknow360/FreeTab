package device

import (
	"testing"
)

func TestCheckWindowsDrivers(t *testing.T) {
	report := CheckWindowsDrivers()
	if report == nil {
		t.Fatal("expected non-nil DriverReport")
	}

	t.Logf("Driver Summary: %s", report.Summary)
	t.Logf("ADB Status: %s (Installed: %v, Seen: %v, Service: %s, Details: %s)",
		report.AdbDriver.Status, report.AdbDriver.Installed, report.AdbDriver.DeviceSeen, report.AdbDriver.Service, report.AdbDriver.Details)
	t.Logf("Fastboot Status: %s (Installed: %v, Seen: %v, Service: %s, Details: %s)",
		report.FastbootDriver.Status, report.FastbootDriver.Installed, report.FastbootDriver.DeviceSeen, report.FastbootDriver.Service, report.FastbootDriver.Details)
	t.Logf("MTK Status: %s (Installed: %v, Seen: %v, Service: %s, Details: %s)",
		report.MtkDriver.Status, report.MtkDriver.Installed, report.MtkDriver.DeviceSeen, report.MtkDriver.Service, report.MtkDriver.Details)
	t.Logf("Action Needed: %s", report.ActionNeeded)
}
