package device

const (
	GoogleUSBDriverURL = "https://dl.google.com/android/repository/usb_driver_r13-windows.zip"
	UniversalDriverURL = "https://adb.clockworkmod.com/"
)

// DriverItem represents the diagnostic state of an individual USB driver.
type DriverItem struct {
	Name       string `json:"name"`
	Installed  bool   `json:"installed"`
	DeviceSeen bool   `json:"deviceSeen"`
	Service    string `json:"service"`
	Status     string `json:"status"` // "OK", "Missing", "Driver Error (Code 28)"
	Details    string `json:"details"`
}

// DriverReport contains the full Windows USB driver diagnostic analysis.
type DriverReport struct {
	Ready          bool       `json:"ready"` // True if both ADB and Fastboot drivers are operational
	AdbDriver      DriverItem `json:"adbDriver"`
	FastbootDriver DriverItem `json:"fastbootDriver"`
	MtkDriver      DriverItem `json:"mtkDriver"`
	ProblemFound   bool       `json:"problemFound"`
	Summary        string     `json:"summary"`
	DownloadURL    string     `json:"downloadUrl"`
	ActionNeeded   string     `json:"actionNeeded"`
}
