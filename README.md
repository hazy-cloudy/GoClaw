# GoClaw - 桌面宠物 AI 助手

> 🐾 一个基于 AI 的桌面宠物伴侣，支持流式对话、情绪交互、语音合成

![Version](https://img.shields.io/badge/version-2.5-blue)
![Go](https://img.shields.io/badge/Go-%3E%3D1.23-00ADD8)
![Node.js](https://img.shields.io/badge/Node.js-%3E%3D18-339933)
![License](https://img.shields.io/badge/license-MIT-green)

---

## 📖 项目简介

GoClaw 是一个**桌面宠物 AI 助手**，它将 AI 对话能力与可爱的桌宠形象结合，为用户提供陪伴式交互体验。

### ✨ 核心特性

- 🎭 **AI 桌宠对话**：基于 LLM 的智能对话，支持流式输出
- 😊 **情绪系统**：桌宠会根据对话内容产生情绪变化（喜、怒、哀、惧、惊、厌）
- 🎤 **语音合成**：支持 TTS 语音流式播放
- 🎬 **动作触发**：LLM 自动解析动作标签并触发表情动画
- 📅 **定时任务**：支持提醒、关怀等定时推送
- 💾 **记忆系统**：桌宠会记住用户的偏好和历史对话
- 🖥️ **桌面宠物**：透明窗口渲染，可拖拽互动
- 🌐 **多平台**：支持 Windows、Linux、macOS

---

## 🏗️ 架构设计

### 核心组件

GoClaw 由 4 个核心组件组成：

| 组件 | 端口 | 说明 | 技术栈 |
|------|------|------|--------|
| **Launcher** | 18800 | 本地服务管理器，负责启动和监控所有组件 | Go |
| **Gateway** | 18790 | PicoClaw 后端网关，处理 AI 对话、MCP 工具调用 | Go |
| **Petclaw** | 3000 | 桌宠控制台（Web UI），用于聊天、设置、管理 | Next.js + React |
| **Electron** | 5173 | 桌面宠物渲染窗口，显示桌宠形象 | Electron + React |

### 组件关系

```
┌─────────────────────────────────────────────┐
│                GoClaw 架构                    │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────────┐    ┌──────────────┐       │
│  │   Launcher   │───▶│   Gateway    │       │
│  │  (18800)     │    │  (18790)     │       │
│  └──────┬───────┘    └──────┬───────┘       │
│         │                   │                │
│         ▼                   ▼                │
│  ┌──────────────┐    ┌──────────────┐       │
│  │  Petclaw     │    │   Electron   │       │
│  │  (3000)      │    │  (5173)      │       │
│  │  控制台 UI    │    │  桌宠窗口     │       │
│  └──────────────┘    └──────────────┘       │
│                                             │
└─────────────────────────────────────────────┘
```

**工作流程**：
1. Launcher 启动并监控 Gateway、Petclaw、Electron
2. Gateway 连接 AI 模型（OpenAI、Anthropic 等）
3. Petclaw 提供 Web 控制台（聊天、设置、管理）
4. Electron 渲染桌宠窗口，通过 WebSocket 与 Gateway 通信

---

## 🚀 快速开始

### 环境要求

- **Go**: `>= 1.23`
- **Node.js**: `>= 18`（推荐 `20` 或 `22`）
- **npm** 或 **pnpm**
- **Windows**: 建议安装 C/C++ 构建环境（用于 SQLite）

### 方式 1：使用 Release 包（推荐最终用户）

1. 从 [GitHub Releases](https://github.com/1024XEngineer/GoClaw/releases) 下载最新包：
   - `clawpet_AllInOne_Windows_x86_64.zip`

2. 解压到任意目录

3. 双击运行：
   ```
   GoClaw-OneClickStart.bat
   ```

4. 首次启动会自动安装依赖，等待启动完成

5. 访问控制台：`http://127.0.0.1:3000`

### 方式 2：从源码运行（推荐开发者）

```bash
# 1. 克隆仓库
git clone https://github.com/1024XEngineer/GoClaw.git
cd GoClaw

# 2. 安装前端依赖
cd petclaw && npm install && cd ..
cd electron-frontend && npm install && cd ..

# 3. 启动开发环境
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### 方式 3：使用 Makefile

```bash
# 构建主程序
make build

# 构建 Launcher
make build-launcher

# 运行
make run
```

---

## 📡 服务端口

启动后，以下端口将被占用：

| 服务 | 地址 | 用途 |
|------|------|------|
| Launcher | http://127.0.0.1:18800 | 服务管理 API |
| Gateway | http://127.0.0.1:18790 | AI 网关 + WebSocket |
| Petclaw | http://127.0.0.1:3000 | Web 控制台 |
| Electron | http://127.0.0.1:5173 | 桌宠渲染窗口 |

**重要页面**：
- 控制台首页：`http://127.0.0.1:3000`
- 初始化引导：`http://127.0.0.1:3000/onboarding`
- 重新引导：`http://127.0.0.1:3000/onboarding?mode=rerun`

---

## 🎨 界面说明

### 主控制台（Petclaw）

Petclaw 是**主要的控制台界面**，提供：

- 💬 **聊天界面**：与 AI 桌宠对话
- ⚙️ **设置页面**：配置模型、角色、语音等
- 📋 **会话历史**：查看和管理对话记录
- 🎭 **角色管理**：创建和切换桌宠角色
- 📅 **定时任务**：设置提醒和关怀任务

### 桌宠窗口（Electron）

Electron 窗口负责：

- 🐾 **渲染桌宠**：显示 2D/3D 宠物形象
- 🎬 **播放动画**：根据动作触发播放表情
- 🔊 **播放语音**：播放 TTS 合成的语音
- 🖱️ **桌面交互**：拖拽、点击等桌面互动

**注意**：Electron 窗口不是控制台，点击 `S` 按钮会打开 Petclaw 控制台。

---

## 🔌 API 接口

GoClaw 提供完整的 HTTP + WebSocket API。

### HTTP API

- **Gateway Token**: `GET /pet/token` - 获取 WebSocket 连接地址
- **Gateway 状态**: `GET /api/gateway/status` - 检查服务状态
- **Pet 初始化**: `POST /api/pet/setup` - 初始化桌宠配置
- **会话管理**: `GET/DELETE /api/sessions` - 管理对话会话

### WebSocket API

连接地址：`ws://127.0.0.1:18790/pet/ws?session={sessionId}`

**主要操作**：
- `chat` - 发送聊天消息
- `onboarding_config` - 提交初始化配置
- `character_switch` - 切换角色
- `config_update` - 更新应用配置

**推送类型**：
- `init_status` - 连接初始化状态
- `ai_chat` - AI 聊天回复（流式）
- `audio_and_voice` - 语音流式合成
- `emotion_change` - 情绪变化
- `action_trigger` - 动作触发
- `heartbeat` - 心跳保活

详细 API 文档请查看：[API.md](./API.md) 或 [docs/PET_CHANNEL_API.md](./docs/PET_CHANNEL_API.md)

---

## 🛠️ 常用命令

### 启动完整开发栈

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### 启动生产模式

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode prod
```

### 仅启动 Launcher

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode launcher
```

### 隐藏所有终端窗口

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev -NoTerminalWindows
```

---

## 📁 项目结构

```
GoClaw/
├── cmd/
│   ├── picoclaw/              # Gateway 入口
│   └── picoclaw-launcher-tui/ # Launcher TUI
├── pkg/
│   ├── agent/                 # AI Agent 逻辑
│   ├── channels/pet/          # Pet WebSocket 通道
│   ├── pet/                   # 桌宠业务逻辑
│   ├── memory/                # 记忆系统
│   ├── session/               # 会话管理
│   └── tools/                 # MCP 工具
├── petclaw/                   # Web 控制台（Next.js）
│   ├── app/                   # 页面路由
│   ├── components/            # React 组件
│   └── hooks/                 # React Hooks
├── electron-frontend/         # 桌面宠物（Electron）
│   ├── src/
│   │   ├── main.js            # Electron 主进程
│   │   └── preload.js         # 预加载脚本
│   └── startup.html           # 启动页
├── scripts/
│   └── run-goclaw-dev.ps1     # 开发启动脚本
├── workspace/                 # 工作区
│   ├── skills/                # 技能目录
│   ├── SOUL.md                # 角色定义
│   └── USER.md                # 用户配置
└── docs/                      # 文档
    ├── PET_CHANNEL_API.md     # WebSocket API 文档
    └── configuration.md       # 配置说明
```

---

## 🔍 故障排查

### 控制台空白

检查 Petclaw 是否运行：

```powershell
curl.exe -i http://127.0.0.1:3000
```

如果 3000 端口无响应，点击桌宠 `S` 按钮会显示空白窗口。

### 打开了错误的界面

如果打开的是深色设置页而不是米色控制台，重新启动：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### 重新运行初始化引导

访问：
```
http://127.0.0.1:3000/onboarding?mode=rerun
```

### 检查 Gateway 健康状态

```powershell
curl.exe -i http://127.0.0.1:18790/health
curl.exe -i http://127.0.0.1:18790/ready
```

### 长回复缺失

Petclaw 现已支持 `ai_chat` 推送的三种格式（对象、JSON 字符串、原始字符串）。如果长回复仍然缺失，请在 DevTools 中检查 WebSocket 帧的原始载荷。

### 有文字无语音

检查是否收到 `push_type=audio_and_voice` 推送，以及 `voice_enabled` 是否启用。

---

## 📚 推荐阅读顺序

如果你是新接触这个代码库，建议按以下顺序阅读：

1. **`scripts/run-goclaw-dev.ps1`** - 了解启动流程
2. **`petclaw/`** - 了解控制台 UI
3. **`electron-frontend/src/main.js`** - 了解桌宠窗口
4. **`pkg/channels/pet/`** - 了解 WebSocket 通信
5. **`docs/PET_CHANNEL_API.md`** - 了解 API 接口

这样可以最清晰地理解启动流程、UI 架构和对话管道。

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

- 🐛 **报告 Bug**：提交 Issue
- 💡 **功能建议**：提交 Issue 并标注 `enhancement`
- 🔧 **代码贡献**：Fork 仓库并提交 PR
- 📝 **文档改进**：欢迎翻译和补充文档

详细贡献指南请查看：[CONTRIBUTING.md](./CONTRIBUTING.md)

---

## 📄 许可证

本项目采用 [MIT 许可证](./LICENSE)。

---

## 🌟 致谢

- [PicoClaw](https://github.com/sipeed/picoclaw) - 底层 AI 网关框架
- [Sipeed](https://github.com/sipeed) - 硬件和开源支持
- 所有贡献者和用户

---

**🐾 享受与你的 AI 桌宠互动的时光吧！**
