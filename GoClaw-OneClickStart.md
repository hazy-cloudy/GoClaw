# GoClaw One Click Start

This file was recreated to provide a stable startup entry for this repository.

Current default one-click mode is `petclaw`.

## Quick start (recommended)

Run from `cmd.exe` in the repo root:

```cmd
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1
```

If you are not already in PowerShell:

- `cmd.exe`: run the same command directly
- `Git Bash` / `MSYS2` / `Nushell`: use `powershell.exe` instead of `powershell`

Example for Git Bash:

```bash
powershell.exe -ExecutionPolicy Bypass -File ./GoClaw-OneClickStart.ps1
```

This delegates to `scripts/run-goclaw-dev.ps1` and starts the full local stack:

- launcher: `127.0.0.1:18800`
- PetClaw dashboard: `127.0.0.1:3000`
- desktop pet renderer: `127.0.0.1:5173`

## Launcher mode

```cmd
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode launcher
```

This starts the Web UI launcher (`picoclaw-web.exe`).

## Alternative: dev mode

```cmd
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode dev
```

What it does:

1. Delegates to `scripts/run-goclaw-dev.ps1`.
2. Starts launcher + gateway.
3. Starts the beige `petclaw` dashboard with onboarding on `127.0.0.1:3000`.
4. Starts the Electron desktop pet and points its settings/onboarding window at the `petclaw` dashboard.

## Optional flags

- `-NoBrowser`: launcher mode only, do not auto-open browser.
- `-SkipNpmInstall`: ignored when delegating to `scripts/run-goclaw-dev.ps1`.

Examples:

```cmd
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
