# PetClaw Console

`petclaw` is the current canonical beige console for ClawPet.

It is responsible for:

- the main chat UI
- onboarding at `/onboarding`
- local session history display
- gateway status display
- receiving and playing backend `audio` push

## Start

The recommended way is to start the full local stack from the repo root:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

If you only want this console:

```bash
npm install
npm run dev -- --hostname 127.0.0.1 --port 3000
```

Default URL:

- `http://127.0.0.1:3000`

Onboarding:

- `http://127.0.0.1:3000/onboarding?mode=rerun`

## Production Check

```bash
npm run build
npm run start -- --hostname 127.0.0.1 --port 3000
```

## Current Behavior

### Chat Sessions

- `New Chat` creates a new frontend session id
- left-side history shows only the first user input of each session
- history is currently frontend-local, not server-persisted history

### Voice

- backend `audio` push is the primary playback source
- browser local TTS fallback is disabled
- the microphone button is voice input and depends on browser speech recognition support

## Directory Roles

- `app/`: Next.js routes (`/` main console, `/onboarding` wizard)
- `components/`: UI and feature components
- `hooks/`: chat, gateway data, and voice input state
- `lib/`: API client, onboarding storage, utilities
- `styles/`: global/theme styles

## Common Troubleshooting

- `401 /api/pico/token`: launcher token or auth mismatch; restart with `scripts/run-goclaw-dev.ps1 -Restart`
- websocket cannot connect: verify launcher is listening at `127.0.0.1:18800`
- blank panel: confirm `petclaw` itself is running on `127.0.0.1:3000`
- long text appears missing: inspect the `ai_chat` frame shape in DevTools WS panel
- text reply but no voice: confirm backend actually sent `push_type=audio`
- desktop pet opens the wrong page: confirm `GOCLAW_DASHBOARD_URL` points to this console
