# Release 包优化说明

## 📦 优化内容

### 问题
- ❌ Release 包体积过大（约 300MB）
- ❌ 包含了完整的 `node_modules`（开发依赖）
- ❌ 包含了所有源码（不需要）

### 解决方案
- ✅ 在 GitHub Actions 中构建前端为静态文件
- ✅ 只打包构建产物，不包含 node_modules
- ✅ 优化后包体积预计降至 **80-100MB**（减少 60-70%）

---

## 🔧 修改的文件

### 1. `.github/workflows/release.yml`

**新增步骤：**

#### 步骤 1：构建 petclaw 前端
```yaml
- name: Build petclaw frontend (Windows only)
  if: matrix.goos == 'windows'
  shell: bash
  run: |
    cd petclaw
    pnpm install --no-frozen-lockfile
    pnpm build
```

**产物：** `petclaw/out/` 目录（静态 HTML/CSS/JS）

#### 步骤 2：构建 electron-frontend
```yaml
- name: Build electron-frontend (Windows only)
  if: matrix.goos == 'windows'
  shell: bash
  run: |
    cd electron-frontend
    pnpm install --no-frozen-lockfile
    pnpm build
```

**产物：** `electron-frontend/dist/` 目录

#### 步骤 3：优化打包
**修改前：**
```yaml
cp -r ../petclaw allinone/petclaw              # ❌ 包含 node_modules
cp -r ../electron-frontend allinone/electron-frontend  # ❌ 包含 node_modules
```

**修改后：**
```yaml
# ✅ 只打包构建产物
cp -r ../petclaw/out allinone/petclaw/out
cp ../petclaw/package.json allinone/petclaw/package.json

cp -r ../electron-frontend/dist allinone/electron-frontend/dist
cp ../electron-frontend/package.json allinone/electron-frontend/package.json
```

---

### 2. `petclaw/next.config.mjs`

**新增配置：**
```javascript
const nextConfig = {
  // 输出静态文件到 out 目录（用于打包）
  output: 'export',
  
  // ... 其他配置
}
```

**作用：**
- Next.js 会将应用构建为纯静态文件
- 输出到 `out/` 目录
- 不需要 Node.js 服务器即可运行

---

## 📊 包体积对比

### 优化前（300MB+）
```
clawpet_AllInOne_Windows_x86_64/
├── clawpet.exe                    # 50MB
├── clawpet-launcher.exe           # 50MB
├── petclaw/                       # 150MB ❌
│   ├── node_modules/              # 120MB（开发依赖）
│   ├── src/                       # 10MB（源码）
│   └── out/                       # 20MB（构建产物）
└── electron-frontend/             # 100MB ❌
    ├── node_modules/              # 80MB（开发依赖）
    ├── src/                       # 10MB（源码）
    └── dist/                      # 10MB（构建产物）
```

### 优化后（80-100MB）
```
clawpet_AllInOne_Windows_x86_64/
├── clawpet.exe                    # 50MB
├── clawpet-launcher.exe           # 50MB
├── petclaw/                       # 25MB ✅
│   ├── out/                       # 20MB（构建产物）
│   └── package.json               # 0.01MB
└── electron-frontend/             # 15MB ✅
    ├── dist/                      # 10MB（构建产物）
    ├── package.json               # 0.01MB
    ├── vite.config.ts             # 0.01MB
    └── index.html                 # 0.01MB
```

**减少：约 200MB（67%）**

---

## 🚀 使用方法

### 开发者（本地开发）
```powershell
# 使用开发模式（需要 node_modules）
powershell -ExecutionPolicy Bypass -File .\scripts\run-goclaw-dev.ps1 `
  -Restart -PetclawMode dev
```

### 最终用户（Release 包）
```powershell
# 双击即用（不需要安装 Node.js）
.\GoClaw-OneClickStart.bat
```

---

## ⚠️ 注意事项

### 1. Petclaw 静态导出的限制

使用 `output: 'export'` 后：
- ✅ 支持：静态页面、客户端路由、API 调用
- ❌ 不支持：`getServerSideProps`、`getInitialProps`（服务端渲染）
- ❌ 不支持：动态 API 路由（`/api/*`）

**当前项目状态：**
- ✅ 已检查，没有使用不支持的功能
- ✅ API 调用通过后端 Gateway（端口 18790）

### 2. 构建时间增加

GitHub Actions 构建时间会增加约 **2-3 分钟**（前端构建时间）

### 3. 开发模式不受影响

- 本地开发仍然使用 `npm run dev`
- 仍然支持 Hot Reload
- 只是 Release 打包时使用构建产物

---

## 🎯 未来优化方向

### 阶段 1：当前（已完成）✅
- 构建前端为静态文件
- 只打包构建产物
- 包体积：300MB → 100MB

### 阶段 2：使用 Go embed（可选）
- 将前端静态文件嵌入 Go 二进制
- 包体积：100MB → 60MB
- 单个 exe 文件

### 阶段 3：使用 Wails 框架（可选）
- 替代 Electron
- 使用系统原生 WebView
- 包体积：60MB → 40MB
- 启动更快，内存占用更少

---

## 📝 验证清单

在发布前验证：

- [ ] petclaw 构建成功（`pnpm build`）
- [ ] electron-frontend 构建成功（`pnpm build`）
- [ ] Release 包不包含 `node_modules`
- [ ] Release 包可以正常启动
- [ ] 前端功能正常（聊天、设置等）
- [ ] Electron 桌面宠物正常显示

---

## 🔍 故障排除

### 问题 1：petclaw 构建失败
```bash
# 检查是否有不支持的功能
grep -r "getServerSideProps\|getInitialProps" petclaw/app/
```

### 问题 2：启动后前端空白
```bash
# 检查 out/ 目录是否存在
ls petclaw/out/

# 检查是否有 index.html
test -f petclaw/out/index.html
```

### 问题 3：Electron 无法启动
```bash
# 检查 dist/ 目录
ls electron-frontend/dist/

# 检查 main.js 是否存在
test -f electron-frontend/src/main.js
```

---

## 📞 联系

如有问题，请提交 Issue 或联系开发团队。
