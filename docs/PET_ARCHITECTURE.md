# PicoClaw Pet 模块架构文档

## 概述

Pet 模块是 PicoClaw 的桌面宠（桌宠）功能模块，提供基于 WebSocket 的桌面客户端连接、角色管理、情绪系统、动作系统和语音功能。

---

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         Gateway                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  AgentLoop                                                │   │
│  │  └── MountHook("pet", petHook)                            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  ChannelManager                                           │   │
│  │  └── PetChannel (WebSocket)                              │   │
│  │       └── PetService                                     │   │
│  │            ├── ConfigManager                              │   │
│  │            │    ├── PetConfig (公开配置)                  │   │
│  │            │    └── CharacterPrivateConfig (私有配置)      │   │
│  │            ├── CharManager                                │   │
│  │            │    └── Character[] + EmotionEngine           │   │
│  │            ├── ActionManager                              │   │
│  │            ├── MemoryStore                                │   │
│  │            └── VoiceLoader                                │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 启动条件

要启动 Pet 模块，需要满足以下条件：

### 1. 配置文件条件

**主配置文件 (`config.json`)** 中需要启用 Pet Channel：

```json
{
  "channels": {
    "pet": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 18792,
      "workspace_path": "./workspace"
    }
  }
}
```

**Pet 配置文件 (`workspace/pet_config.json`)**：

```json
{
  "characters": [
    {
      "id": "pet_001",
      "name": "艾莉",
      "persona": "温柔体贴，善于关心他人",
      "persona_type": "gentle",
      "avatar": "default"
    }
  ],
  "active_id": "pet_001",
  "voice": {
    "model_list": [],
    "default_model": "",
    "asr_enabled": false
  },
  "app": {
    "emotion_enabled": true,
    "reminder_enabled": true,
    "proactive_care": true,
    "proactive_interval_minutes": 30,
    "voice_enabled": false,
    "language": "zh-CN"
  }
}
```

### 2. 私有配置（可选，首次自动创建）

```
workspace/workspaces/{character_id}/config.json
```

```json
{
  "id": "pet_001",
  "emotion_state": {
    "joy": 50,
    "anger": 50,
    "sadness": 50,
    "disgust": 50,
    "surprise": 50,
    "fear": 50
  },
  "mbti": {
    "ie": 50,
    "sn": 50,
    "tf": 50,
    "jp": 50
  },
  "volatility": 0.5,
  "last_update": 1234567890
}
```

---

## 启动流程

```
1. Gateway 启动
   └── config.LoadConfig() 加载主配置
         ↓
2. setupAndStartServices()
   └── channels.NewManager()
         ↓
3. ChannelManager 初始化各 Channel
   └── PetChannel 工厂函数被调用
         ↓
4. NewPetChannel(cfg, msgBus, workspacePath)
   └── pet.NewPetService()
         ↓
5. PetService 初始化
   └── config.NewManager(workspacePath)     加载 pet_config.json
   └── characters.NewManager(chars, cfgMgr)  初始化角色
   └── action.NewActionManager(workspacePath) 加载动作
   └── voice.NewLoader(voiceCfg)              加载语音
   └── memory.NewStore(workspacePath)         初始化记忆
         ↓
6. Service.Start()
   └── 启动情绪衰减定时器
         ↓
7. Gateway 注册 Pet Hook
   └── agentLoop.MountHook("pet", petHook)
```

---

## 目录结构

```
{workspacePath}/
├── config.json              # 主配置文件（Gateway 配置）
│
├── pet_config.json          # Pet 公开配置
│   ├── characters[]         # 角色列表（公开信息）
│   ├── active_id            # 当前激活角色 ID
│   ├── voice                # 语音配置
│   └── app                  # 应用运行时配置
│
├── workspaces/              # 私有数据目录
│   └── {character_id}/
│       └── config.json      # 角色私有配置（情绪、MBTI）
│
└── memory/                  # 记忆存储（可选）
    ├── conversation/        # 对话记忆
    ├── preference/          # 偏好记忆
    ├── fact/               # 事实记忆
    └── persona/            # 人格记忆
```

---

## 核心组件

### PetService (`pkg/pet/service.go`)

主服务入口，协调各子模块。

```go
type PetService struct {
    configManager *config.Manager    // 统一配置管理
    charManager   *characters.Manager // 角色管理
    actionManager *action.ActionManager // 动作管理
    memoryStore   *memory.Store        // 记忆存储
    voiceLoader   *voice.Loader        // 语音加载
}
```

**关键方法：**
- `NewPetService()` - 创建服务
- `Start()` - 启动服务（启动情绪衰减）
- `Stop()` / `Shutdown()` - 停止服务并保存配置

### ConfigManager (`pkg/pet/config/manager.go`)

统一配置管理器，负责加载和保存所有配置。

```go
type Manager struct {
    configLoader           *ConfigLoader  // 底层加载器
    characters             []*CharacterConfig
    voiceConfig            *VoiceConfig
    appConfig             *AppConfig
    activeID              string
    characterPrivateConfig *CharacterPrivateConfig
}
```

**存储位置：**
- 公开配置：`{workspacePath}/pet_config.json`
- 私有配置：`{workspacePath}/workspaces/{id}/config.json`

### CharactersManager (`pkg/pet/characters/manager.go`)

多角色管理器，支持角色切换。

```go
type Manager struct {
    characters    map[string]*Character
    configManager *config.Manager
}
```

**核心功能：**
- 加载角色列表
- 切换当前角色
- 获取/设置角色情绪状态
- 保存角色私有配置

### EmotionEngine (`pkg/pet/emotion/engine.go`)

情绪状态机，管理角色情绪。

```go
type EmotionEngine struct {
    emotions   SixEmotions      // 六维情绪值
    volatility float64          // 波动系数
    lastUpdate time.Time        // 最后更新时间
}
```

**情绪衰减：**
- 每 5 秒执行一次衰减
- 根据 `volatility` 计算衰减幅度

### ActionManager (`pkg/pet/action/manager.go`)

动作管理器，解析 LLM 输出中的动作标签。

**标签格式：** `[action:wave]`

**存储位置：** `{workspacePath}/pet/actions.json`

---

## API 接口

Pet Channel 通过 WebSocket 提供 JSON-RPC 风格接口。

### 请求类型

| Action | 说明 | 数据 |
|--------|------|------|
| `chat` | 发送聊天消息 | `ChatRequest` |
| `character_get` | 获取当前角色信息 | - |
| `character_update` | 更新角色信息 | `CharacterUpdateRequest` |
| `character_switch` | 切换角色 | `CharacterSwitchRequest` |
| `config_get` | 获取应用配置 | - |
| `config_update` | 更新应用配置 | `ConfigUpdateRequest` |
| `emotion_get` | 获取情绪状态 | - |
| `onboarding_config` | 首次配置 | `OnboardingConfigRequest` |
| `health_check` | 健康检查 | - |

### 推送类型

| PushType | 说明 | 数据 |
|----------|------|------|
| `character_switch` | 角色切换通知 | `CharacterSwitchPush` |
| `init_status` | 初始化状态 | `InitStatusPush` |
| `emotion_update` | 情绪更新 | `EmotionState` |

---

## Hook 系统

Pet Hook 挂载到 AgentLoop，在 LLM 调用前后执行。

```go
// pkg/pet/hooks.go
petHook := pet.NewPetHook(
    svc.CharManager(),
    svc.ActionManager(),
    svc,
)
agentLoop.MountHook("pet", petHook)
```

**功能：**
1. **注入角色信息** - 在 System Prompt 中注入角色人格
2. **解析情绪标签** - 从 LLM 输出中解析 `[emotion:happy]` 标签
3. **解析动作标签** - 从 LLM 输出中解析 `[action:wave]` 标签
4. **解析 MBTI 标签** - 从 LLM 输出中解析 `[mbti:ie=60]` 标签

---

## 依赖关系图

```
PetChannel
    └── PetService
          ├── ConfigManager
          │     └── ConfigLoader
          │           └── pet_config.json
          ├── CharactersManager
          │     └── Character
          │           └── EmotionEngine
          ├── ActionManager
          │     └── actions.json
          ├── MemoryStore
          │     └── memory/*.json
          └── VoiceLoader
                └── VoiceConfig
```

---

## PicoClaw 初始化

PicoClaw 提供 `onboard` 命令来初始化配置和工作空间。

### 初始化命令

```bash
# 首次初始化
picoclaw onboard

# 带加密的初始化（加密存储的 API Key）
picoclaw onboard --enc
```

### 初始化流程

```
picoclaw onboard
    ↓
1. 检查 config.json 是否存在
    ├── 不存在 → 使用默认配置
    └── 存在 → 保留现有配置
    ↓
2. （可选）设置加密密钥和密码
    ↓
3. 保存 config.json 到配置目录
    ↓
4. 创建工作空间模板
    └── 复制 workspace/ 目录下的文件
        ├── SOUL.md      # 角色灵魂定义
        ├── USER.md      # 用户信息
        ├── AGENT.md     # Agent 提示词
        ├── memory/      # 记忆目录
        └── skills/      # 技能目录
    ↓
5. 完成提示
```

### 初始化后的文件

```
~/.picoclaw/                    # 默认配置目录（Linux/macOS）
%USERPROFILE%\.picoclaw\        # 默认配置目录（Windows）
├── config.json                 # 主配置文件
└── workspace/                  # 工作空间
    ├── SOUL.md
    ├── USER.md
    ├── AGENT.md
    ├── memory/
    └── skills/
```

### 添加 Pet 模块配置

初始化完成后，需要在 `config.json` 中添加 Pet Channel 配置：

```json
{
  "channels": {
    "pet": {
      "enabled": true,
      "host": "0.0.0.0",
      "port": 18792,
      "workspace_path": ".picoclaw/workspace"
    }
  }
}
```

或者通过环境变量：

```bash
export PICOCLAW_CHANNELS_PET_ENABLED=true
export PICOCLAW_CHANNELS_PET_PORT=18792
```

---

## 启动检查清单

1. ✅ PicoClaw 已安装（`picoclaw version` 可正常执行）
2. ✅ 已运行 `picoclaw onboard` 初始化配置
3. ✅ 主配置 `config.json` 中 `channels.pet.enabled = true`
4. ✅ 指定了 `workspace_path`
5. ✅ （可选）`workspace/pet_config.json` 存在，不存在则自动创建默认配置
6. ✅ （可选）`workspace/workspaces/{id}/config.json` 存在，不存在则自动创建
7. ✅ Gateway 已启动并监听端口

---

## 快速启动流程

```bash
# 1. 安装 PicoClaw（如果未安装）
# 请参考项目 README

# 2. 初始化配置
picoclaw onboard

# 3. 编辑配置，启用 Pet Channel
# 编辑 ~/.picoclaw/config.json 或 %USERPROFILE%\.picoclaw\config.json
# 添加 pet channel 配置

# 4. 启动 Gateway
picoclaw gateway

# 5. Pet Channel 将在端口 18792 上启动
```

---

## 常见问题

### Q: Pet Channel 无法启动？
检查：
1. `config.json` 中 `channels.pet.enabled` 是否为 `true`
2. `workspace_path` 目录是否存在且有读写权限
3. 端口是否被占用

### Q: 角色切换后配置没保存？
确认调用了 `PetService.Shutdown()`，它会在停止时保存所有配置。

### Q: 情绪衰减没有生效？
检查：
1. `app.emotion_enabled` 是否为 `true`
2. `CharacterPrivateConfig.volatility` 是否大于 0

---

## 相关文件

| 文件 | 说明 |
|------|------|
| `pkg/pet/service.go` | Pet 主服务 |
| `pkg/pet/config/manager.go` | 配置管理器 |
| `pkg/pet/characters/manager.go` | 角色管理器 |
| `pkg/pet/emotion/engine.go` | 情绪引擎 |
| `pkg/pet/action/manager.go` | 动作管理器 |
| `pkg/pet/hooks.go` | Hook 实现 |
| `pkg/channels/pet/channel.go` | WebSocket 通道 |
| `pkg/channels/pet/init.go` | 通道注册 |
