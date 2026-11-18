package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	enableVirtualTerminalProcessing = 0x0004
	codePageUTF8                    = 65001
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode  = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode  = kernel32.NewProc("SetConsoleMode")
	procGetConsoleCP    = kernel32.NewProc("GetConsoleCP")
	procGetConsoleOutCP = kernel32.NewProc("GetConsoleOutputCP")
	procSetConsoleCP    = kernel32.NewProc("SetConsoleCP")
	procSetConsoleOutCP = kernel32.NewProc("SetConsoleOutputCP")
)

// prepareConsole switches an attached Windows console to UTF-8 (so accents,
// the epsilon and box drawing display correctly whatever the code page was)
// and turns on ANSI escape handling. It reports whether colors can be used:
// false when stdout is not a console (for example, piped) or the mode can't
// be set. The returned function restores the original code pages; call it
// before the program exits.
func prepareConsole() (colors bool, restore func()) {
	restore = func() {}
	handle := os.Stdout.Fd()
	var mode uint32
	if ok, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode))); ok == 0 {
		return false, restore
	}

	oldOut, _, _ := procGetConsoleOutCP.Call()
	oldIn, _, _ := procGetConsoleCP.Call()
	if oldOut != 0 && oldOut != codePageUTF8 {
		if ok, _, _ := procSetConsoleOutCP.Call(codePageUTF8); ok != 0 {
			prev := restore
			restore = func() { prev(); procSetConsoleOutCP.Call(oldOut) }
		}
	}
	if oldIn != 0 && oldIn != codePageUTF8 {
		if ok, _, _ := procSetConsoleCP.Call(codePageUTF8); ok != 0 {
			prev := restore
			restore = func() { prev(); procSetConsoleCP.Call(oldIn) }
		}
	}

	ok, _, _ := procSetConsoleMode.Call(handle, uintptr(mode|enableVirtualTerminalProcessing))
	return ok != 0, restore
}
