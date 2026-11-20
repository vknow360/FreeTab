package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"freetab/internal/actions"
	"freetab/internal/device"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Overall device status sent to UI
type DeviceState struct {
	Mode           string             `json:"mode"` // "adb", "fastboot", "disconnected"
	ADBInfo        *device.DeviceInfo `json:"adbInfo,omitempty"`
	FastbootSerial string             `json:"fastbootSerial,omitempty"`
}

// App struct
type App struct {
	ctx          context.Context
	adb          *device.ADB
	fastboot     *device.Fastboot
	actions      *actions.Actions
	tools        device.ToolPaths
	runningMutex sync.Mutex
	isBusy       bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Emit logs to Wails frontend event stream
	logFunc := func(line string) {
		wailsRuntime.EventsEmit(a.ctx, "log", line)
	}

	// Initial tools lookup
	a.tools = device.FindTools()

	a.adb = device.NewADB(a.tools.Adb, logFunc)
	a.fastboot = device.NewFastboot(a.tools.Fastboot, logFunc)
	a.actions = actions.NewActions(a.adb, a.fastboot, logFunc)

	// Background device polling loop
	go a.devicePollLoop()
}

func (a *App) emitLog(msg string) {
	wailsRuntime.EventsEmit(a.ctx, "log", msg)
}

func (a *App) devicePollLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			// Do not interfere with device status checks during active operations
			if !a.isBusy && a.ctx != nil {
				state := a.DetectDevice()
				wailsRuntime.EventsEmit(a.ctx, "device_state", state)
			}
		}
	}
}

// CheckTools returns the paths of adb and fastboot binaries.
func (a *App) CheckTools() device.ToolPaths {
	a.tools = device.FindTools()
	return a.tools
}

// EnsureTools downloads and sets up tools if not present.
func (a *App) EnsureTools() (device.ToolPaths, error) {
	logFunc := func(line string) {
		a.emitLog(line)
	}

	tools, err := device.EnsureTools(logFunc)
	if err != nil {
		return tools, err
	}

	a.tools = tools
	a.adb.ExePath = tools.Adb
	a.fastboot.ExePath = tools.Fastboot

	return tools, nil
}

// DetectDevice queries connection status across ADB and Fastboot.
func (a *App) DetectDevice() DeviceState {
	state := DeviceState{Mode: "disconnected"}

	// 1. Check Fastboot first
	if a.fastboot != nil && a.fastboot.ExePath != "" {
		fbConnected, fbSerial := a.fastboot.Detect(a.ctx)
		if fbConnected {
			state.Mode = "fastboot"
			state.FastbootSerial = fbSerial
			return state
		}
	}

	// 2. Check ADB
	if a.adb != nil && a.adb.ExePath != "" {
		adbInfo := a.adb.Detect(a.ctx)
		if adbInfo.Connected {
			state.Mode = "adb"
			state.ADBInfo = adbInfo
			return state
		}
	}

	return state
}

// SelectDirectory opens a Windows native directory picker dialog.
func (a *App) SelectDirectory(title string) (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
	})
}

// RunStep1_Root executes Step 1 (Fastboot Unlock & Root).
func (a *App) RunStep1_Root() error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if a.fastboot.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("fastboot tool unavailable: %w", err)
		}
	}

	return a.actions.RunFastbootRoot(a.ctx)
}

// RunStep2A_InstallMagisk installs Magisk APK and launches setup on tablet.
func (a *App) RunStep2A_InstallMagisk() error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if a.adb.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("adb tool unavailable: %w", err)
		}
	}

	return a.actions.InstallMagiskApp(a.ctx)
}

// RunStep2B_NeutralizeMDM verifies root grant and eliminates MDM locks.
func (a *App) RunStep2B_NeutralizeMDM() error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if a.adb.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("adb tool unavailable: %w", err)
		}
	}

	return a.actions.NeutralizeMDM(a.ctx)
}

// RunStep2_MDM executes Step 2 (MDM / Zero-Touch neutralization).
func (a *App) RunStep2_MDM() error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if a.adb.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("adb tool unavailable: %w", err)
		}
	}

	return a.actions.RunMDMBypass(a.ctx)
}

// RunStep3_Performance executes Step 3 (0.5x animations, debloat, setup wizard finalization).
func (a *App) RunStep3_Performance() error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if a.adb.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("adb tool unavailable: %w", err)
		}
	}

	return a.actions.RunPerformanceBoost(a.ctx)
}

// RunStep4_Backup executes Step 4 (Stock dynamic partition dump).
func (a *App) RunStep4_Backup(outDir string) error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if outDir == "" {
		return fmt.Errorf("no backup output directory selected")
	}

	if a.adb.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("adb tool unavailable: %w", err)
		}
	}

	return a.actions.RunStockBackup(a.ctx, outDir)
}

// RunStep5_Restore executes Step 5 (Stock ROM restore via fastbootd).
func (a *App) RunStep5_Restore(backupDir string) error {
	a.runningMutex.Lock()
	a.isBusy = true
	defer func() {
		a.isBusy = false
		a.runningMutex.Unlock()
	}()

	if backupDir == "" {
		return fmt.Errorf("no backup source directory selected")
	}

	if a.fastboot.ExePath == "" {
		if _, err := a.EnsureTools(); err != nil {
			return fmt.Errorf("fastboot tool unavailable: %w", err)
		}
	}

	return a.actions.RunStockRestore(a.ctx, backupDir)
}
