# GoClaw 快速集成说明

GoClaw 现在由三部分组成：

- `picoclaw` / `picoclaw-launcher`: 后端和 API 网关
- `petclaw`: 完整前端控制台（Next.js）
- `electron-frontend`: 桌宠形象与桌面窗口（Electron + React）

## 集成结果

- 桌宠主窗体继续使用 `electron-frontend` 的角色动画。
- 在桌宠窗口中新增 `控制台` 按钮，会打开 `petclaw` 前端页面。
- `petclaw` WebSocket 不再使用硬编码测试 token，而是通过 `/api/pico/token` 获取真实 token，并使用 Pico 协议连接。
- 如果 Pico channel 没启用，会自动调用 `/api/pico/setup` 完成初始化。

## 本地启动（Windows）

仓库根目录运行：

```powershell
./scripts/run-goclaw-dev.ps1
```

这个脚本会尝试同时启动：

1. PicoClaw launcher（自动查找 `build/picoclaw-launcher.exe` / `picoclaw-launcher.exe` / `picoclaw-web.exe`）
2. `petclaw`（默认生产模式 `npm run start -- --hostname 127.0.0.1 --port 3000`）
3. `electron-frontend`（`npm run dev` + `npx electron src/main.js`）

你也可以手动指定 launcher：

```powershell
./scripts/run-goclaw-dev.ps1 -LauncherBin "D:\tools\picoclaw-launcher.exe"
```

## 自定义控制台地址

Electron 打开的控制台地址默认是 `http://127.0.0.1:3000`。

- 通过脚本参数覆盖：`-DashboardUrl "http://127.0.0.1:3001"`
- 或直接设置环境变量：`GOCLAW_DASHBOARD_URL`
