@echo off
rem Runs the LL(1) parser with a UTF-8 console so the epsilon shows correctly.
chcp 65001 >nul
cd /d "%~dp0"
if exist "analisador.exe" (
    analisador.exe
) else (
    dist\analisador.exe
)
