//go:build windows

package main

import (
	"errors"
	"syscall"
)

func isCrossDeviceError(err error) bool {
	// ERROR_NOT_SAME_DEVICE (winerror 17). It is not exported by syscall on
	// every Go release, so use the numeric value.
	const errorNotSameDevice = syscall.Errno(17)
	return errors.Is(err, errorNotSameDevice)
}
