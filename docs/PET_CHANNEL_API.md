# Pet Channel API 接口文档

> 版本：v2.1  
> 日期：2026-04-14  
> 协议：WebSocket + JSON

---

## 一、连接方式

### 1.1 WebSocket 连接地址

```
ws://{host}:{port}/ws?session={sessionId}
```

**参数说明**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 是 | 服务器地址，默认 `0.0.0.0` |
| port | int | 是 | 服务器端口，默认 `8080` |
| sessionId | string | 否 | 会话ID，默认为 `default` |

**连接示例**：
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?session=user_001');
```

**重要**：连接建立后，服务器会主动推送 `init_status`，告知前端是否需要初始化配置。

---

## 二、消息格式

### 2.1 请求消息 (Client → Server)

```json
{
  "action": "string",
  "data": {},
  "request_id": "string"
}
```

### 2.2 响应消息 (Server → Client)

```json
{
  "status": "ok|error|pending",
  "action": "string",
  "data": {},
  "error": "string",
  "request_id": "string"
}
```

### 2.3 推送消息 (Server → Client)

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

## 三、推送类型

### 3.1 init_status - 连接初始化状态

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
    "mbti": {
      "ie": 60,
      "sn": 40,
      "tf": 30,
      "jp": 50
    },
    "emotion_state": {
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
| character.pet_id | string | 桌宠ID |
| character.pet_name | string | 桌宠名称 |
| character.pet_persona | string | 性格描述 |
| character.pet_persona_type | string | 性格类型 |
| character.avatar | string | 头像/模型ID |
| mbti | object | MBTI 配置 |
| mbti.ie | int | 内向/外向 (0-100, 50中性, <50偏内向, >50偏外向) |
| mbti.sn | int | 实感/直觉 (0-100, 50中性, <50偏实感, >50偏直觉) |
| mbti.tf | int | 理性/感性 (0-100, 50中性, <50偏理性, >50偏感性) |
| mbti.jp | int | 判断/感知 (0-100, 50中性, <50偏判断, >50偏感知) |
| emotion_state | object | 当前情绪状态 |
| emotion_state.emotion | string | 主要情绪标签 (neutral/joy/anger/sadness/disgust/surprise/fear) |
| emotion_state.joy | int | 快乐值 0-100 |
| emotion_state.anger | int | 愤怒值 0-100 |
| emotion_state.sadness | int | 悲伤值 0-100 |
| emotion_state.disgust | int | 厌恶值 0-100 |
| emotion_state.surprise | int | 惊讶值 0-100 |
| emotion_state.fear | int | 恐惧值 0-100 |
| emotion_state.description | string | 情绪中文描述 |

---

### 3.2 ai_chat - AI 聊天回复（流式）

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
3. 情绪和动作标签在流式输出时已由后端处理，不在 text 中显示

---

### 3.3 emotion_change - 情绪变化

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

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| emotion | string | 情绪标签 (neutral/joy/anger/sadness/disgust/surprise/fear) |
| score | int | 情绪强度 0-100 |
| description | string | 情绪中文描述 |

---

### 3.4 action_trigger - 动作触发

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

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| action | string | 动作名称 |
| expression | string | 表情标识 |

---

### 3.5 heartbeat - 心跳保活

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

### 3.6 character_switch - 角色切换

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

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| character_id | string | 切换后的角色ID |

---

### 3.7 ai_audio - 语音合成音频

当 `voice_enabled` 启用时，AI 回复会同时触发语音合成，并将音频数据推送客户端播放。

```json
{
  "type": "push",
  "push_type": "ai_audio",
  "data": {
    "chat_id": 1,
    "type": "audio",
    "text": "//uQxAAAAs3gAAFBYy...",
    "is_final": false
  },
  "timestamp": 1712610000,
  "is_final": false
}
```

**最终推送（结束标记）**：

```json
{
  "type": "push",
  "push_type": "ai_audio",
  "data": {
    "chat_id": 1,
    "type": "audio",
    "text": "",
    "is_final": true
  },
  "timestamp": 1712610000,
  "is_final": true
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| chat_id | int | 聊天会话ID |
| type | string | 内容类型：`audio`=音频块, `error`=错误信息 |
| text | string | Base64 编码的音频数据（MP3 格式），`is_final=true` 时为空 |
| is_final | bool | 是否为最后一个音频块 |

**音频解码示例**（JavaScript）：

```javascript
// 接收 audio 推送
const audioData = base64ToArrayBuffer(msg.data.text);
const audioBlob = new Blob([audioData], { type: 'audio/mpeg' });
const audioUrl = URL.createObjectURL(audioBlob);
const audio = new Audio(audioUrl);
audio.play();
```

---

## 四、请求接口

### 4.1 chat - 发送聊天消息

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

**推送**：AI 回复通过 `ai_chat` 推送，详见 3.2

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | 是 | 用户输入的文本内容 |
| session_key | string | 是 | 会话标识符 |

---

### 4.2 onboarding_config - 提交初始化配置

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

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pet_name | string | 是 | 桌宠名称 |
| pet_persona | string | 是 | 性格描述 |
| pet_persona_type | string | 否 | 性格类型ID |

**性格类型映射**：

| persona_type | 性格描述 |
|--------------|----------|
| gentle | 温柔体贴，善于关心他人，说话轻声细语 |
| playful | 活泼可爱，精力充沛，喜欢开玩笑和撒娇 |
| cool | 高冷傲娇，表面冷淡但内心关心主人 |
| wise | 睿智沉稳，知识渊博，说话有条理 |

---

### 4.3 character_get - 获取角色配置

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

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| pet_id | string | 桌宠ID |
| pet_name | string | 桌宠名称 |
| pet_persona | string | 性格描述 |
| pet_persona_type | string | 性格类型 |
| avatar | string | 头像/模型ID |
| created_at | string | 创建时间（ISO 8601） |
| updated_at | string | 更新时间（ISO 8601） |

---

### 4.4 character_update - 更新角色配置

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

**响应**：

```json
{
  "status": "ok",
  "action": "character_update",
  "data": {
    "pet_id": "pet_001",
    "pet_name": "星璃",
    "pet_persona": "活泼可爱，精力充沛",
    "pet_persona_type": "playful",
    "avatar": "default",
    "created_at": "2024-04-01T00:00:00Z",
    "updated_at": "2024-04-08T12:30:00Z"
  }
}
```

**注意**：修改后新的对话会立即使用新配置。

---

### 4.5 character_switch - 切换角色

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

**响应**：

```json
{
  "status": "ok",
  "action": "character_switch",
  "data": {
    "character_id": "pet_002"
  }
}
```

**推送**：角色切换成功后，所有客户端会收到 `character_switch` 推送，详见 3.6

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| character_id | string | 是 | 目标角色ID |

---

### 4.6 config_get - 获取应用配置

获取应用功能开关和设置。

**请求**：

```json
{
  "action": "config_get",
  "data": {}
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "config_get",
  "data": {
    "emotion_enabled": true,
    "reminder_enabled": true,
    "proactive_care": true,
    "proactive_interval_minutes": 30,
    "voice_enabled": false,
    "language": "zh-CN"
  }
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| emotion_enabled | bool | 是否启用情绪表情 |
| reminder_enabled | bool | 是否启用提醒功能 |
| proactive_care | bool | 是否启用主动关怀 |
| proactive_interval_minutes | int | 主动关怀间隔（分钟） |
| voice_enabled | bool | 是否启用语音 |
| language | string | 语言设置 |

---

### 4.7 config_update - 更新应用配置

修改应用功能设置。

**请求**：

```json
{
  "action": "config_update",
  "data": {
    "emotion_enabled": true,
    "proactive_care": false,
    "language": "zh-CN"
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "config_update",
  "data": {
    "emotion_enabled": true,
    "reminder_enabled": true,
    "proactive_care": false,
    "proactive_interval_minutes": 30,
    "voice_enabled": false,
    "language": "zh-CN"
  }
}
```

---

### 4.8 emotion_get - 获取情绪状态

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

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| pet_id | string | 桌宠ID |
| emotion | string | 主要情绪标签 |
| joy | int | 快乐值 0-100 |
| anger | int | 愤怒值 0-100 |
| sadness | int | 悲伤值 0-100 |
| disgust | int | 厌恶值 0-100 |
| surprise | int | 惊讶值 0-100 |
| fear | int | 恐惧值 0-100 |
| description | string | 情绪中文描述 |

---

### 4.9 health_check - 健康检查

检查服务健康状态和连接状态。

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

## 五、错误码

### 5.1 WebSocket 错误 (status: error)

| error | 说明 | 可能原因 |
|-------|------|----------|
| `invalid chat data` | 无效的聊天数据 | data 字段 JSON 格式错误或缺少必填字段 |
| `invalid onboarding config data` | 无效的配置数据 | 配置数据格式错误 |
| `invalid character data` | 无效的角色数据 | 角色数据格式错误 |
| `invalid config data` | 无效的配置数据 | 应用配置数据格式错误 |
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

## 六、客户端示例

### 6.1 JavaScript 客户端示例

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?session=user_001');

ws.onopen = function() {
  console.log('Connected to Pet Channel');
};

ws.onmessage = function(event) {
  const msg = JSON.parse(event.data);

  // 处理 init_status 推送
  if (msg.push_type === 'init_status') {
    console.log('Init status:', msg.data);
    if (msg.data.need_config) {
      showOnboardingUI();
    } else {
      showChatUI();
    }
    return;
  }

  // 处理 AI 聊天流式推送
  if (msg.push_type === 'ai_chat') {
    const data = msg.data;
    if (data.type === 'text') {
      appendToChat(data.text);
    } else if (data.type === 'final') {
      finishChat();
    }
    return;
  }

  // 处理情绪变化推送
  if (msg.push_type === 'emotion_change') {
    updateEmotionDisplay(msg.data.emotion, msg.data.score);
    return;
  }

  // 处理动作触发推送
  if (msg.push_type === 'action_trigger') {
    playAction(msg.data.action);
    return;
  }

  // 处理角色切换推送
  if (msg.push_type === 'character_switch') {
    console.log('Character switched to:', msg.data.character_id);
    reloadCharacter(msg.data.character_id);
    return;
  }

  // 处理语音音频推送
  if (msg.push_type === 'ai_audio') {
    const data = msg.data;
    if (data.type === 'audio' && data.text) {
      const audioData = base64ToArrayBuffer(data.text);
      playAudioChunk(audioData);
    } else if (data.type === 'error') {
      console.error('Audio error:', data.text);
    }
    return;
  }

  // 处理心跳
  if (msg.push_type === 'heartbeat') {
    console.log('Heartbeat:', msg.data.timestamp);
    return;
  }
};

ws.onclose = function() {
  console.log('Disconnected');
  setTimeout(reconnect, 5000);
};
```

### 6.2 Python 客户端示例

```python
import asyncio
import websockets
import json

async def client():
    uri = "ws://localhost:8080/ws?session=user_001"
    async with websockets.connect(uri) as ws:
        while True:
            msg = await ws.recv()
            data = json.loads(msg)

            if data.get('push_type') == 'init_status':
                if data['data']['need_config']:
                    print("需要初始化配置")
                else:
                    print("已初始化，可以聊天")
                    await send_chat(ws, "你好")

            elif data.get('push_type') == 'ai_chat':
                chat_data = data.get('data', {})
                if chat_data.get('type') == 'text':
                    print(chat_data['text'], end='', flush=True)
                elif chat_data.get('type') == 'final':
                    print()

            elif data.get('push_type') == 'emotion_change':
                print(f"\n情绪: {data['data']['emotion']}")

            elif data.get('push_type') == 'action_trigger':
                print(f"\n动作: {data['data']['action']}")

            elif data.get('push_type') == 'character_switch':
                print(f"\n角色切换: {data['data']['character_id']}")

            elif data.get('push_type') == 'ai_audio':
                audio_data = data.get('data', {})
                if audio_data.get('type') == 'audio' and audio_data.get('text'):
                    audio_bytes = base64.b64decode(audio_data['text'])
                    play_audio_chunk(audio_bytes)
                elif audio_data.get('type') == 'error':
                    print(f"\n音频错误: {audio_data['text']}")

async def send_chat(ws, text):
    await ws.send(json.dumps({
        "action": "chat",
        "data": {"text": text, "session_key": "test"}
    }))
```

---

## 七、快速索引

### 7.1 action 快速索引

| action | 功能 | 用途 |
|--------|------|------|
| chat | 聊天交互 | 发送消息，获得 AI 回复（流式） |
| onboarding_config | 提交初始化配置 | 首次启动时提交配置 |
| character_get | 获取角色配置 | 查看当前角色 |
| character_update | 更新角色配置 | 修改角色设置 |
| character_switch | 切换角色 | 切换当前激活的角色 |
| config_get | 获取应用配置 | 查看应用设置 |
| config_update | 更新应用配置 | 修改应用设置 |
| emotion_get | 获取情绪状态 | 查看当前情绪 |
| health_check | 健康检查 | 检测连接状态 |

### 7.2 push_type 快速索引

| push_type | 触发时机 | 用途 |
|-----------|----------|------|
| init_status | 连接建立时 | 推送初始化状态，是否需要配置 |
| ai_chat | AI 回复时 | 流式推送 AI 回复文本 |
| emotion_change | 情绪变化时 | 推送情绪状态更新 |
| action_trigger | LLM 解析到动作时 | 推送动作触发 |
| character_switch | 角色切换后 | 推送切换后的角色ID |
| ai_audio | 语音合成完成 | 推送音频数据用于播放 |
| heartbeat | 每 30 秒 | 保活检测 |

---

## 八、注意事项

1. **连接初始化**：连接建立后先等待 `init_status` 推送，根据 `need_config` 判断是否需要初始化。

2. **流式输出**：AI 回复通过 `ai_chat` 流式推送，前端应逐块追加显示，直到收到 `type: "final"`。

3. **情绪衰减**：情绪会每 5 秒自动衰减回归中性，变化幅度过大时推送 `emotion_change`。

4. **动作触发**：LLM 在回复中输出 `[action:xxx]` 标签时，后端自动解析并推送 `action_trigger`。

5. **错误处理**：收到 `status: error` 时，检查 `error` 字段获取具体错误信息。
