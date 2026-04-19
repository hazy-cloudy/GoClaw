# ClawPet Development Guide

ClawPet is the current desktop-pet product shape being developed in this repo.

The current source of truth is the beige `petclaw` console. The desktop pet opens this console, the onboarding page lives there, chat session history is shown there, and backend voice output is consumed there.

## Local Components

The local stack currently has 4 main parts:

- `launcher`: local launcher and control entry, default `127.0.0.1:18800`
- `gateway`: PicoClaw backend gateway, default `127.0.0.1:18790`
- `petclaw`: beige console UI, default `127.0.0.1:3000`
- `electron-frontend`: desktop pet renderer, default `127.0.0.1:5173`

Current expected relationships:

- the desktop pet `S` button opens `petclaw`
- onboarding lives at `petclaw /onboarding`
- `electron-frontend` is the pet window shell, not the main control console
- the old dark settings UI is no longer the target UI for daily use

## Recommended Startup

Use the full local dev script from the repo root:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

This script will:

1. Clean old GoClaw / Electron / Node processes
2. Start launcher on `127.0.0.1:18800`
3. Start the gateway
4. Start `petclaw` on `127.0.0.1:3000`
5. Start the desktop renderer on `127.0.0.1:5173`
6. Start Electron and point its settings/onboarding window to `petclaw`

## One-Click Startup

You can also use the root one-click entry:

- `GoClaw-OneClickStart.bat`

or:

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1
```

This entry currently delegates to `scripts/run-goclaw-dev.ps1`, so it starts the same full stack rather than a legacy simplified flow.

## Expected URLs

After a normal startup:

- launcher: `http://127.0.0.1:18800`
- petclaw console: `http://127.0.0.1:3000`
- onboarding page: `http://127.0.0.1:3000/onboarding?mode=rerun`
- desktop renderer: `http://127.0.0.1:5173`

## UI Rules

### Canonical UI

Use the beige `petclaw` UI as the main console:

- main console: `http://127.0.0.1:3000`
- onboarding: `http://127.0.0.1:3000/onboarding?mode=rerun`

`electron-frontend` is responsible for:

- rendering the desktop pet
- hosting the Electron shell
- opening the `petclaw` console

If you see a dark chat/settings page, that is not the current canonical UI.

### Chat Behavior

Current `petclaw` chat behavior:

- `New Chat` creates a new local frontend session id and reconnects the websocket
- left-side session history shows only the first user message for each session
- session history is currently frontend-local, not server-persisted history

### Voice Behavior

Current voice behavior:

- backend `audio` push is the primary source of playback
- `petclaw` merges backend audio chunks and plays them
- browser local TTS fallback is disabled
- the microphone button is voice input only and depends on browser speech recognition support

## Environment

- Go `>= 1.23`
- Node.js `>= 18`, recommended `20` or `22`
- npm
- on Windows, a C/C++ build environment is recommended or some SQLite capabilities may be degraded

## Common Commands

### Start full local dev stack

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### Start full stack with production `petclaw`

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode prod
```

### Start launcher only

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode launcher
```

## Troubleshooting

### Blank console window

Check that `petclaw` is actually running:

```powershell
curl.exe -i http://127.0.0.1:3000
```

If `3000` is down and you click the desktop pet `S` button, Electron will only show a blank shell window.

### Wrong UI opens

If a dark settings page opens instead of the beige console, restart with:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### Re-run onboarding

Open:

```text
http://127.0.0.1:3000/onboarding?mode=rerun
```

### Check gateway health

```powershell
curl.exe -i http://127.0.0.1:18790/health
curl.exe -i http://127.0.0.1:18790/ready
```

### Long replies missing

`petclaw` now accepts `ai_chat` and `audio` push payloads in object, JSON-string, and raw-string forms. If long replies still appear missing, inspect the WS payload shape in DevTools and compare it with the parser in `petclaw/lib/api/websocket.ts`.

## Important Directories

- `scripts/run-goclaw-dev.ps1`: full local startup script
- `GoClaw-OneClickStart.ps1`: one-click entry that delegates to the full startup script
- `petclaw/`: beige console and onboarding UI
- `electron-frontend/`: desktop pet window and Electron shell
- `cmd/picoclaw/`: gateway entry
- `pkg/channels/pet/`: `pet` HTTP / websocket channel
- `pkg/pet/`: desktop-pet business logic

## Suggested Reading Order

If you are onboarding to this codebase, this order is the easiest:

1. `scripts/run-goclaw-dev.ps1`
2. `petclaw/`
3. `electron-frontend/src/main.js`
4. `pkg/channels/pet/`

That gives the clearest picture of startup, canonical UI, pet window ownership, and the real chat pipeline.
