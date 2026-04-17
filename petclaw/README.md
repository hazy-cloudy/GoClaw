# PetClaw Console

PetClaw is the web console for GoClaw. It provides chat, channel/config management, logs, and onboarding.

## Start

From this directory:

```bash
npm install
npm run dev
```

Default URL: `http://127.0.0.1:3000`

For production check:

```bash
npm run build
npm run start -- --hostname 127.0.0.1 --port 3000
```

## Directory Roles

- `app/`: Next.js routes (`/` main console, `/onboarding` wizard)
- `components/`: UI and feature components
- `hooks/`: stateful logic (chat, gateway data)
- `lib/`: API client, onboarding storage, utilities
- `styles/`: global/theme styles

## Common Troubleshooting

- `401 /api/pico/token`: launcher token or auth mismatch; restart with root script `scripts/run-goclaw-dev.ps1 -Restart`
- WebSocket cannot connect: verify launcher is listening at `127.0.0.1:18800`
- "重新初始化" button missing: hard refresh with `Ctrl+F5`
- Wrong page opens from desktop pet: ensure `GOCLAW_DASHBOARD_URL` points to this console URL
