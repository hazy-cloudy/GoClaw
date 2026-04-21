# ClawPet API Notes

This document describes the interface shape currently used by the beige `petclaw` frontend and the desktop pet.

Current facts:

- the canonical console is `petclaw`
- the primary chat transport is the `pet` websocket
- chat sessions use frontend-generated `session_id`
- voice playback relies on backend `audio` push
- browser local TTS fallback is disabled

## Base Addresses

- launcher: `http://127.0.0.1:18800`
- gateway: `http://127.0.0.1:18790`
- petclaw console: `http://127.0.0.1:3000`
- desktop renderer: `http://127.0.0.1:5173`

## Frontend Boot Flow

Typical `petclaw` boot flow:

1. validate launcher auth session
2. check gateway status
3. run `/api/pet/setup` when needed
4. fetch token + `ws_url`
5. connect `pet` websocket
6. send `chat`
7. consume `init_status`, `ai_chat`, `audio`, `emotion_change`, etc.

## HTTP Endpoints

### Direct gateway token

- Method: `GET`
- Path: `/pet/token`
- Host: gateway `127.0.0.1:18790`
- Purpose: return `ws_url` and protocol data for websocket connection

Example response:

```json
{
  "enabled": true,
  "token": "",
  "ws_url": "ws://127.0.0.1:18790/pet/ws",
  "protocol": "pet"
}
```

### Launcher auth status

- Method: `GET`
- Path: `/api/auth/status`
- Host: launcher `127.0.0.1:18800`
- Purpose: `petclaw` uses this before chat bootstrap

### Gateway status

- Method: `GET`
- Path: `/api/gateway/status`
- Host: launcher `127.0.0.1:18800`

### Start gateway

- Method: `POST`
- Path: `/api/gateway/start`
- Host: launcher `127.0.0.1:18800`

### Gateway logs

- Method: `GET`
- Path: `/api/gateway/logs`
- Host: launcher `127.0.0.1:18800`

### Pet setup

- Method: `POST`
- Path: `/api/pet/setup`
- Host: launcher `127.0.0.1:18800`

Notes:

- if `/api/pet/setup` is unavailable, the frontend may try `/api/pico/setup`

### Pet token via launcher proxy

- Method: `GET`
- Path: `/api/pet/token`
- Host: launcher `127.0.0.1:18800`

Notes:

- if `/api/pet/token` is unavailable, the frontend may try `/api/pico/token`

## WebSocket

### Connection URL

- URL: `ws://127.0.0.1:18790/pet/ws`
- Source: `ws_url` returned from `/pet/token`

The frontend attaches both:

- `session`
- `session_id`

using the same frontend-generated local session id.

### Client request format

Current chat request shape:

```json
{
  "action": "chat",
  "data": {
    "text": "你好",
    "session_key": "session-1710000000000-abcdef123"
  },
  "request_id": "req-1-1710000000000"
}
```

Actions currently used or expected:

- `chat`
- `onboarding_config`
- `emotion_get`

### Action response format

Success example:

```json
{
  "status": "ok",
  "action": "emotion_get",
  "data": {
    "emotion": "neutral",
    "description": "平静"
  }
}
```

Error examples:

```json
{
  "status": "error",
  "action": "chat",
  "error": "LLM call failed"
}
```

or:

```json
{
  "status": "error",
  "action": "chat",
  "data": {
    "error": "LLM call failed"
  }
}
```

## Push Messages

### Envelope

```json
{
  "type": "push",
  "push_type": "ai_chat",
  "data": {},
  "timestamp": 1710000000,
  "is_final": false
}
```

### `init_status`

Purpose:

- initial session state
- used to decide whether onboarding should be shown

### `ai_chat`

Purpose:

- assistant streaming text
- assistant final text
- tool text blocks

The frontend now accepts `data` in 3 forms:

1. object
2. JSON string
3. raw text string

Example object payload:

```json
{
  "text": "你好呀",
  "emotion": "joy",
  "type": "text"
}
```

Final block:

```json
{
  "text": "完整回复",
  "type": "final"
}
```

Tool block:

```json
{
  "text": "正在查询日历",
  "type": "tool"
}
```

Notes:

- `petclaw` strips `{...}` noise fragments from assistant text
- long replies previously appeared missing because early parsing only assumed object payloads; string and JSON-string forms are now handled

### `audio`

Purpose:

- backend TTS audio chunks

Accepted frontend payload shapes:

1. object
2. JSON string
3. raw base64 string

Typical object payload:

```json
{
  "chat_id": 1,
  "type": "audio",
  "text": "<base64-audio>",
  "is_final": false
}
```

Final block:

```json
{
  "chat_id": 1,
  "type": "audio",
  "text": "<base64-audio>",
  "is_final": true
}
```

Error block:

```json
{
  "chat_id": 1,
  "type": "error",
  "text": "tts failed"
}
```

Frontend behavior:

- aggregate chunks by `chat_id`
- merge on `is_final=true`
- play the merged audio
- no automatic browser TTS fallback

### `emotion_change`

Purpose:

- update pet emotion state

### `action_trigger`

Purpose:

- trigger pet action or UI hints

### `heartbeat`

Purpose:

- connection keepalive

## Onboarding Flow

Current onboarding flow:

1. frontend connects the `pet` websocket
2. backend pushes `init_status`
3. if `need_config=true`, frontend shows onboarding
4. user submits `onboarding_config`
5. frontend sends `emotion_get`
6. onboarding closes and chat starts

Current validation rules:

- `pet_name`: `2-24`
- `pet_persona`: `8-300`
- `pet_persona_type`: `gentle | playful | cool`
- onboarding submit is blocked when backend is unavailable

## Sessions And History

Current `petclaw` session behavior:

- each chat session uses a frontend-generated local session id
- `New Chat` generates a fresh session id and reconnects websocket
- left-side history shows only the first user message from each session
- this is not a server-persisted history API yet

## Troubleshooting Checklist

1. `GET http://127.0.0.1:18800/api/gateway/status` returns successfully
2. `GET http://127.0.0.1:18790/health` returns `200`
3. `GET http://127.0.0.1:18790/pet/token` returns `200`
4. DevTools shows `ws://127.0.0.1:18790/pet/ws?...session_id=...`
5. if there is text but no voice, check whether `push_type=audio` was actually sent
6. if long replies appear missing, inspect the raw `ai_chat` payload shape in WS frames

---

## Skills Marketplace API

Base URL: `http://127.0.0.1:18800`

### `GET /api/skills`

List installed skills from builtin/global/workspace sources.

```json
{
  "skills": [
    {
      "name": "github",
      "path": "D:/workspace/skills/github/SKILL.md",
      "source": "workspace",
      "description": "GitHub integration",
      "origin_kind": "third_party",
      "registry_name": "clawhub",
      "registry_url": "https://clawhub.ai/skills/github",
      "installed_version": "1.2.3",
      "installed_at": 1770000000000
    }
  ]
}
```

Notes:

- `source`: `builtin | global | workspace`
- `origin_kind`: `builtin | manual | third_party`

### `GET /api/skills/{name}`

Return detail for one skill. Response fields are the same as list item plus:

- `content` (the markdown body with frontmatter stripped)

### `GET /api/skills/search`

Search registry-backed market skills.

Query params:

- `q` (optional string; empty returns empty result list)
- `limit` (optional int, default `20`, valid range `1..50`)
- `offset` (optional int, default `0`)

Example:

`/api/skills/search?q=github&limit=20&offset=0`

```json
{
  "results": [
    {
      "score": 0.95,
      "slug": "github",
      "display_name": "GitHub",
      "summary": "GitHub integration skill",
      "version": "1.2.3",
      "registry_name": "clawhub",
      "url": "https://clawhub.ai/skills/github",
      "installed": true,
      "installed_name": "github"
    }
  ],
  "limit": 20,
  "offset": 0,
  "next_offset": 20,
  "has_more": true
}
```

### `POST /api/skills/install`

Install a skill from a registry.

Request JSON:

```json
{
  "slug": "github",
  "registry": "clawhub",
  "version": "1.2.3",
  "force": false
}
```

Response JSON:

```json
{
  "status": "ok",
  "slug": "github",
  "registry": "clawhub",
  "version": "1.2.3",
  "summary": "GitHub integration skill",
  "is_suspicious": false,
  "skill": {
    "name": "github",
    "path": "D:/workspace/skills/github/SKILL.md",
    "source": "workspace",
    "description": "GitHub integration skill",
    "origin_kind": "third_party",
    "registry_name": "clawhub",
    "registry_url": "https://clawhub.ai/skills/github",
    "installed_version": "1.2.3",
    "installed_at": 1770000000000
  }
}
```

### `POST /api/skills/import`

Import local skill file.

- Content-Type: `multipart/form-data`
- Form field: `file`
- Accepted: `.md` / `.zip`
- Current size limit: `1MB`

Response: imported skill object (same shape as `skill` in install response).

### `DELETE /api/skills/{name}`

Delete one installed skill by name.

```json
{
  "status": "ok"
}
```

Only workspace skills can be deleted.

### Skills API Error Notes

- `400`: invalid request or related skills tools disabled
- `404`: skill not found
- `409`: import/install target already exists
- `502`: upstream registry call failed
