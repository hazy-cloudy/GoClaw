# Pet Channel API 接口文档

> 版本：v1.0  
> 日期：2026-04-08  
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
  "timestamp": 1234567890       
}
```

---

## 三、功能模块

---

### 3.1 聊天交互

#### 3.1.1 chat - 发送聊天消息

发送用户消息给 AI，获得 AI 回复。

**请求**：

```json
{
  "action": "chat",
  "data": {
    "text": "今天心情不错",
    "session_key": "pet:default:user_001"
  }
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

**推送 (ai_chat)**：

```json
{
  "type": "push",
  "push_type": "ai_chat",
  "data": {
    "text": "主人今天心情好，我也很开心呢~",
    "emotion": "happy",
    "action": "Idle"
  },
  "timestamp": 1712610000
}
```

**用途**：实现用户与桌宠的自然语言对话交互。

### 3.1.2 voice - 发送语音消息

```json
{
  "action": "voice",
  "data": {
    "text": "今天心情不错",
    "session_key": "pet:default:user_001"
  }
}
```

**响应**：

```json

{
  "status": "ok",
  "action": "voice",
  "data": {
    "session_key": "pet:default:user_001"
  }
}
```
**推送 (ai_voice)**：

```json

{
  "type": "push",
  "push_type": "ai_voice",
  "data": {
    "text": "主人今天心情好，我也很开心呢~",
    "emotion": "happy",
    "action": "Idle",
    "voice_url": "https://example.com/voice.mp3"
  },
  "timestamp": 1712610000
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| text | string | 用户输入的文本内容 |
| session_key | string | 会话标识符 |

**ai_chat 推送字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| text | string | AI 回复文本 |
| emotion | string | 情绪标签 (neutral/happy/sad/angry/worried/love) |
| action | string | Live2D 动作名称 (Idle/TapBody/Shake...) |

**ai_voice 推送字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| text | string | AI 回复文本 |
| emotion | string | 情绪标签 (neutral/happy/sad/angry/worried/love) |
| action | string | Live2D 动作名称 (Idle/TapBody/Shake...) |
| voice_url | string | 语音URL |
---

### 3.2 引导配置

#### 3.2.1 onboarding_start - 查询是否需要初始化

前端启动时发送此请求查询是否需要初始化。

**请求**：

```json
{
  "action": "onboarding_start",
  "data": {}
}
```

**响应（需要初始化）**：

```json
{
  "status": "ok",
  "action": "onboarding_start",
  "data": {}
}
```

**响应（已有配置）**：

```json
{
  "status": "error",
  "action": "onboarding_start",
  "data": {
    "error": "onboarding already completed"
  }
}
```

**用途**：前端启动时查询是否需要初始化引导流程。

---

#### 3.2.2 onboarding_config - 提交配置

前端收集完用户与桌宠的配置信息后发送给后端，后端保存并推送欢迎消息。

**请求**：

```json
{
  "action": "onboarding_config",
  "data": {
    "user_name": "user_001",
    "user_sex": "male",
    "pet_name": "桌宠",
    "pet_sex": "male",
    "pet_persona_context": "温柔体贴，善于关心他人，说话轻声细语",
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

**推送 (onboarding_config)**：

```json
{
  "type": "push",
  "push_type": "onboarding_config",
  "data": {
    "text": "你好，我是桌宠，你可以和我聊天吗？"
  },
  "timestamp": 1712610000
}
```

**用途**：前端收集完用户与桌宠的配置信息后发送，后端保存并推送欢迎消息。

**字段说明**：

| 字段           | 类型 | 必填 | 说明 |
|--------------|------|------|------|
| user_name    | string | 是 | 用户名 |
| user_sex     | string | 是 | 用户性别 |
| pet_name     | string | 是 | 桌宠名称 |
| pet_sex      | string | 是 | 桌宠性别 |
| pet_persona_context | string | 否 | 详细性格描述（如果选择自定义） |
| pet_persona_type | string | 否 | 性格类型ID，与 pet_persona_context 二选一 |

**性格类型映射**：

| persona_type | 性格描述 |
|--------------|----------|
| gentle | 温柔体贴，善于关心他人，说话轻声细语，会用温暖的话语安慰主人 |
| playful | 活泼可爱，精力充沛，喜欢开玩笑和撒娇，说话充满活力 |
| cool | 高冷傲娇，表面冷淡但内心关心主人，说话简短直接，偶尔别扭地表达关心 |
| wise | 睿智沉稳，知识渊博，说话有条理，会给主人提供理性的建议 |

---

### 3.3 角色管理

#### 3.3.1 character_get - 获取角色配置

获取当前桌宠的角色配置信息。

**请求**：

```json
{
  "action": "character_get",
  "data": {
    "pet_id": "pet_001"
  }
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
    "pet_persona": "温柔体贴，善于关心他人，说话轻声细语",
    "pet_persona_type": "gentle",
    "pet_mbti": {
      "ie": 60,
      "sn": 40,
      "tf": 30,
      "jp": 50
    },
    "created_at": "2024-04-01T00:00:00Z",
    "updated_at": "2024-04-08T12:00:00Z"
  }
}
```

**用途**：查看当前角色配置，用于前端展示和编辑。

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| pet_id | string | 桌宠ID |
| pet_name | string | 桌宠名称 |
| pet_persona | string | 详细性格描述 |
| pet_persona_type | string | 性格类型ID |
| pet_mbti | object | MBTI 性格配置 |
| pet_mbti.ie | int | 内向/外向 (0-100, 50中性, >50偏外向) |
| pet_mbti.sn | int | 实感/直觉 (0-100, 50中性, >50偏实感) |
| pet_mbti.tf | int | 理性/感性 (0-100, 50中性, >50偏理性) |
| pet_mbti.jp | int | 判断/感知 (0-100, 50中性, >50偏判断) |
| created_at | string | 创建时间 (RFC3339) |
| updated_at | string | 更新时间 (RFC3339) |

---

#### 3.3.2 character_update - 更新角色配置

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
    "pet_mbti": {
      "ie": 70,
      "sn": 30,
      "tf": 50,
      "jp": 45
    },
    "created_at": "2024-04-01T00:00:00Z",
    "updated_at": "2024-04-08T12:30:00Z"
  }
}
```

**用途**：修改角色配置，影响后续 AI 对话的人格表现。

**注意**：修改后新的对话会立即使用新配置，无需重启服务。

---

### 3.4 应用配置

#### 3.4.1 config_get - 获取应用配置

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
    "reminder_enabled": true,
    "proactive_care": true,
    "proactive_interval_minutes": 30,
    "voice_enabled": false,
    "language": "zh-CN"
  }
}
```

**用途**：获取应用设置，用于前端展示和修改。

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| reminder_enabled | bool | 是否启用提醒功能 |
| proactive_care | bool | 是否启用主动关怀 |
| proactive_interval_minutes | int | 主动关怀间隔（分钟） |
| voice_enabled | bool | 是否启用语音 |
| language | string | 语言设置 |

---

#### 3.4.2 config_update - 更新应用配置

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

**用途**：修改应用设置，实时生效。

---

### 3.5 情绪状态

#### 3.5.1 emotion_get - 获取当前情绪

获取桌宠当前的情绪状态。

**请求**：

```json
{
  "action": "emotion_get",
  "data": {
    "pet_id": "pet_001"
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "emotion_get",
  "data": {
    "pet_id": "pet_001",
    "emotion": "happy",
    "score": 0.85,
    "expression": "F01",
    "motion": "Idle",
    "description": "开心"
  }
}
```

**用途**：前端获取情绪状态，用于动画和表情展示。


**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| pet_id | string | 桌宠ID |
| emotion | string | 情绪标签 (neutral/happy/sad/angry/worried/love/surprised) |
| score | float | 情绪强度 (0.0-1.0) |
| expression | string | Live2D 表情文件 ID |
| motion | string | Live2D 动作名称 |
| description | string | 情绪中文描述 |

---

### 3.6 事件交互

#### 3.6.1 event_action - 触发事件动作

前端触发的事件动作，如点击、敲击等。

**请求**：

```json
{
  "action": "event_action",
  "data": {
    "event": "tap",
    "params": {
      "x": 100,
      "y": 200
    }
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "event_action",
  "data": {
    "handled": true
  }
}
```

**用途**：用户与桌宠交互时，触发相应事件，获得 AI 响应。

**event 类型说明**：

| event | 说明 | params |
|-------|------|--------|
| tap | 点击桌宠 | x, y 坐标 |
| double_tap | 双击桌宠 | x, y 坐标 |
| knock | 敲击屏幕 | - |
| long_press | 长按 | x, y 坐标 |
| idle | 空闲检测触发 | idle_seconds |

---

### 3.7 系统

#### 3.7.1 health_check - 健康检查

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
    "timestamp": "1712610000"
  }
}
```

**用途**：检测 WebSocket 连接是否正常，服务是否可用。

---

### 3.8 记忆管理

#### 3.8.1 memory_get - 获取记忆记录

获取桌宠的最近记忆记录。

**请求**：

```json
{
  "action": "memory_get",
  "data": {
    "pet_id": "pet_001",
    "limit": 10
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "memory_get",
  "data": [
    {
      "id": "mem_123",
      "pet_id": "pet_001",
      "user_id": "user_001",
      "type": "text",
      "text": "主人今天心情好，我也很开心呢~",
      "timestamp": 1712610000
    },
    // 其他记忆记录...
  ]
}
````

**用途**：前端获取桌宠的最近记忆记录，用于展示和分析。

---

#### 3.8.2 memory_type_add - 添加记忆类型

添加新的记忆类型，如语音记忆。

**请求**：

```json
{
  "action": "memory_type_add",
  "data": {
    "pet_id": "pet_001",
    "type": "voice"
  }
}
```

**响应**：

```json
{
  "status": "ok",
  "action": "memory_type_add",
  "data": {}
}
```

**用途**：添加新的记忆类型，如语音记忆，用于记录和管理桌宠的语音记忆。

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 记忆ID |
| pet_id | string | 桌宠ID |
| user_id | string | 用户ID |
| type | string | 记忆类型 (text/voice) |
| text | string | 记忆文本 |
| timestamp | int | 记忆时间戳 |




---

#### 3.8.3 memory_old_get - 获取旧记忆摘要

获取桌宠的旧记忆摘要，用于展示和分析。

**请求**：

```json
{
  "action": "memory_old_get",
  "data": {
    "pet_id": "pet_001",
    "limit": 10
  }
}
```


```json
{
  "status": "ok",
  "action": "memory_old_get",
  "data": [
    {
      "pet_id": "pet_001",
      "summary": "主人今天心情好，我也很开心呢~"
    },
    // 其他旧记忆摘要...
  ]
}
```

**用途**：前端获取桌宠的旧记忆摘要，用于展示和分析。

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| pet_id | string | 桌宠ID |
| summary | string | 旧记忆摘要文本 |

---

### 3.9 动态管理

#### 3.9.1 dynamic_get - 获取桌宠最近发布的动态

获取桌宠的最近发布的动态，用于展示和分析。

```json
{
  "action": "dynamic_get",
  "data": {
    "pet_id": "pet_001",
    "limit": 10
  }
}
````````

**响应**：

```json
{
  "status": "ok",
  "action": "dynamic_get",
  "data": [
    {
      "id": "dyn_123",
      "pet_id": "pet_001",
      "user_id": "user_001",
      "text": "主人今天心情好，我也很开心呢~",
      "picture": "https://example.com/picture.jpg",
      "timestamp": 1712610000
    },
    // 其他动态...
  ]
}
```

**用途**：前端获取桌宠的最近发布的动态，用于展示和分析。

---

#### 3.9.2 dynamic_add - 桌宠动态更新 (推送)

桌宠通过内部定时实现动态更新，由服务端主动推送。

**推送类型**：dynamic_add

```json
{
  "type": "push",
  "push_type": "dynamic_add",
  "data": {
    "id": "dyn_123",
    "pet_id": "pet_001",
    "user_id": "user_001",
    "text": "主人今天心情好，我也很开心呢~",
    "picture": "https://example.com/picture.jpg",
    "timestamp": 1712610000
  }
}
```

推送桌宠的动态更新，包含动态 ID、桌宠 ID、用户 ID、动态文本和时间戳。

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 动态ID |
| pet_id | string | 桌宠ID |
| user_id | string | 用户ID |
| text | string | 动态文本 |
| picture | string | 动态图片URL |
| timestamp | int | 动态时间戳 |

---

## 四、推送类型

### 4.1 ai_chat - AI 回复

当 AI 生成回复时推送。

```json 
{
  "type": "push",
  "push_type": "ai_chat",
  "data": {
    "text": "回复文本内容",
    "emotion": "happy",
    "action": "Idle"
  }
}
```

### 4.2 emotion_change - 情绪变化

当情绪状态发生变化时推送。

```json
{
  "type": "push",
  "push_type": "emotion_change",
  "data": {
    "emotion": "happy",
    "score": 0.85
  }
}
```

### 4.3 reminder_trigger - 提醒触发

当定时提醒到期时推送。

```json
{
  "type": "push",
  "push_type": "reminder_trigger",
  "data": {
    "reminder_id": "rem_123",
    "content": "喝水时间到啦，主人快补水！",
    "timestamp": 1712610000
  }
}
```

### 4.4 heartbeat - 心跳保活

定期发送的心跳保活推送，间隔 30 秒。

```json
{
  "type": "push",
  "push_type": "heartbeat",
  "data": {
    "timestamp": 1712610000
  }
}
```

### 4.5 system_status - 系统状态

系统状态变更时推送。

```json
{
  "type": "push",
  "push_type": "system_status",
  "data": {
    "status": "ready",
    "message": "服务就绪"
  }
}
```

---

## 五、错误码

### 5.1 HTTP 错误

| 错误 | 说明 |
|------|------|
| 400 Bad Request | 请求格式错误 |
| 401 Unauthorized | 未授权 |
| 403 Forbidden | 禁止访问 |
| 404 Not Found | 资源不存在 |
| 500 Internal Server Error | 服务器内部错误 |
| 503 Service Unavailable | 服务不可用 |

### 5.2 WebSocket 错误 (status: error)

| error | 说明 | 可能原因 |
|-------|------|----------|
| `INVALID_ACTION` | 无效的操作类型 | action 字段值不正确 |
| `INVALID_DATA` | 无效的数据格式 | data 字段 JSON 格式错误或缺少必填字段 |
| `SESSION_NOT_FOUND` | 会话不存在 | sessionId 不存在或已过期 |
| `CONFIG_ERROR` | 配置错误 | 配置项格式或值不正确 |
| `INTERNAL_ERROR` | 内部错误 | 服务器内部处理出错 |
| `unknown action: xxx` | 未知操作 | 发送了未定义的 action |

### 5.3 错误响应示例

```json
{
  "status": "error",
  "action": "chat",
  "data": {
    "code": "INVALID_DATA",
    "error": "invalid chat data"
  }
}
```

---

## 六、配置文件格式

角色配置存储在 `workspace/characters/default/config.md`，格式如下：

```markdown
---
name: 艾莉
persona: 温柔体贴，善于关心他人，说话轻声细语
persona_type: gentle
mbti:
  ie: 60
  sn: 40
  tf: 30
  jp: 50
avatar: default
created_at: 2024-04-01T00:00:00Z
updated_at: 2024-04-08T12:00:00Z
---
# Soul
## Personality
温柔体贴，善于关心他人，说话轻声细语，会用温暖的话语安慰主人
```

---

## 七、客户端示例

### 7.1 JavaScript 客户端示例

```javascript
// 建立 WebSocket 连接
const ws = new WebSocket('ws://localhost:8080/ws?session=user_001');

// 连接成功
ws.onopen = function() {
  console.log('Connected to Pet Channel');
  
  // 发送聊天消息
  const chatRequest = {
    action: 'chat',
    data: {
      text: '今天心情不错',
      session_key: 'pet:default:user_001'
    },
    request_id: 'req_001'
  };
  ws.send(JSON.stringify(chatRequest));
  
  // 发送健康检查
  setInterval(() => {
    const healthRequest = {
      action: 'health_check',
      data: {}
    };
    ws.send(JSON.stringify(healthRequest));
  }, 30000); // 每30秒发送一次
};

// 接收消息
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  // 处理响应消息
  if (message.status) {
    console.log('Received response:', message);
    
    // 处理不同类型的响应
    switch (message.action) {
      case 'chat':
        console.log('Chat response received:', message.data);
        break;
      case 'health_check':
        console.log('Health check response:', message.data);
        break;
      // 其他响应类型...
    }
  }
  
  // 处理推送消息
  else if (message.type === 'push') {
    console.log('Received push:', message);
    
    // 处理不同类型的推送
    switch (message.push_type) {
      case 'ai_chat':
        console.log('AI response:', message.data.text);
        console.log('Emotion:', message.data.emotion);
        console.log('Action:', message.data.action);
        // 更新UI显示AI回复和表情动作
        break;
      case 'emotion_change':
        console.log('Emotion changed:', message.data.emotion);
        // 更新UI显示情绪变化
        break;
      case 'heartbeat':
        console.log('Heartbeat received:', message.data.timestamp);
        break;
      // 其他推送类型...
    }
  }
};

// 连接关闭
ws.onclose = function() {
  console.log('Disconnected from Pet Channel');
  // 实现重连逻辑
  setTimeout(() => {
    console.log('Attempting to reconnect...');
    // 重新建立连接
  }, 5000);
};

// 连接错误
ws.onerror = function(error) {
  console.error('WebSocket error:', error);
};
```

### 7.2 Python 客户端示例

```python
import websocket
import json
import time

def on_message(ws, message):
    data = json.loads(message)
    
    # 处理响应消息
    if 'status' in data:
        print('Received response:', data)
        
        # 处理不同类型的响应
        if data.get('action') == 'chat':
            print('Chat response received:', data.get('data'))
        elif data.get('action') == 'health_check':
            print('Health check response:', data.get('data'))
    
    # 处理推送消息
    elif data.get('type') == 'push':
        print('Received push:', data)
        
        # 处理不同类型的推送
        if data.get('push_type') == 'ai_chat':
            print('AI response:', data['data'].get('text'))
            print('Emotion:', data['data'].get('emotion'))
            print('Action:', data['data'].get('action'))
        elif data.get('push_type') == 'emotion_change':
            print('Emotion changed:', data['data'].get('emotion'))
        elif data.get('push_type') == 'heartbeat':
            print('Heartbeat received:', data['data'].get('timestamp'))

def on_error(ws, error):
    print('WebSocket error:', error)

def on_close(ws, close_status_code, close_msg):
    print('Disconnected from Pet Channel')
    # 实现重连逻辑
    time.sleep(5)
    print('Attempting to reconnect...')
    # 重新建立连接

def on_open(ws):
    print('Connected to Pet Channel')
    
    # 发送聊天消息
    chat_request = {
        'action': 'chat',
        'data': {
            'text': '今天心情不错',
            'session_key': 'pet:default:user_001'
        },
        'request_id': 'req_001'
    }
    ws.send(json.dumps(chat_request))
    
    # 定期发送健康检查
    def send_health_check():
        while True:
            time.sleep(30)
            health_request = {
                'action': 'health_check',
                'data': {}
            }
            ws.send(json.dumps(health_request))
    
    import threading
    threading.Thread(target=send_health_check, daemon=True).start()

if __name__ == '__main__':
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp(
        'ws://localhost:8080/ws?session=user_001',
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close
    )
    ws.run_forever()
```

### 7.3 客户端实现要点

1. **连接管理**：
   - 建立 WebSocket 连接时指定 session 参数
   - 实现重连机制，确保连接稳定性
   - 定期发送健康检查，维持连接活跃

2. **消息处理**：
   - 区分响应消息（带 status 字段）和推送消息（带 type: "push" 字段）
   - 根据 action 或 push_type 处理不同类型的消息
   - 实现错误处理，优雅处理连接错误和消息解析错误

3. **UI 集成**：
   - 收到 ai_chat 推送时，更新聊天界面显示 AI 回复
   - 收到 emotion_change 推送时，更新桌宠表情和动作
   - 收到 reminder_trigger 推送时，显示提醒弹窗

4. **性能优化**：
   - 使用 JSON 解析库高效处理消息
   - 实现消息队列，避免 UI 阻塞
   - 合理设置心跳间隔，平衡实时性和网络开销

## 八、快速参考

### 8.1 action 快速索引

| action | 功能 | 用途 |
|--------|------|------|
| chat | 聊天交互 | 发送消息，获得 AI 回复 |
| voice | 语音交互 | 发送语音消息，获得 AI 回复 |
| onboarding_start | 查询初始化状态 | 查询是否需要初始化引导 |
| onboarding_config | 提交引导配置 | 首次启动时提交用户与桌宠配置 |
| character_get | 获取角色配置 | 查看当前角色 |
| character_update | 更新角色配置 | 修改角色设置 |
| config_get | 获取应用配置 | 查看应用设置 |
| config_update | 更新应用配置 | 修改应用设置 |
| emotion_get | 获取情绪状态 | 查看当前情绪 |
| event_action | 触发事件 | 用户交互事件 |
| health_check | 健康检查 | 检测连接状态 |
| memory_get | 获取记忆记录 | 获取最近记忆 |
| memory_type_add | 添加记忆类型 | 添加记忆类型 |
| memory_old_get | 获取旧记忆摘要 | 获取旧记忆摘要 |
| dynamic_get | 获取动态 | 获取桌宠动态 |

### 8.2 push_type 快速索引

| push_type | 触发时机 | 用途 |
|-----------|----------|------|
| ai_chat | AI 回复生成完成 | 显示 AI 回复和表情动作 |
| ai_voice | AI 语音回复生成完成 | 显示 AI 语音回复 |
| emotion_change | 情绪状态变化 | 更新界面情绪显示 |
| reminder_trigger | 定时提醒到期 | 显示提醒弹窗 |
| onboarding_config | 引导配置完成 | 推送欢迎消息 |
| heartbeat | 每 30 秒 | 保活检测 |
| system_status | 系统状态变化 | 显示系统状态 |
| dynamic_add | 桌宠动态更新 | 推送新动态 |

---

## 九、注意事项

1. **连接保活**：建议客户端每 30 秒发送一次 `health_check` 请求维持连接。

2. **配置生效**：修改 `character_update` 或 `config_update` 后，新对话会立即使用新配置。

3. **错误处理**：收到 `status: error` 时，检查 `error` 字段获取具体错误信息。

4. **重连机制**：WebSocket 断开后，建议客户端自动重连。

5. **session 隔离**：不同的 `sessionId` 会话数据隔离，互不影响。