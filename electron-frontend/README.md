# GoClaw Electron Frontend

This package renders the desktop pet window and opens the PetClaw console window.

## Start

Recommended: start the whole stack from repository root:

```powershell
powershell -ExecutionPolicy Bypass -File "scripts/run-goclaw-dev.ps1" -Restart
```

Local-only start from this directory:

```bash
npm install
npm run dev      # Vite renderer server (5173)
npm start        # Electron main process
```

`run.bat` will open both commands for convenience on Windows.

## Directory Roles

- `src/main.js`: Electron main process (pet window, console window)
- `src/preload.js`: secure renderer bridge (`window.electronAPI`)
- `src/App.tsx`: pet UI and bubble behavior
- `public/`: pet animation assets

## Common Troubleshooting

- Top menu (`File/Help`) appears: this should be disabled by app menu removal; restart Electron process
- Console window opens blank: ensure PetClaw is running at `GOCLAW_DASHBOARD_URL` (default `http://127.0.0.1:3000`)
- Onboarding should open in a new page: use console sidebar `重新初始化`
- Desktop pet does not respond: check Vite server is up on `127.0.0.1:5173`
