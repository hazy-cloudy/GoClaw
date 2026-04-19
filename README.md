# ClawPet 开发与启动说明

ClawPet 是这个仓库当前正在联调的桌宠产品形态。

当前请以米色 `petclaw` 面板为准。桌宠右下角 `S` 按钮打开的控制台、初始化页面、聊天会话历史和语音播放，都是围绕 `petclaw` 这一套实现的。

## 先说结论

如果你只是想把项目跑起来，请直接在仓库根目录的 `cmd.exe` 执行：

```cmd
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

执行成功后，你主要只需要关心两个入口：

- 米色控制台：`http://127.0.0.1:3000`
- 初始化页面：`http://127.0.0.1:3000/onboarding?mode=rerun`

桌宠窗口会自动启动，右下角会出现宠物；点击桌宠上的 `S` 会打开米色控制台。

## 这套东西由什么组成

当前本地联调会启动 4 个组件：

- `launcher`：本地启动器和控制入口，默认 `127.0.0.1:18800`
- `gateway`：PicoClaw 网关后端，默认 `127.0.0.1:18790`
- `petclaw`：米色控制台面板，默认 `127.0.0.1:3000`
- `electron-frontend`：桌宠窗口渲染层，默认 `127.0.0.1:5173`

当前默认关系：

- `petclaw` 是主控制台
- `petclaw /onboarding` 是初始化页
- `electron-frontend` 负责桌宠窗口本体
- 桌宠 `S` 按钮通过 Electron 打开 `petclaw`

## 推荐启动方式

推荐直接在 `cmd.exe` 使用完整联调脚本：

```cmd
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

这条命令会自动完成：

1. 清理旧的 GoClaw / Electron / Node 进程
2. 启动 launcher：`127.0.0.1:18800`
3. 启动 gateway
4. 启动 `petclaw`：`127.0.0.1:3000`
5. 启动桌宠渲染：`127.0.0.1:5173`
6. 启动 Electron 桌宠主进程，并把设置/初始化窗口指向 `petclaw`

### 如果用户不是 PowerShell

这个问题需要在文档里明确说明。

当前仓库的“一键完整联调脚本”是 `PowerShell` 版本，所以不同环境建议这样启动：

#### 1. Windows 用户，但不是在 PowerShell 里

可以直接用下面任意一种方式：

- 在 `cmd` 里执行：

```cmd
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

- 在 `Git Bash` / `MSYS2` / `Nushell` 里执行：

```bash
powershell.exe -ExecutionPolicy Bypass -File ./scripts/run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

也就是说，文档不能写成“只能在 PowerShell 里启动”，而应该写成“当前官方启动脚本由 PowerShell 执行，但 Windows 下可以从别的 shell 调它”。

#### 2. macOS / Linux 用户

当前仓库还没有提供同等功能的 `run-goclaw-dev.sh` 一键脚本。

所以这两类用户目前有两种选择：

- 安装 PowerShell 7 后，直接运行同一个 `.ps1`
- 或者手动分别启动各组件

如果是给外部用户看的文档，应该明确写出这一点，不要让人误以为仓库已经提供了 shell 原生一键脚本。

## 启动入口

仓库根目录当前保留的是 PowerShell 启动入口：

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1
```

这个入口当前会委托给 `scripts/run-goclaw-dev.ps1`，启动的也是同一套完整链路。

但如果是给外部用户看的文档，当前更推荐直接写 `cmd` 命令，而不是要求他们先切到 PowerShell。

## 启动后应该看到什么

正常情况下：

- `http://127.0.0.1:18800`：launcher 可访问
- `http://127.0.0.1:3000`：米色 `petclaw` 控制台可访问
- `http://127.0.0.1:3000/onboarding?mode=rerun`：初始化页可直接打开
- `http://127.0.0.1:5173`：桌宠渲染页已启动
- 桌面右下角出现桌宠窗口

## 当前主界面说明

### 现在应该用哪个前端

当前请使用米色 `petclaw` 面板：

- 主面板：`http://127.0.0.1:3000`
- 初始化页：`http://127.0.0.1:3000/onboarding?mode=rerun`

`electron-frontend` 当前职责是：

- 渲染桌宠本体
- 承接 Electron 外壳
- 打开 `petclaw` 面板

如果你看到一套深色聊天设置页，那不是当前默认主界面。

### 聊天行为

当前 `petclaw` 聊天行为：

- `新建聊天` 会创建新的前端 session id，并重连 websocket
- 左侧对话记录只显示每个会话的首条用户输入
- 当前历史是前端本地会话历史，不是服务端持久化历史接口

### 引导页数据

当前 onboarding 已经不再只是前端本地保存。

- 引导页会把完整快照提交到后端
- 后端会持久化保存这份快照
- 其中一部分字段会同步到当前 `pet` 运行配置

详细数据结构和接口契约见：

- [docs/zh/clawpet-onboarding.md](docs/zh/clawpet-onboarding.md)

### 语音行为

当前语音行为：

- 优先使用后端 `audio` push 做播放
- `petclaw` 会合并后端音频分片后播放
- 浏览器本地 TTS 自动兜底已关闭
- 输入框右侧麦克风按钮用于语音输入，依赖浏览器语音识别能力

## 环境要求

- Go `>= 1.23`
- Node.js `>= 18`，建议 `20` 或 `22`
- npm
- Windows 下建议安装 C/C++ 构建环境，否则部分 SQLite 能力可能降级

## 常用命令

### 启动完整联调环境

```cmd
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### 启动完整联调环境，但使用 `petclaw` 生产模式

```cmd
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode prod
```

### 只启动 launcher

```cmd
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode launcher
```

### 按端口逐个手动启动（cmd）

如果你不想跑完整联调脚本，而是想明确知道每个端口怎么起来，可以在多个 `cmd` 窗口里分别执行：

#### 18800: launcher

前提：仓库里已经有 launcher 二进制，例如 `build\picoclaw-launcher.exe`

```cmd
set PICOCLAW_BINARY=%CD%\picoclaw.exe
set PICOCLAW_LAUNCHER_TOKEN=goclaw-local-token
set PICOCLAW_HOME=%CD%\.goclaw-runtime
set PICOCLAW_CONFIG=%CD%\.goclaw-runtime\config.json
build\picoclaw-launcher.exe -no-browser "%PICOCLAW_CONFIG%"
```

说明：

- 这个进程会监听 `127.0.0.1:18800`
- 负责启动 gateway、做鉴权和控制面

#### 18790: gateway

一般由 launcher 负责拉起，不推荐手动单独启动。

如果只是调试 gateway，可以单独执行：

```cmd
picoclaw.exe gateway
```

说明：

- 这个进程默认监听 `127.0.0.1:18790`
- 负责真实的 `pet` HTTP / WebSocket 通道

#### 3000: petclaw 米色面板

```cmd
cd /d %CD%\petclaw
npm install
npm run dev -- --hostname 127.0.0.1 --port 3000
```

说明：

- 这是当前给用户看的主控制台
- 初始化页也在这里：`/onboarding`

#### 5173: electron-frontend 渲染层

```cmd
cd /d %CD%\electron-frontend
npm install
npm run dev -- --host 127.0.0.1 --port 5173 --strictPort
```

说明：

- 这是桌宠窗口的开发期渲染服务
- 主要给 Electron 用，不是当前主控制台入口

#### Electron 主进程

在另一个 `cmd` 窗口里：

```cmd
cd /d %CD%\electron-frontend
set GOCLAW_DASHBOARD_URL=http://127.0.0.1:3000
set GOCLAW_LAUNCHER_TOKEN=goclaw-local-token
npx electron src/main.js
```

说明：

- 这一步不会额外占一个 HTTP 端口
- 它负责把桌宠窗口和米色面板串起来

## 常见问题

### 1. 面板是空白的

先检查 `petclaw` 是否真的起来了：

```powershell
curl.exe -i http://127.0.0.1:3000
```

如果 `3000` 没起来，但你点了桌宠的 `S`，Electron 只会弹出一个空白壳窗口。

### 2. 打开的不是米色面板，而是另一套深色界面

请确认你启动的是完整联调脚本：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

当前默认应当由桌宠打开 `petclaw:3000`，不是旧的深色设置页。

### 3. 想强制重新走初始化页

直接打开：

```text
http://127.0.0.1:3000/onboarding?mode=rerun
```

### 4. 想看网关是否正常

```powershell
curl.exe -i http://127.0.0.1:18790/health
curl.exe -i http://127.0.0.1:18790/ready
```

### 5. 长文本没有回复

当前 `petclaw` 已兼容解析对象、JSON 字符串、纯字符串三种 `ai_chat` / `audio` push 形态。如果再次出现该问题，请优先检查浏览器 DevTools 的 `WS` 消息内容，以及网关日志里实际推送的结构。

### 6. `make run` 不可用

Windows 默认没有 `make`，当前推荐直接使用上面的 PowerShell 启动脚本。

## 4 个端口能不能统一

可以讨论统一，但要先区分“对外统一入口”和“内部实际进程端口”。

### 当前为什么会有 4 个端口

- `18800` 是 launcher，用来做控制、鉴权、网关管理
- `18790` 是 gateway，负责真实的 `pet` HTTP / WebSocket 通道
- `3000` 是 `petclaw` 前端页面
- `5173` 是 Electron 开发时的前端渲染服务

也就是说，这 4 个端口里：

- `3000` 是你真正给人看的控制台入口
- `5173` 基本只是开发期给 Electron 用的
- `18800` 和 `18790` 更偏内部服务端口

### 能不能彻底只剩 1 个端口

短期内不太适合直接“物理合并成一个进程一个端口”，因为这几层职责不同：

- launcher 负责控制面
- gateway 负责真实通道
- petclaw 是独立前端
- electron-frontend 是桌宠壳层

如果现在硬合并，改动会比较大，而且会影响现有启动脚本、Electron 行为和鉴权链路。

### 更现实的统一方式

更推荐做“对外统一入口”，而不是强行把 4 个内部监听都抹掉。

推荐的统一方向是：

1. 对用户只暴露 `3000`
2. `petclaw` 作为唯一对外入口
3. `18800`、`18790`、`5173` 继续作为内部开发 / 运行端口存在
4. Electron 打包后把 `5173` 消掉，改成加载本地静态资源

这样别人看文档时，只需要知道：

- 运行一条命令
- 打开 `3000`
- 其它端口都是内部实现细节

### 如果后面真要继续统一

优先级建议：

1. 先把文档和启动入口统一成“只讲 3000”
2. 再把 Electron 的 `5173` 从开发端口改成打包静态资源
3. 最后再评估 launcher / gateway 是否要做反向代理收口

## 目录说明

- `scripts/run-goclaw-dev.ps1`：完整联调启动脚本
- `GoClaw-OneClickStart.ps1`：一键启动入口，会转调完整联调脚本
- `petclaw/`：米色控制台面板与初始化页
- `electron-frontend/`：桌宠窗口与 Electron 外壳
- `cmd/picoclaw/`：网关入口
- `pkg/channels/pet/`：`pet` 通道 HTTP / WebSocket 入口
- `pkg/pet/`：桌宠业务核心

## 建议阅读顺序

如果你是第一次接手这个仓库，建议优先按下面顺序理解：

1. `scripts/run-goclaw-dev.ps1`
2. `petclaw/`
3. `electron-frontend/src/main.js`
4. `pkg/channels/pet/`

这样最容易看清楚：谁负责启动、谁负责米色面板、谁负责桌宠窗口、谁负责实际聊天链路。
