//go:build !windows

package device

// CheckWindowsDrivers is a stub on non-Windows platforms.
func CheckWindowsDrivers() *DriverReport {
	return &DriverReport{
		Ready:   true,
		Summary: "Non-Windows host: USB drivers managed by udev / kernel.",
	}
}
