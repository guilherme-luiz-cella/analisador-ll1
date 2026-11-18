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

// prepareConsole põe o console do Windows em UTF-8 (para acentos, o ε e as
// molduras aparecerem certos qualquer que seja a página de código) e liga os
// códigos ANSI. Diz se as cores podem ser usadas: false quando a saída não é
// um console (por exemplo, redirecionada) ou o modo não pode ser mudado. A
// função devolvida restaura as páginas de código e o modo originais; chame-a
// antes de o programa sair.
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
	if ok != 0 {
		prev := restore
		restore = func() { prev(); procSetConsoleMode.Call(handle, uintptr(mode)) }
	}
	return ok != 0, restore
}
