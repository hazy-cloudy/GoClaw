# ClawPet 引导页数据结构与接口说明

本文档说明当前米色 `petclaw` 引导页（onboarding）采集的数据结构、前后端接口、持久化位置，以及和运行时配置的映射关系。

## 1. 设计目标

引导页采集的信息是重要用户输入，不能只停留在前端 `localStorage`。

当前目标：

1. 明确一份稳定的数据结构
2. 提供正式的前后端提交接口
3. 在后端持久化保存
4. 将其中一部分字段同步到 `pet` 当前运行配置

## 2. 入口页面

- 页面地址：`/onboarding`
- 当前主前端：`petclaw`
- 推荐入口：`http://127.0.0.1:3000/onboarding?mode=rerun`

## 3. 数据结构

当前提交到后端的完整 onboarding 快照结构如下：

```json
{
  "version": 1,
  "completed": true,
  "completed_at": "2026-04-19T12:00:00Z",
  "profile": {
    "display_name": "义父",
    "role": "计算机",
    "language": "中文"
  },
  "pet": {
    "personality_tone": "阴阳怪气",
    "activity_level": 65,
    "nickname": "义父",
    "custom_nickname": "",
    "final_nickname": "义父",
    "voice_style": "活泼碎碎念"
  },
  "app_preferences": {
    "auto_connect_on_launch": true,
    "enable_desktop_bubble": true,
    "open_console_on_pet_click": true
  },
  "permissions": {
    "screen_view": true,
    "app_check": true,
    "local_doc_read": true,
    "popup_reminder": true
  },
  "study_preferences": {
    "sleep_hour": 1,
    "anxiety_level": 58,
    "selected_breakers": ["早八", "论文"],
    "ics_file_name": "schedule.ics"
  },
  "student_insights": {
    "learning_rhythm": {
      "chronotype": "night",
      "focus_windows": ["10:00-12:00", "15:00-17:00"],
      "quiet_windows": ["23:30-07:30 默认静默"],
      "reminder_cadence": "standard",
      "summary": "..."
    },
    "pressure_plan": {
      "level": "medium",
      "strategy": "...",
      "reminder_intervals_minutes": [480, 240, 90],
      "tone_guide": "...",
      "template_soft": "...",
      "template_normal": "...",
      "template_strong": "..."
    }
  }
}
```

## 4. 接口

### 4.1 launcher 入口

- Method: `POST`
- Path: `/api/pet/onboarding`
- Host: `127.0.0.1:18800`

说明：

- 这是 `petclaw` 前端当前使用的正式提交入口
- 如果新路径不可用，前端会回退尝试 `/api/pico/onboarding`

### 4.2 gateway 直连入口

- Method: `POST`
- Path: `/pet/onboarding`
- Host: `127.0.0.1:18790`

说明：

- launcher 会把 `/api/pet/onboarding` 转发到这里
- 由 `pkg/channels/pet` 直接处理

### 4.3 成功响应

```json
{
  "status": "ok",
  "data": {
    "pet_id": "pet_001",
    "language": "中文",
    "reminder_enabled": true,
    "voice_enabled": true
  }
}
```

## 5. 后端持久化

当前 onboarding 快照会单独保存为：

```text
<workspace>/pet_onboarding.json
```

这份文件用于保存引导页采集的完整信息，不和现有 `pet_config.json` 强耦合，便于后续扩展。

## 6. 运行时映射

当前后端会把 onboarding 快照中的一部分字段同步到运行时配置：

- `profile.language` -> `app.language`
- `app_preferences.enable_desktop_bubble` + `permissions.popup_reminder` -> `app.reminder_enabled`
- `app_preferences.enable_desktop_bubble` + `permissions.popup_reminder` -> `app.voice_enabled`
- `student_insights.pressure_plan.reminder_intervals_minutes[0]` -> `app.proactive_interval_minutes`
- `pet.personality_tone` / `pet.activity_level` / `pet.final_nickname` -> 当前角色的 `persona` / `persona_type`

## 7. 前端本地保存

除了正式接口提交，前端仍然会保存一份本地状态：

- key: `petclaw.onboarding.v1`

用途：

- 页面重新进入时可复用部分前端状态
- UI 动画 / 主题切换可以直接读取本地快照

但这不再是唯一数据来源。

## 8. 当前限制

当前这套 onboarding 对接已经满足“正式结构 + 正式接口 + 后端持久化”，但仍有几个限制：

1. 引导页加载时还没有从后端反向拉取这份快照进行预填
2. 当前主要是保存与运行时映射，不是完整的“后台资料管理系统”
3. 一些字段目前只做保存，还没有全部参与运行逻辑

## 9. 相关代码

- 前端引导页：`petclaw/components/onboarding-wizard.tsx`
- 前端本地结构：`petclaw/lib/onboarding.ts`
- launcher API：`web/backend/api/pico.go`
- gateway 入口：`pkg/channels/pet/channel.go`
- 后端保存与映射：`pkg/pet/onboarding_service.go`
- onboarding 快照定义：`pkg/pet/config/onboarding.go`
