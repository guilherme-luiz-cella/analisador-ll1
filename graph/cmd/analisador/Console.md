---
class: analisador.prepareConsole
path: cmd/analisador/console_windows.go
kind: cli
step: 1
---

# Console

`prepareConsole() (colors bool, restore func())`, with one file per platform (`console_windows.go` and `console_other.go`). It says whether the output can show ANSI colors and returns the function that undoes its console changes.

## Public functions
- `prepareConsole() (colors bool, restore func())`: on Windows switches the console to UTF-8 (code page 65001) and turns on virtual terminal processing through `kernel32`; `restore` puts back the original code pages and console mode. Elsewhere it checks that stdout is a character device and `restore` does nothing. Colors are false when the output is piped.

## Notes
The Windows file uses the standard `syscall` package; `golang.org/x/sys` is not allowed in this project.
