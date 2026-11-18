---
class: analisador.prepareConsole
path: cmd/analisador/console_windows.go
kind: cli
step: 1
---

# Console

`prepareConsole() bool`, with one file per platform (`console_windows.go` and `console_other.go`). It says whether the output can show ANSI colors.

## Public functions
- `prepareConsole() bool`: on Windows turns on virtual terminal processing through `kernel32`; elsewhere checks that stdout is a character device. False when the output is piped.

## Notes
The Windows file uses the standard `syscall` package; `golang.org/x/sys` is not allowed in this project.
