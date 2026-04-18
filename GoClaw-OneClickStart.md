# GoClaw One Click Start

This file was recreated to provide a stable startup entry for this repository.

Current default one-click mode is `petclaw`.

## Quick start (recommended)

Double-click:

- `GoClaw-OneClickStart.bat`

or run from terminal:

From `D:\opencode\GoClaw` run:

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1
```

This starts backend + PetClaw frontend and opens:

- `http://localhost:3000/onboarding`

## Launcher mode

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode launcher
```

This starts the Web UI launcher (`picoclaw-web.exe`).

## Alternative: dev mode (backend + electron frontend)

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode dev
```

What it does:

1. Checks `go` and `npm` are available.
2. Installs frontend dependencies if `electron-frontend\node_modules` is missing.
3. Opens two PowerShell windows:
   - backend: `go run -tags "goolm,stdjson" ./cmd/picoclaw gateway`
   - frontend: `npm run start` in `electron-frontend`

## Optional flags

- `-NoBrowser`: launcher mode only, do not auto-open browser.
- `-SkipNpmInstall`: skip auto `npm install` (dev / petclaw modes).

Examples:

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -NoBrowser
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode dev -SkipNpmInstall
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode petclaw
```

## If script execution is blocked

If Windows blocks local scripts, run this once in an elevated PowerShell:

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

Then run the start command again.
