# ClawPet 打包与启动流程文档

## 项目概述

ClawPet 是一个基于 Electron + Go 的跨平台桌面宠物应用，包含以下核心组件：

- **Electron 主进程**：管理窗口和后端服务
- **Next.js 前端**：提供控制面板和设置界面
- **Go Gateway**：后端网关服务（端口 18790）
- **Go Launcher**：Launcher 服务（端口 18800），内嵌前端静态资源

---

## 一、窗口尺寸配置

### 设置窗口尺寸调整

**文件位置**：`clawpet-frontend/clawpet/electron/main.js` (第 648-670 行)

**修改内容**：
```javascript
// 计算窗口尺寸（宽度 7/10，高度 2/3）
const { width, height } = screen.getPrimaryDisplay().workAreaSize;
const settingsWidth = Math.round(width * 0.7);     // 7/10 宽度
const settingsHeight = Math.round(height * 0.667); // 2/3 高度
```

**窗口特性**：
- `frame: false` - 无窗口边框（隐藏标题栏）
- `transparent: false` - 非透明背景
- `backgroundColor: '#f7ecdf'` - 米色背景
- `autoHideMenuBar: true` - 自动隐藏菜单栏

### 桌宠窗口尺寸

**文件位置**：`clawpet-frontend/clawpet/electron/main.js` (第 42-45 行)

```javascript
const PET_WIDTH = 280;   // 桌宠窗口宽度
const PET_HEIGHT = 380;  // 桌宠窗口高度
```

---

## 二、本地打包流程

### 前置要求

- **Go 1.21+**
- **Node.js 18+**
- **npm 或 pnpm**
- **Windows 10/11**（或其他支持 Electron 的系统）

### 打包步骤

**脚本位置**：`clawpet-frontend/clawpet/build-and-test.ps1`

#### Step 1: 编译 Go 后端

```powershell
go build -tags "goolm,stdjson" -o "dist\picoclaw.exe" ./cmd/picoclaw
```

输出：`dist/picoclaw.exe` (Gateway 服务)

#### Step 2: 编译 Launcher 后端

```powershell
go build -tags "goolm,stdjson" -o "dist\picoclaw-web.exe" ./web/backend
```

输出：`dist/picoclaw-web.exe` (Launcher 服务)

#### Step 3: 复制后端二进制

```powershell
Copy-Item "dist\picoclaw.exe" "clawpet-frontend/clawpet/picoclaw.exe" -Force
Copy-Item "dist\picoclaw-web.exe" "clawpet-frontend/clawpet/picoclaw-web.exe" -Force
```

#### Step 4: 构建 Next.js 前端

```bash
cd clawpet-frontend/clawpet
npm install        # 安装依赖（如果 node_modules 不存在）
npm run build      # 生成 .next 目录
```

**重要**：必须执行 `npm run build`，否则打包后的 Electron 应用没有前端静态资源。

#### Step 5: Electron 打包

```bash
npx electron-builder --win --publish never
```

输出：
- `dist/win-unpacked/` - 未打包的目录（用于调试）
- `dist/ClawPet Setup 0.1.0.exe` - NSIS 安装包
- `dist/ClawPet 0.1.0.exe` - 便携版

#### Step 6: 复制后端二进制到打包目录

```powershell
Copy-Item "picoclaw.exe" "dist\win-unpacked\picoclaw.exe" -Force
Copy-Item "picoclaw-web.exe" "dist\win-unpacked\picoclaw-web.exe" -Force
```

### 运行打包后的应用

```powershell
# 方法 1: 运行 win-unpacked 目录（用于调试）
& "clawpet-frontend/clawpet/dist/win-unpacked/ClawPet.exe"

# 方法 2: 运行便携版
& "clawpet-frontend/clawpet/dist/ClawPet 0.1.0.exe"
```

---

## 三、Electron 启动流程

### 启动顺序

```
1. Electron 主进程启动
2. 自动启动后端服务（Launcher + Gateway）
3. 创建启动进度窗口（可选）
4. 轮询检查服务状态（每秒一次）
5. 所有服务就绪后，创建桌宠窗口和设置窗口
```

### 后端服务启动

**文件位置**：`clawpet-frontend/clawpet/electron/main.js` (第 1152-1188 行)

#### Launcher 启动

**端口**：18800  
**功能**：提供前端静态资源和 API  
**配置路径**：`~/.goclaw-runtime/config.json`

```javascript
launcherProcess = spawn(exePath, [
  '-port', '18800',
  '-no-browser',
  configPath
], {
  env: {
    PICOCLAW_LAUNCHER_TOKEN: 'goclaw-local-token',
    PICOCLAW_HOME: configDir,
    PICOCLAW_CONFIG: configPath
  }
});
```

#### Gateway 启动

**端口**：18790  
**功能**：提供 AI 代理和 WebSocket 网关

```javascript
gatewayProcess = spawn(exePath, [
  'gateway',
  '-E'  // 启用环境变量
], {
  env: {
    PICOCLAW_HOME: configDir,
    PICOCLAW_CONFIG: configPath
  }
});
```

### 配置初始化

**自动初始化流程**（第 1193-1223 行）：

1. 检查 `~/.goclaw-runtime/config.json` 是否存在
2. 如果不存在，自动运行 `picoclaw-web onboard` 命令
3. 初始化成功后继续启动 Launcher

```javascript
if (!fs.existsSync(configPath)) {
  const { execSync } = require('child_process');
  execSync(`"${exePath}" onboard`, {
    cwd: workDir,
    env: {
      PICOCLAW_HOME: configDir,
      PICOCLAW_CONFIG: configPath
    },
    timeout: 10000
  });
}
```

### 启动轮询

**轮询函数**：`pollStartupProgress()` (第 854-939 行)

**检查项目**：
1. **后端服务** (18790)：检查 `/health` 端点
2. **Gateway 状态**：检查 `/pet/token` 端点
3. **前端面板** (18800/3000)：检查根路径
4. **桌宠渲染**：检查 `/desktop-pet` 路径

**超时保护**：
- **45 秒强制完成**：超过 45 秒后强制完成启动，避免无限等待
- **20 秒后端警告**：后端超过 20 秒未就绪时显示警告

```javascript
if (startupAttemptCount >= STARTUP_MAX_ATTEMPTS && !startupForceCompleted) {
  startupForceCompleted = true;
  // 设置警告状态
  completeStartupAndShowDesktop({ openPanel: panelReady && openPanelOnReady });
  return;
}
```

### 生产模式适配

**文件位置**：`clawpet-frontend/clawpet/electron/main.js` (第 205-214 行)

在打包后的生产环境中，没有独立的 Next.js 开发服务器（3000 端口），因此：

```javascript
const isProduction = app.isPackaged;
let effectiveDashboardBaseUrl = dashboardBaseUrl;

if (isProduction && !process.env.GOCLAW_DASHBOARD_URL) {
  // 生产模式：使用 Launcher URL (18800) 替代 Dashboard URL (3000)
  effectiveDashboardBaseUrl = launcherBaseUrl;
}
```

---

## 四、GitHub Actions Release 流程

### 触发条件

**文件位置**：`.github/workflows/release.yml`

当推送 Git Tag 时自动触发：

```yaml
on:
  push:
    tags:
      - 'v*'
      - 'clawpet-*'
```

### 构建矩阵

支持多平台构建：

| 操作系统 | 架构 |
|---------|------|
| Linux | amd64, arm64, arm, riscv64, loong64 |
| Windows | amd64, arm64, arm, riscv64, loong64 |
| macOS | amd64, arm64 |
| FreeBSD | amd64, arm64, arm, riscv64, loong64 |

### 构建步骤

#### 1. Go 后端构建

使用 GoReleaser 或手动构建：

```bash
GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} GOARM="${GOARM}" CGO_ENABLED=0 \
  go build \
    -trimpath \
    -tags "${BUILD_TAGS}" \
    -ldflags "-s -w" \
    -o "dist/clawpet" \
    ./cmd/picoclaw
```

#### 2. Launcher 前端构建（所有平台）

```yaml
- name: Build launcher embedded frontend
  shell: bash
  run: |
    cd web/frontend
    pnpm install --no-frozen-lockfile
    pnpm build:backend
```

**重要**：已移除 Windows-only 限制，所有平台都会构建 Launcher 前端。

#### 3. Launcher 二进制构建（所有平台）

```bash
GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} GOARM="${GOARM}" CGO_ENABLED=0 \
  go build \
    -trimpath \
    -tags "${BUILD_TAGS}" \
    -ldflags "-s -w" \
    -o "dist/picoclaw-web" \
    ./web/backend
```

**重要**：已移除 Windows-only 限制，所有平台都会构建 Launcher。

#### 4. 打包发布

**Windows**：
```bash
# 使用 Inno Setup 打包
iscc scripts/setup.iss
```

**Linux/macOS/FreeBSD**：
```bash
# 打包为 tar.gz（包含 Gateway + Launcher + 启动脚本）
tar -czf "${BASE}.tar.gz" "${BIN}" "${LAUNCHER_BIN}" "${BIN}-gateway"
```

---

## 五、关键修复记录

### 1. 启动卡死问题（75% 位置）

**问题**：打包后的 Electron 应用卡在启动界面 75%，面板和渲染服务无法完成。

**原因**：
- `build-and-test.ps1` 缺少 `npm run build` 步骤
- `.next` 目录未生成，前端静态资源缺失

**修复**：添加 Step 4 构建 Next.js 前端

### 2. 生产模式面板 URL 错误

**问题**：生产模式下仍等待 3000 端口，但实际没有 Next.js 服务器。

**原因**：
- `pollStartupProgress()` 只在 `rendererReady` 时退出，无超时机制
- 生产模式未检测 `app.isPackaged`

**修复**：
1. 添加生产模式检测
2. 自动将 Dashboard URL 降级到 Launcher URL（18800）
3. 添加 45 秒超时保护

### 3. Release 平台限制

**问题**：Launcher 前端和二进制只在 Windows 平台构建。

**原因**：`release.yml` 中的 `if: matrix.goos == 'windows'` 限制。

**修复**：
- 移除 Launcher 前端构建的 Windows-only 限制
- 移除 Launcher 二进制构建的 Windows-only 限制
- 非 Windows 平台打包包含 Launcher 二进制

---

## 六、常见问题排查

### 日志位置

所有日志输出到：`~/.goclaw/logs.txt`

查看日志：
```powershell
Get-Content "$env:USERPROFILE\.goclaw\logs.txt" -Tail 50
```

### 配置目录

- **运行时配置**：`~/.goclaw-runtime/config.json`
- **用户数据**：`~/.goclaw/`

### 端口配置

| 服务 | 端口 | 功能 |
|------|------|------|
| Gateway | 18790 | AI 代理和 WebSocket 网关 |
| Launcher | 18800 | 前端静态资源和 API |
| Next.js 开发 | 3000 | 仅开发模式使用 |

### 环境变量

```bash
# 后端服务地址
GOCLAW_BACKEND_URL=http://127.0.0.1:18790

# 前端面板地址
GOCLAW_DASHBOARD_URL=http://127.0.0.1:3000  # 开发模式
GOCLAW_LAUNCHER_URL=http://127.0.0.1:18800  # 生产模式

# Launcher 认证令牌
GOCLAW_LAUNCHER_TOKEN=goclaw-local-token

# 桌宠渲染路径
GOCLAW_PET_RENDERER_PATH=/desktop-pet

# 显示启动进度窗口（默认显示）
GOCLAW_SHOW_STARTUP=1

# 就绪后打开面板（默认打开）
GOCLAW_OPEN_PANEL_ON_READY=1

# 打开开发者工具（仅调试）
ELECTRON_OPEN_DEVTOOLS=1
```

---

## 七、窗口尺寸配置参考

### 当前配置

| 窗口类型 | 宽度 | 高度 | 说明 |
|---------|------|------|------|
| 设置窗口 | 屏幕宽度 × 0.7 | 屏幕高度 × 0.667 | 7/10 宽度，2/3 高度 |
| 桌宠窗口 | 280px | 380px | 固定尺寸 |
| 启动窗口 | 860px | 560px | 固定尺寸 |

### 修改窗口尺寸

**设置窗口**：编辑 `electron/main.js` 第 648-651 行

```javascript
const settingsWidth = Math.round(width * 0.7);     // 修改宽度比例
const settingsHeight = Math.round(height * 0.667); // 修改高度比例
```

**桌宠窗口**：编辑 `electron/main.js` 第 42-44 行

```javascript
const PET_WIDTH = 280;   // 修改宽度
const PET_HEIGHT = 380;  // 修改高度
```

---

## 八、注意事项

1. **Next.js 构建**：打包前必须执行 `npm run build`，否则前端资源缺失
2. **端口冲突**：确保 18790 和 18800 端口未被占用
3. **配置文件**：首次启动会自动运行 onboard 初始化配置
4. **超时保护**：45 秒超时后强制完成启动，避免无限等待
5. **生产模式**：打包后使用 Launcher (18800) 提供前端，不再依赖 Next.js 开发服务器

---

**文档版本**：v1.0  
**最后更新**：2026-04-26  
**适用版本**：ClawPet 0.1.0+
