//go:build !windows

package main

import "os"

// prepareConsole reports whether stdout is a terminal that can show colors.
// The returned function is a no-op on this platform.
func prepareConsole() (colors bool, restore func()) {
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0, func() {}
}
