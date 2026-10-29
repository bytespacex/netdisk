//go:build windows

package main

import (
	"os"
	"strings"
	"syscall"
)

func isHiddenFile(name string, fi os.FileInfo) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	if data, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok {
		const hiddenOrSystem = syscall.FILE_ATTRIBUTE_HIDDEN | syscall.FILE_ATTRIBUTE_SYSTEM
		return data.FileAttributes&hiddenOrSystem != 0
	}
	return false
}
