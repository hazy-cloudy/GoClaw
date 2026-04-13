# GoClaw - 桌面宠物 Go 版本

基于 Wails + React 的智能桌面宠物助手。

## 技术栈

- **后端**: Go + Wails
- **前端**: React + TypeScript
- **AI**: PicoClaw
- **TTS**: edge-tts

## 项目结构

```
go-claw/
├── main.go          # 入口 + Wails 配置
├── app.go           # 主应用
├── tools.go         # 工具系统
├── tts.go           # TTS 语音合成
├── frontend/        # React 前端
│   ├── src/
│   │   ├── App.tsx  # 主组件
│   │   ├── App.css  # 样式
│   │   ├── usePicoClaw.ts  # WebSocket 连接
│   │   └── wails.ts # Wails 绑定
│   └── package.json
├── wails.json       # Wails 配置
└── go.mod
```

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+
- Wails CLI

```bash
# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装前端依赖
cd frontend && npm install

# 运行开发模式
wails dev
```

## 功能

- [x] 桌宠显示和动画
- [x] 气泡系统（优先级）
- [x] PicoClaw AI 对话
- [x] 情感识别和回复
- [x] 工具调用（提醒、搜索等）
- [x] 记忆系统
- [ ] TTS 语音
- [ ] 设置界面

## License

MIT
