//go:build !windows

package main

import "os"

// prepareConsole reports whether stdout is a terminal that can show colors.
func prepareConsole() bool {
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
