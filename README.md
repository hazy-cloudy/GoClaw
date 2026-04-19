# ClawPet（桌宠）开发与启动说明

ClawPet 是这个仓库里当前正在联调的桌宠产品形态。

你真正会用到的本地组件有 4 个：

- `launcher`：本地启动器与控制面板入口，默认 `127.0.0.1:18800`
- `gateway`：PicoClaw 网关后端，由 launcher 拉起，默认 `127.0.0.1:18790`
- `petclaw`：米色控制台面板，带初始化页，默认 `127.0.0.1:3000`
- `electron-frontend`：桌宠本体渲染层，默认 `127.0.0.1:5173`

当前默认交互关系：

- 桌宠右下角 `S` 按钮打开的是 `petclaw` 米色面板
- 初始化页来自 `petclaw` 的 `/onboarding`
- `electron-frontend` 负责桌宠窗口本身，不再作为主要控制台面板使用

## 推荐启动方式

推荐直接使用仓库内的联调脚本：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

这条命令会自动完成：

1. 清理旧的 GoClaw / Electron / Node 进程
2. 启动 launcher：`127.0.0.1:18800`
3. 启动 gateway
4. 启动 `petclaw` 米色面板：`127.0.0.1:3000`
5. 启动桌宠渲染：`127.0.0.1:5173`
6. 启动 Electron 桌宠主进程，并把设置/初始化窗口指向 `petclaw`

## 一键启动

如果你希望双击启动，可以直接用：

- `GoClaw-OneClickStart.bat`

或：

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1
```

现在这个一键入口会委托给 `scripts/run-goclaw-dev.ps1`，启动的也是同一套完整链路，而不是旧的简化版流程。

## 启动后应该看到什么

正常情况下：

- `http://127.0.0.1:18800`：launcher 已可访问
- `http://127.0.0.1:3000`：米色 `petclaw` 控制台已可访问
- `http://127.0.0.1:3000/onboarding?mode=rerun`：初始化页可直接打开
- `http://127.0.0.1:5173`：桌宠渲染页已启动
- 桌面右下角会出现桌宠窗口

## 面板说明

### 你应该使用哪一个前端

日常联调和初始化，应该使用 `petclaw` 这套米色面板：

- 地址：`http://127.0.0.1:3000`
- 初始化页：`http://127.0.0.1:3000/onboarding?mode=rerun`

`electron-frontend` 的职责是：

- 渲染桌宠本体
- 承接桌宠窗口
- 通过 Electron 打开 `petclaw` 面板

如果你看到的是一套深色聊天设置页，那不是当前默认应该使用的主控制台。

## 环境要求

- Go `>= 1.23`
- Node.js `>= 18`，建议 `20` 或 `22`
- npm
- Windows 下建议安装 C/C++ 构建环境，否则部分 SQLite 能力可能降级

## 常用命令

### 启动完整联调环境

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

### 启动完整联调环境但使用 `petclaw` 生产模式

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode prod
```

### 仅启动 launcher

```powershell
powershell -ExecutionPolicy Bypass -File .\GoClaw-OneClickStart.ps1 -Mode launcher
```

## 常见问题

### 1. 面板是空白的

先检查 `petclaw` 是否真的起来了：

```powershell
curl.exe -i http://127.0.0.1:3000
```

如果 `3000` 没起来，但你点了桌宠的 `S`，Electron 只会弹出一个空窗口底板。

### 2. 打开的不是米色面板，而是另一套深色界面

请确认你启动的是完整联调脚本：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

当前默认应当由桌宠打开 `petclaw:3000`，不是 `electron-frontend/settings.html` 那套旧面板。

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

### 5. `make run` 不可用

Windows 默认没有 `make`，本仓库当前推荐直接使用上面的 PowerShell 启动脚本。

## 目录说明

- `scripts/run-goclaw-dev.ps1`：完整联调启动脚本
- `GoClaw-OneClickStart.ps1`：一键启动入口，会转调完整联调脚本
- `petclaw/`：米色控制台面板与初始化页
- `electron-frontend/`：桌宠本体与 Electron 外壳
- `cmd/picoclaw/`：网关入口
- `pkg/channels/pet/`：`pet` 通道 HTTP / WebSocket 入口
- `pkg/pet/`：桌宠业务核心

## 补充说明

如果你是第一次接手这个仓库，建议优先按下面顺序理解：

1. `scripts/run-goclaw-dev.ps1`
2. `petclaw/`
3. `electron-frontend/src/main.js`
4. `pkg/channels/pet/`

这样最容易看清楚：谁负责启动、谁负责米色面板、谁负责桌宠窗口、谁负责实际聊天链路。
