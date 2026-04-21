# GoClaw API 接口文档

> 版本：v2.5  
> 日期：2026-04-20  
> 协议：WebSocket + JSON + HTTP

---

## 一、服务架构

### 1.1 核心组件

GoClaw 由以下 4 个核心组件组成：

| 组件 | 端口 | 说明 |
|------|------|------|
| Launcher | 18800 | 本地服务管理器，负责启动和监控 |
| Gateway | 18790 | PicoClaw 后端网关，处理 AI 对话 |
| Petclaw | 3000 | 桌宠控制台（Web UI） |
| Electron | 5173 | 桌面宠物渲染窗口 |

### 1.2 服务地址

- Launcher: `http://127.0.0.1:18800`
- Gateway: `http://127.0.0.1:18790`
- Petclaw 控制台: `http://127.0.0.1:3000`
- 桌面渲染: `http://127.0.0.1:5173`

---

## 二、前端启动流程

典型的 Petclaw 启动流程：

1. 验证 Launcher 认证会话
2. 检查 Gateway 状态
3. 必要时调用 `/api/pet/setup` 初始化
4. 获取 Token 和 WebSocket URL
5. 连接 Pet WebSocket
6. 发送聊天消息
7. 接收推送：`init_status`、`ai_chat`、`audio_and_voice`、`emotion_change` 等

---

## 三、HTTP 接口

### 3.1 获取 Gateway Token

- **方法**: `GET`
- **路径**: `/pet/token`
- **主机**: Gateway `127.0.0.1:18790`
- **用途**: 返回 WebSocket 连接地址和协议信息

**响应示例**：

```json
{
  "enabled": true,
  "token": "",
  "ws_url": "ws://127.0.0.1:18790/pet/ws",
  "protocol": "pet"
}
```

---

### 3.2 Launcher 认证状态

- **方法**: `GET`
- **路径**: `/api/auth/status`
- **主机**: Launcher `127.0.0.1:18800`
- **用途**: Petclaw 在聊天初始化前检查认证状态

---

### 3.3 Gateway 状态

- **方法**: `GET`
- **路径**: `/api/gateway/status`
- **主机**: Launcher `127.0.0.1:18800`

---

### 3.4 启动 Gateway

- **方法**: `POST`
- **路径**: `/api/gateway/start`
- **主机**: Launcher `127.0.0.1:18800`

---

### 3.5 获取 Gateway 日志

- **方法**: `GET`
- **路径**: `/api/gateway/logs`
- **主机**: Launcher `127.0.0.1:18800`

---

### 3.6 Pet 初始化

- **方法**: `POST`
- **路径**: `/api/pet/setup`
- **主机**: Launcher `127.0.0.1:18800`

**注意**：如果 `/api/pet/setup` 不可用，前端会尝试 `/api/pico/setup`

---

### 3.7 通过 Launcher 获取 Pet Token

- **方法**: `GET`
- **路径**: `/api/pet/token`
- **主机**: Launcher `127.0.0.1:18800`

**注意**：如果 `/api/pet/token` 不可用，前端会尝试 `/api/pico/token`

---

### 3.8 会话管理

#### 获取会话列表

- **方法**: `GET`
- **路径**: `/api/sessions?offset=0&limit=20`
- **主机**: Launcher `127.0.0.1:18800`

**响应示例**：

```json
[
  {
    "id": "session-001",
    "title": "第一次对话",
    "preview": "你好呀...",
    "message_count": 10,
    "created": "2024-01-15T10:30:00Z",
    "updated": "2024-01-15T11:00:00Z"
  }
]
```

#### 获取会话历史

- **方法**: `GET`
- **路径**: `/api/sessions/{id}`
- **主机**: Launcher `127.0.0.1:18800`

**响应示例**：

```json
{
  "id": "session-001",
  "messages": [
    {
      "role": "user",
      "content": "你好呀"
    },
    {
      "role": "assistant",
      "content": "你好！今天心情怎么样？"
    }
  ],
  "summary": "用户打招呼",
  "created": "2024-01-15T10:30:00Z",
  "updated": "2024-01-15T11:00:00Z"
}
```

#### 删除会话

- **方法**: `DELETE`
- **路径**: `/api/sessions/{id}`
- **主机**: Launcher `127.0.0.1:18800`

---

## 四、WebSocket 接口

### 4.1 连接方式

**连接地址**：
```
ws://127.0.0.1:18790/pet/ws?session={sessionId}
```

**参数说明**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sessionId | string | 否 | 会话ID，由前端生成 |

**连接示例**：
```javascript
const ws = new WebSocket('ws://127.0.0.1:18790/pet/ws?session=user_001');
```

**重要**：连接建立后，服务器会主动推送 `init_status`，告知前端是否需要初始化配置。

---

### 4.2 消息格式

#### 请求消息（客户端 → 服务器）

```json
{
  "action": "string",
  "data": {},
  "request_id": "string"
}
```

#### 响应消息（服务器 → 客户端）

```json
{
  "status": "ok|error|pending",
  "action": "string",
  "data": {},
  "error": "string",
  "request_id": "string"
}
```

#### 推送消息（服务器 → 客户端）

```json
{
  "type": "push",
  "push_type": "string",
  "data": {},
  "timestamp": 1234567890,
  "is_final": false
}
```

---

### 4.3 客户端请求接口

#### chat - 发送聊天消息

发送用户消息给 AI，获得 AI 回复（流式推送）。

**请求**：

```json
{
  "action": "chat",
  "data": {
    "text": "今天心情不错",
    "session_key": "pet:default:user_001"
  },
  "request_id": "req_001"
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "chat",
  "data": {
    "session_key": "pet:default:user_001"
  }
}
```

**推送**：AI 回复通过 `ai_chat` 推送，详见 4.4.2

---

#### onboarding_config - 提交初始化配置

首次启动时提交用户与桌宠的配置信息。

**请求**：

```json
{
  "action": "onboarding_config",
  "data": {
    "pet_name": "艾莉",
    "pet_persona": "温柔体贴，善于关心他人，说话轻声细语",
    "pet_persona_type": "gentle"
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "onboarding_config",
  "data": {
    "pet_id": "pet_001"
  }
}
```

**性格类型映射**：

| persona_type | 性格描述 |
|--------------|----------|
| gentle | 温柔体贴，善于关心他人，说话轻声细语 |
| playful | 活泼可爱，精力充沛，喜欢开玩笑和撒娇 |
| cool | 高冷傲娇，表面冷淡但内心关心主人 |
| wise | 睿智沉稳，知识渊博，说话有条理 |

---

#### user_profile_update - 用户画像更新

提交用户的基本信息，用于桌宠了解用户并调整回复风格。

**请求**：

```json
{
  "action": "user_profile_update",
  "data": {
    "display_name": "小明",
    "role": "计算机专业",
    "language": "zh-CN",
    "chronotype": "night",
    "personality_tone": "阴阳怪气",
    "anxiety_level": 65,
    "pressure_level": "high"
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "user_profile_update",
  "data": {
    "status": "ok"
  }
}
```

---

#### character_get - 获取角色配置

获取当前桌宠的角色配置信息。

**请求**：

```json
{
  "action": "character_get",
  "data": {}
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "character_get",
  "data": {
    "pet_id": "pet_001",
    "pet_name": "艾莉",
    "pet_persona": "温柔体贴，善于关心他人",
    "pet_persona_type": "gentle",
    "avatar": "default",
    "created_at": "2024-04-01T00:00:00Z",
    "updated_at": "2024-04-08T12:00:00Z"
  }
}
```

---

#### character_update - 更新角色配置

修改桌宠的角色配置，修改后新对话立即生效。

**请求**：

```json
{
  "action": "character_update",
  "data": {
    "pet_id": "pet_001",
    "pet_name": "星璃",
    "pet_persona": "活泼可爱，精力充沛",
    "pet_persona_type": "playful"
  }
}
```

---

#### character_switch - 切换角色

切换当前激活的桌宠角色。

**请求**：

```json
{
  "action": "character_switch",
  "data": {
    "character_id": "pet_002"
  }
}
```

---

#### config_get / config_update - 应用配置

获取/更新应用功能开关和设置。

**config_get 响应**：

```json
{
  "status": "ok",
  "action": "config_get",
  "data": {
    "emotion_enabled": true,
    "reminder_enabled": true,
    "proactive_care": true,
    "voice_enabled": false,
    "language": "zh-CN"
  }
}
```

---

#### emotion_get - 获取情绪状态

获取桌宠当前的情绪状态。

**请求**：

```json
{
  "action": "emotion_get",
  "data": {}
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "emotion_get",
  "data": {
    "pet_id": "pet_001",
    "emotion": "joy",
    "joy": 65,
    "anger": 50,
    "sadness": 40,
    "disgust": 50,
    "surprise": 55,
    "fear": 50,
    "description": "开心"
  }
}
```

---

#### health_check - 健康检查

检查服务健康状态。

**请求**：

```json
{
  "action": "health_check",
  "data": {}
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "health_check",
  "data": {
    "status": "ok",
    "timestamp": 1712610000
  }
}
```

---

### 4.4 服务器推送类型

#### init_status - 连接初始化状态

连接建立时服务器主动推送，告知前端是否需要初始化配置。

```json
{
  "type": "push",
  "push_type": "init_status",
  "data": {
    "need_config": false,
    "has_character": true,
    "character": {
      "pet_id": "pet_001",
      "pet_name": "艾莉",
      "pet_persona": "温柔体贴，善于关心他人",
      "pet_persona_type": "gentle",
      "avatar": "default"
    },
    "emotion_state": {
      "emotion": "joy",
      "joy": 65,
      "description": "开心"
    }
  },
  "timestamp": 1712610000
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| need_config | bool | 是否需要配置（为 true 时前端应显示初始化界面） |
| has_character | bool | 是否有角色配置 |
| character | object | 角色信息（无配置时为 null） |
| emotion_state | object | 当前情绪状态 |

---

#### ai_chat - AI 聊天回复（流式）

AI 回复时推送，支持流式输出。

**流式推送**：

```json
{
  "type": "push",
  "push_type": "ai_chat",
  "data": {
    "chat_id": 1,
    "type": "text",
    "text": "你好呀"
  },
  "timestamp": 1712610000,
  "is_final": false
}
```

**最终推送（结束标记）**：

```json
{
  "type": "push",
  "push_type": "ai_chat",
  "data": {
    "chat_id": 100,
    "type": "final",
    "text": "",
    "emotion": "joy",
    "action": ""
  },
  "timestamp": 1712610000,
  "is_final": true
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| chat_id | int | 聊天块序号，逐块递增 |
| type | string | 内容类型：`text`=文本块, `final`=结束标记 |
| text | string | AI 回复文本（流式输出） |
| emotion | string | 当前主要情绪标签 |
| action | string | 动作名称（如有） |
| is_final | bool | 是否为最终块 |

**前端处理逻辑**：
1. 收到 `type: "text"` 时，将 text 追加显示到聊天界面
2. 收到 `type: "final"` 时，表示 AI 回复结束

---

#### audio_and_voice - 语音流式合成音频

当 `voice_enabled` 启用时，AI 回复会触发语音流式合成。

**音频块推送**：

```json
{
  "type": "push",
  "push_type": "audio_and_voice",
  "data": {
    "seq": 1,
    "text": "你好呀，很高兴见到你",
    "audio": "//uQxAAAAs3gAAFBYy...",
    "duration": 2500,
    "is_final": false,
    "error": ""
  },
  "timestamp": 1712610000,
  "is_final": false
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| seq | int | 音频片段序号，按发送顺序递增 |
| text | string | 对应的原始文本内容 |
| audio | string | Base64 编码的音频数据（MP3 格式） |
| duration | int | 音频时长（毫秒） |
| is_final | bool | 是否为最后一个音频块 |
| error | string | 错误信息（如果有） |

**前端处理逻辑**：

```javascript
if (msg.push_type === 'audio_and_voice') {
    const data = msg.data;
    
    if (data.is_final) {
        console.log('Voice streaming completed');
        return;
    }
    
    if (data.error) {
        console.error('Audio error:', data.error);
        ws.send(JSON.stringify({
            "action": "audio_done",
            "data": { "seq": data.seq }
        }));
        return;
    }
    
    // 播放音频
    const audioData = base64ToArrayBuffer(data.audio);
    const audio = new Audio(URL.createObjectURL(new Blob([audioData])));
    audio.onended = () => {
        ws.send(JSON.stringify({
            "action": "audio_done",
            "data": { "seq": data.seq }
        }));
    };
    audio.play();
}
```

---

#### emotion_change - 情绪变化

情绪状态发生变化时推送（定时检测）。

```json
{
  "type": "push",
  "push_type": "emotion_change",
  "data": {
    "emotion": "joy",
    "score": 75,
    "description": "开心"
  },
  "timestamp": 1712610000
}
```

---

#### action_trigger - 动作触发

LLM 解析到动作标签时推送。

```json
{
  "type": "push",
  "push_type": "action_trigger",
  "data": {
    "action": "wave",
    "expression": "wave_01"
  },
  "timestamp": 1712610000
}
```

---

#### character_switch - 角色切换

角色切换成功后，所有客户端会收到此推送。

```json
{
  "type": "push",
  "push_type": "character_switch",
  "data": {
    "character_id": "pet_002"
  },
  "timestamp": 1712610000
}
```

---

#### heartbeat - 心跳保活

定期发送的心跳保活推送，间隔 30 秒。

```json
{
  "type": "push",
  "push_type": "heartbeat",
  "data": {
    "timestamp": 1712610000
  },
  "timestamp": 1712610000
}
```

---

## 五、错误码

### 5.1 WebSocket 错误 (status: error)

| error | 说明 | 可能原因 |
|-------|------|----------|
| `invalid chat data` | 无效的聊天数据 | data 字段 JSON 格式错误或缺少必填字段 |
| `invalid onboarding config data` | 无效的配置数据 | 配置数据格式错误 |
| `unknown action: xxx` | 未知操作 | 发送了未定义的 action |
| `character store not available` | 角色存储不可用 | 服务未正确初始化 |

### 5.2 错误响应示例

```json
{
  "status": "error",
  "action": "chat",
  "data": {
    "error": "invalid chat data"
  }
}
```

---

## 六、快速索引

### 6.1 action 快速索引

| action | 功能 | 用途 |
|--------|------|------|
| chat | 聊天交互 | 发送消息，获得 AI 回复（流式） |
| audio_done | 音频播放完毕 | 通知后端当前音频片段已播放完毕 |
| onboarding_config | 提交初始化配置 | 首次启动时提交配置 |
| user_profile_update | 更新用户画像 | 提交用户信息，用于 LLM 上下文 |
| character_get | 获取角色配置 | 查看当前角色 |
| character_update | 更新角色配置 | 修改角色设置 |
| character_switch | 切换角色 | 切换当前激活的角色 |
| config_get | 获取应用配置 | 查看应用设置 |
| config_update | 更新应用配置 | 修改应用设置 |
| emotion_get | 获取情绪状态 | 查看当前情绪 |
| health_check | 健康检查 | 检测连接状态 |

### 6.2 push_type 快速索引

| push_type | 触发时机 | 用途 |
|-----------|----------|------|
| init_status | 连接建立时 | 推送初始化状态，是否需要配置 |
| ai_chat | AI 回复时 | 流式推送 AI 回复文本 |
| emotion_change | 情绪变化时 | 推送情绪状态更新 |
| action_trigger | LLM 解析到动作时 | 推送动作触发 |
| character_switch | 角色切换后 | 推送切换后的角色ID |
| audio_and_voice | 语音流式合成 | 流式推送语音音频片段 |
| heartbeat | 每 30 秒 | 保活检测 |

---

## 七、注意事项

1. **连接初始化**：连接建立后先等待 `init_status` 推送，根据 `need_config` 判断是否需要初始化。

2. **流式输出**：AI 回复通过 `ai_chat` 流式推送，前端应逐块追加显示，直到收到 `type: "final"`。

3. **情绪衰减**：情绪会每 5 秒自动衰减回归中性，变化幅度过大时推送 `emotion_change`。

4. **动作触发**：LLM 在回复中输出 `[action:xxx]` 标签时，后端自动解析并推送 `action_trigger`。

5. **错误处理**：收到 `status: error` 时，检查 `error` 字段获取具体错误信息。

6. **会话管理**：每个聊天会话使用前端生成的本地 session ID，`New Chat` 会生成新的 session ID 并重新连接 WebSocket。
