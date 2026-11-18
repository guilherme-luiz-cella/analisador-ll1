//go:build !windows

package main

import "os"

// prepareConsole diz se a saída é um terminal que mostra cores. A função
// devolvida não faz nada nesta plataforma.
func prepareConsole() (colors bool, restore func()) {
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0, func() {}
}
