@echo off
rem Runs the LL(1) parser with a UTF-8 console so the epsilon shows correctly.
rem Full paths are used so it also works when Windows is set not to run
rem programs from the current folder (NoDefaultCurrentDirectoryInExePath).
chcp 65001 >nul
cd /d "%~dp0"
if exist "%~dp0analisador.exe" (
    "%~dp0analisador.exe"
) else (
    "%~dp0dist\analisador.exe"
)
