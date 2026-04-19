# PetClaw 控制台

`petclaw` 是当前 ClawPet 默认使用的米色控制台面板。

它负责：

- 主聊天界面
- 初始化页面 `/onboarding`
- 本地会话历史展示
- 网关状态展示
- 接收并播放后端 `audio` push
- 提交 onboarding 正式数据到后端

## 启动

推荐不要单独只跑这一层，而是从仓库根目录启动完整联调环境：

```cmd
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 -Restart -PetclawMode dev
```

如果你当前不是在 PowerShell 里：

- Windows `cmd` 可以直接调用这条命令
- `Git Bash` / `MSYS2` / `Nushell` 可以改用 `powershell.exe`
- macOS / Linux 目前没有同等功能的 `.sh` 一键脚本

如果你只想单独启动这个面板：

```bash
npm install
npm run dev -- --hostname 127.0.0.1 --port 3000
```

默认地址：

- `http://127.0.0.1:3000`

初始化页：

- `http://127.0.0.1:3000/onboarding?mode=rerun`

## 生产检查

```bash
npm run build
npm run start -- --hostname 127.0.0.1 --port 3000
```

## 当前行为

### 聊天会话

- `新建聊天` 会创建新的前端 session id
- 左侧历史记录只显示每个会话的首条用户输入
- 当前历史是前端本地会话历史，不是服务端持久化列表

### 语音

- 后端 `audio` push 是主要语音播放来源
- 浏览器本地 TTS 自动兜底已关闭
- 麦克风按钮用于语音输入，依赖浏览器语音识别能力

### onboarding 数据

- 当前引导页提交会调用正式后端接口
- 后端会保存完整 onboarding 快照
- 这不再只是单纯的前端 `localStorage`

详细说明见：

- [../docs/zh/clawpet-onboarding.md](../docs/zh/clawpet-onboarding.md)

## 目录职责

- `app/`: Next.js 路由（`/` 主面板，`/onboarding` 初始化页）
- `components/`: UI 与业务组件
- `hooks/`: 聊天、网关状态、语音输入等状态逻辑
- `lib/`: API 客户端、onboarding 存储、工具函数
- `styles/`: 全局样式与主题样式

## 常见问题

- `401 /api/pico/token`: launcher token 或鉴权链路异常，建议重启 `scripts/run-goclaw-dev.ps1 -Restart`
- websocket 连不上：先确认 launcher 是否监听在 `127.0.0.1:18800`
- 面板空白：先确认 `petclaw` 是否真的启动在 `127.0.0.1:3000`
- 长文本看起来没回复：优先检查 DevTools 里 `WS` 的 `ai_chat` 帧具体长什么样
- 有文字无语音：先确认后端是否真的推送了 `push_type=audio`
- 桌宠打开了错误页面：确认 `GOCLAW_DASHBOARD_URL` 指向的是这个控制台
