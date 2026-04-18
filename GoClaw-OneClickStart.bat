@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
set "PS1=%SCRIPT_DIR%GoClaw-OneClickStart.ps1"

if not exist "%PS1%" (
  echo [ERROR] Script not found: "%PS1%"
  pause
  exit /b 1
)

echo [GoClaw] Starting with one click...
powershell -NoProfile -ExecutionPolicy Bypass -File "%PS1%" -Mode petclaw

set "EXIT_CODE=%ERRORLEVEL%"
if not "%EXIT_CODE%"=="0" (
  echo.
  echo [ERROR] Startup failed with code %EXIT_CODE%.
  pause
)

exit /b %EXIT_CODE%
