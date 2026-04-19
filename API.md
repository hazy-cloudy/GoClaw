# ClawPet API 接口说明

本文档描述当前米色 `petclaw` 面板和桌宠窗口实际依赖的接口与消息格式。

当前事实：

- 主控制台是 `petclaw`
- 主聊天通道协议是 `pet`
- 聊天会话使用前端生成的 `session_id`
- 语音输出依赖后端 `audio` push
- 浏览器本地 TTS 自动兜底已关闭

补充：

- onboarding 的正式数据结构与接口契约见 [docs/zh/clawpet-onboarding.md](docs/zh/clawpet-onboarding.md)

## 基础地址

- launcher: `http://127.0.0.1:18800`
- gateway: `http://127.0.0.1:18790`
- petclaw 面板: `http://127.0.0.1:3000`
- 桌宠渲染: `http://127.0.0.1:5173`

## 前端启动链路

当前 `petclaw` 聊天前的典型流程：

1. 检查 launcher 鉴权状态
2. 检查 gateway 状态
3. 必要时执行 `/api/pet/setup`
4. 获取 token 与 `ws_url`
5. 连接 `pet` WebSocket
6. 发送 `chat`
7. 消费 `init_status`、`ai_chat`、`audio`、`emotion_change` 等 push

## HTTP 接口

### 1. 直接从 gateway 获取 pet 连接信息

- Method: `GET`
- Path: `/pet/token`
- Host: gateway `127.0.0.1:18790`
- 用途: 返回 `ws_url`，前端据此建立 WebSocket

示例响应：

```json
{
  "enabled": true,
  "token": "",
  "ws_url": "ws://127.0.0.1:18790/pet/ws",
  "protocol": "pet"
}
```

### 2. launcher 鉴权状态

- Method: `GET`
- Path: `/api/auth/status`
- Host: launcher `127.0.0.1:18800`
- 用途: `petclaw` 在聊天前确认 launcher session 是否可用

### 3. 网关状态

- Method: `GET`
- Path: `/api/gateway/status`
- Host: launcher `127.0.0.1:18800`

### 4. 启动网关

- Method: `POST`
- Path: `/api/gateway/start`
- Host: launcher `127.0.0.1:18800`

### 5. 获取网关日志

- Method: `GET`
- Path: `/api/gateway/logs`
- Host: launcher `127.0.0.1:18800`

### 6. 初始化 pet 通道

- Method: `POST`
- Path: `/api/pet/setup`
- Host: launcher `127.0.0.1:18800`

说明：

- 如果 `/api/pet/setup` 不存在，前端会尝试旧路径 `/api/pico/setup`

### 7. 通过 launcher 代理获取 pet token

- Method: `GET`
- Path: `/api/pet/token`
- Host: launcher `127.0.0.1:18800`

说明：

- 如果 `/api/pet/token` 不存在，前端会尝试旧路径 `/api/pico/token`

### 8. onboarding 提交

- Method: `POST`
- Path: `/api/pet/onboarding`
- Host: launcher `127.0.0.1:18800`

说明：

- `petclaw` onboarding 当前通过这个接口把完整快照提交给后端
- launcher 会把请求转发到 gateway 的 `/pet/onboarding`
- 如果新路径不可用，前端会回退尝试 `/api/pico/onboarding`

## WebSocket

### 1. 建连地址

- URL: `ws://127.0.0.1:18790/pet/ws`
- 来源: `/pet/token` 返回的 `ws_url`

前端会在 URL 上同时附带：

- `session`
- `session_id`

两个参数都使用同一个前端本地会话 ID。

### 2. 客户端请求格式

当前聊天请求格式：

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

当前前端已使用或依赖的 action：

- `chat`
- `onboarding_config`
- `emotion_get`

### 3. 动作响应格式

成功响应示例：

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

错误响应示例：

```json
{
  "status": "error",
  "action": "chat",
  "error": "LLM call failed"
}
```

或者：

```json
{
  "status": "error",
  "action": "chat",
  "data": {
    "error": "LLM call failed"
  }
}
```

## Push 消息

### 1. 通用包裹格式

```json
{
  "type": "push",
  "push_type": "ai_chat",
  "data": {},
  "timestamp": 1710000000,
  "is_final": false
}
```

### 2. `init_status`

用途：

- 首次连接后的初始化状态
- 用于决定是否展示 onboarding

### 3. `ai_chat`

用途：

- 助手流式文本
- 助手最终文本
- 工具文本块

当前前端兼容三种 `data` 形态：

1. 对象
2. JSON 字符串
3. 纯文本字符串

对象载荷示例：

```json
{
  "text": "你好呀",
  "emotion": "joy",
  "type": "text"
}
```

最终块：

```json
{
  "text": "完整回复",
  "type": "final"
}
```

工具块：

```json
{
  "text": "正在查询日历",
  "type": "tool"
}
```

说明：

- `petclaw` 会清理文本中的 `{...}` 噪声片段
- 长文本回复之前容易丢，是因为早期实现只假定 `data` 一定是对象；现在已经兼容对象、JSON 字符串、纯字符串

### 4. `audio`

用途：

- 后端 TTS 音频分片

当前前端兼容：

1. 对象格式
2. JSON 字符串格式
3. 纯 base64 字符串

典型对象：

```json
{
  "chat_id": 1,
  "type": "audio",
  "text": "<base64-audio>",
  "is_final": false
}
```

结束块：

```json
{
  "chat_id": 1,
  "type": "audio",
  "text": "<base64-audio>",
  "is_final": true
}
```

错误块：

```json
{
  "chat_id": 1,
  "type": "error",
  "text": "tts failed"
}
```

前端行为：

- 按 `chat_id` 聚合音频块
- 遇到 `is_final=true` 后合并分片并播放
- 当前不再启用浏览器本地 TTS 自动兜底

### 5. `emotion_change`

用途：

- 更新桌宠情绪状态

### 6. `action_trigger`

用途：

- 触发桌宠动作或 UI 提示

### 7. `heartbeat`

用途：

- 连接保活

## 初始化流程

当前 onboarding 流程：

1. 前端建立 `pet` WebSocket
2. 后端推送 `init_status`
3. 若 `need_config=true`，前端显示 onboarding 表单
4. 前端先执行 `runSetup()`，确保 launcher / gateway / pet channel 就绪
5. 前端调用 `POST /api/pet/onboarding`
6. 成功后保存本地 onboarding 快照
7. 关闭 onboarding，进入聊天

当前前端校验规则：

- `pet_name` 长度：`2-24`
- `pet_persona` 长度：`8-300`
- `pet_persona_type`：`gentle | playful | cool`
- 未连接后端时禁止提交

## 会话与历史

当前 `petclaw` 的会话行为：

- 每个聊天会话由前端本地生成一个 session ID
- `新建聊天` 会创建新的 session ID 并重连 WebSocket
- 左侧对话历史只显示每个会话的首条用户输入
- 当前不是服务端持久化历史接口

## 联调排错清单

1. `GET http://127.0.0.1:18800/api/gateway/status` 能正常返回
2. `GET http://127.0.0.1:18790/health` 返回 `200`
3. `GET http://127.0.0.1:18790/pet/token` 返回 `200`
4. DevTools 中应出现 `ws://127.0.0.1:18790/pet/ws?...session_id=...`
5. 若有文字无语音，先检查是否收到了 `push_type=audio`
6. 若长文本丢失，优先检查 `ai_chat` 的 `data` 实际是对象、JSON 字符串还是纯字符串
