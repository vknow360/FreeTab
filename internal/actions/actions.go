package actions

import (
	"freetab/internal/device"
)

// Actions orchestrates toolkit operations.
type Actions struct {
	ADB      *device.ADB
	Fastboot *device.Fastboot
	LogFunc  func(string)
}

func NewActions(adb *device.ADB, fb *device.Fastboot, logFunc func(string)) *Actions {
	return &Actions{
		ADB:      adb,
		Fastboot: fb,
		LogFunc:  logFunc,
	}
}

func (a *Actions) log(msg string) {
	if a.LogFunc != nil {
		a.LogFunc(msg)
	}
}
