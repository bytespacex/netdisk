package main

import (
	"os"
)

// movePath first tries an atomic rename. When source and destination are on
// different filesystems (Windows drives, macOS external disks, Linux mounts),
// rename fails; in that case it falls back to copy + remove.
func movePath(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !isCrossDeviceError(err) {
		return err
	}

	if err := copyPath(src, dst); err != nil {
		// Do not leave a partial copy behind.
		_ = os.RemoveAll(dst)
		return err
	}
	return os.RemoveAll(src)
}
