//go:build !windows

package main

import (
	"os"
	"strings"
)

func isHiddenFile(name string, _ os.FileInfo) bool {
	return strings.HasPrefix(name, ".")
}
