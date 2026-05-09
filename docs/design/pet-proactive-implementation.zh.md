# 桌宠主动性系统实现设计（第一期）

## 1. 文档定位

本文档属于**实现设计**，重点描述：

- 模块如何拆分
- 数据结构如何定义
- `weekly_report` 和 `progress_nudge` 如何落地
- “用户是否可打扰”这一能力如何实现
- 状态、规则、push、前后端接入点如何设计

本文档建立在《技术提案》和《架构设计》的前提之上。

---

## 2. 本期实现范围

### 包含

- 主动性统一 manager
- 活动记录存储
- 状态快照构建
- 可打扰判断层
- 基础 policy 规则
- `weekly_report`
- `progress_nudge`
- delivery history / cooldown
- push 协议与后端定向投递

### 不包含

- 月报
- 趋势图
- 行为学习
- 多事件复杂竞争排序
- 完整历史事件中心

---

## 3. 模块拆分

建议新增目录：

```text
pkg/pet/activity/
  types.go
  store.go
  collector.go

pkg/pet/proactive/
  types.go
  manager.go
  snapshot.go
  interruptibility.go
  policy.go
  delivery.go
  history.go
  provider.go

pkg/pet/report/
  weekly.go
  aggregator.go
  renderer.go
```

现有文件建议扩展：

- [pkg/pet/service.go](/D:/study%20part/GoClawPet/pkg/pet/service.go)
- [pkg/pet/types.go](/D:/study%20part/GoClawPet/pkg/pet/types.go)
- [pkg/pet/config/types.go](/D:/study%20part/GoClawPet/pkg/pet/config/types.go)
- [pkg/pet/hooks.go](/D:/study%20part/GoClawPet/pkg/pet/hooks.go)

---

## 4. 活动记录设计

### 4.1 目标

主动事件不能只依赖聊天文本回放。

第一期要引入一个结构化活动记录层，用于回答：

- 用户最近让桌宠干了什么
- 哪些事已经完成
- 哪些事还没收口

### 4.2 数据结构

建议定义：

```go
type ActivityEventType string

const (
    ActivityUserMessage ActivityEventType = "user_message"
    ActivityToolCall    ActivityEventType = "tool_call"
    ActivityToolResult  ActivityEventType = "tool_result"
    ActivityFileOutput  ActivityEventType = "file_output"
    ActivityTaskResult  ActivityEventType = "task_result"
)

type ActivityCategory string

const (
    CategoryCode   ActivityCategory = "code"
    CategoryDoc    ActivityCategory = "doc"
    CategoryDebug  ActivityCategory = "debug"
    CategoryPPTX   ActivityCategory = "pptx"
    CategoryConfig ActivityCategory = "config"
    CategoryOther  ActivityCategory = "other"
)

type ActivityStatus string

const (
    ActivityPending ActivityStatus = "pending"
    ActivityDone    ActivityStatus = "done"
    ActivityFailed  ActivityStatus = "failed"
)

type ActivityEvent struct {
    ID          string            `json:"id"`
    CharacterID string            `json:"character_id"`
    SessionID   string            `json:"session_id"`
    Type        ActivityEventType `json:"type"`
    Category    ActivityCategory  `json:"category"`
    Status      ActivityStatus    `json:"status"`
    Title       string            `json:"title"`
    Summary     string            `json:"summary"`
    ToolName    string            `json:"tool_name,omitempty"`
    FilePaths   []string          `json:"file_paths,omitempty"`
    Meta        map[string]any    `json:"meta,omitempty"`
    CreatedAt   time.Time         `json:"created_at"`
}
```

### 4.3 存储方式

第一期建议使用：

- `workspace/pet_activity/YYYY-MM.jsonl`

原因：

- 轻量
- 可读
- 易排查
- 与现有项目会话 / 日志风格一致

### 4.4 采集点

第一期建议三个采集点：

#### A. 用户发起请求时

在桌宠聊天入口记录一条 `user_message`：

- 原始请求
- 自动分类
- 摘要标题

#### B. Tool 执行时

记录：

- `tool_call`
- `tool_result`
- 成功 / 失败

#### C. 产出文件时

记录：

- `file_output`
- 输出路径
- 对应任务摘要

#### D. 一次性定时任务写入时

对于 `cron add` 产生的单次 `at` 任务，额外写入一条 `task_result/pending` 到 activity：

- `schedule_kind = "at"`
- `at_ms`
- `due_at_ms`
- `due_at`
- `job_id`
- `job_name`

这样主动提醒可以真正基于“任务最后时间点”做判断，而不是只靠聊天文本。

---

## 5. Snapshot 设计

### 5.1 目标

主动判断不应该到处现查状态，而应该统一构建快照。

### 5.2 数据结构

建议：

```go
type ProactiveSnapshot struct {
    Now time.Time `json:"now"`

    Pet struct {
        CharacterID     string `json:"character_id"`
        PersonaType     string `json:"persona_type"`
        PersonalityTone string `json:"personality_tone"`
        DominantEmotion string `json:"dominant_emotion"`
        EmotionScore    int    `json:"emotion_score"`
    } `json:"pet"`

    User struct {
        DisplayName     string `json:"display_name"`
        Chronotype      string `json:"chronotype"`
        PersonalityTone string `json:"personality_tone"`
        PressureLevel   string `json:"pressure_level"`
        CurrentMood     string `json:"current_mood"`
        EnergyLevel     int    `json:"energy_level"`
        EngagementLevel int    `json:"engagement_level"`
        StressTrend     string `json:"stress_trend"`
    } `json:"user"`

    Activity struct {
        LastUserMessageAt   time.Time `json:"last_user_message_at"`
        LastPushAt          time.Time `json:"last_push_at"`
        RecentMessageCount  int       `json:"recent_message_count"`
        RecentTaskCount     int       `json:"recent_task_count"`
        UnfinishedTaskCount int       `json:"unfinished_task_count"`
        CurrentSessionBusy  bool      `json:"current_session_busy"`
        ConsoleVisible      bool      `json:"console_visible"`
        PetVisible          bool      `json:"pet_visible"`
    } `json:"activity"`

    Preferences struct {
        ProactiveCare         bool `json:"proactive_care"`
        ProactiveIntervalMins int  `json:"proactive_interval_minutes"`
        WeeklyReportEnabled   bool `json:"weekly_report_enabled"`
        ProgressNudgeEnabled  bool `json:"progress_nudge_enabled"`
        GlobalCooldownMins    int  `json:"global_cooldown_minutes"`
    } `json:"preferences"`
}
```

### 5.3 来源映射

| Snapshot 字段 | 来源 |
|---|---|
| `Pet.*` | `characters.Manager` + `emotion.Engine` |
| `User.*` | `userprofile.Manager` |
| `Activity.*` | `activity.Store` + 当前前端会话状态 |
| `Preferences.*` | `config.Manager` |

---

## 6. Interruptibility 设计

### 6.1 目标

将“用户是否可打扰”抽象为独立判断层，而不是散落在各个事件里。

它回答的问题不是：

- 这个事件值不值得发

而是：

- 这个用户当前时刻适不适合被打扰

### 6.2 数据结构

建议定义：

```go
type InterruptibilityLevel string

const (
    InterruptibilityNo   InterruptibilityLevel = "no"
    InterruptibilitySoft InterruptibilityLevel = "soft"
    InterruptibilityYes  InterruptibilityLevel = "yes"
)

type InterruptibilityDecision struct {
    Interruptible bool                  `json:"interruptible"`
    Level         InterruptibilityLevel `json:"level"`
    ReasonCodes   []string              `json:"reason_codes"`
    NextCheckAfter time.Duration        `json:"next_check_after"`
}
```

### 6.3 建议规则

#### 不可打扰

- 当前会话 busy
- 最近 2 分钟内用户连续发言
- 最近 5 分钟内刚收到主动事件
- 当前命中静默时段

#### 软可打扰

- 用户刚打开面板
- 用户刚发送一条消息，但当前不忙
- 用户刚完成一个任务

#### 强可打扰

- 用户当前有活跃使用，但不在连续忙碌
- 用户重新回到控制台
- 当前存在高价值待投递事件，例如 ready 的周报

### 6.4 触发实现方式

这里建议采用**混合模型**：

#### 后台定时检查

作用：

- 评估哪些事件进入 `ready`
- 更新长期状态

#### 使用时即时检查

作用：

- 当用户出现使用行为时，重新计算 interruptibility
- 决定是否把 ready 事件真正投递

### 6.5 即时触发点

建议第一期支持：

- 用户发消息
- 聊天面板打开
- 桌宠窗口被激活
- 一个任务刚完成

这些入口不直接生成事件，而是触发一次：

- “重新评估当前 ready 事件是否可投递”

---

## 7. Policy 设计

### 7.1 目标

Policy 只负责决定：

- 当前是否允许某类事件进入投递候选
- 当前允许多强的触达等级

### 7.2 数据结构

```go
type DeliveryLevel string

const (
    DeliveryBlocked DeliveryLevel = "blocked"
    DeliveryBubble  DeliveryLevel = "bubble"
    DeliveryCard    DeliveryLevel = "card"
)

type PolicyDecision struct {
    Allowed       bool          `json:"allowed"`
    DeliveryLevel DeliveryLevel `json:"delivery_level"`
    Score         int           `json:"score"`
    ReasonCodes   []string      `json:"reason_codes"`
}
```

### 7.3 基础规则

#### 硬阻断

- `proactive_care == false`
- 距离上次主动提醒小于全局冷却
- `interruptibility == no`

#### 加分项

- 最近 6 小时活跃很低：`+10`
- 有未完成事项：`+20`
- 用户压力高：`+20`
- 用户低能量：`+10`
- 桌宠 fear 高：`+5`
- 桌宠 joy 高：`+5`

#### 结果映射

- `< 20` -> blocked
- `20-39` -> bubble
- `>= 40` -> card

### 7.4 说明

- `weekly_report` 仍要受 policy 限制
- `progress_nudge` 对 cooldown 和 busy 状态更敏感

---

## 8. Provider 接口

建议统一：

```go
type ProactiveIntent struct {
    Type        string         `json:"type"`
    Priority    string         `json:"priority"`
    ReasonCodes []string       `json:"reason_codes"`
    Payload     map[string]any `json:"payload"`
}

type EventProvider interface {
    Name() string
    Evaluate(snapshot ProactiveSnapshot) (*ProactiveIntent, bool, error)
}
```

语义：

- `(*intent, true, nil)`：可触发
- `(_, false, nil)`：本轮不触发
- `(_, _, err)`：provider 异常

---

## 9. weekly_report 实现设计

### 9.1 触发条件

- `weekly_report_enabled == true`
- 当前时间进入周报窗口
- 本周有效任务数 >= 3
- 本周至少有一次用户主动消息
- 本周未发送过周报

### 9.2 两阶段触发模型

`weekly_report` 不建议设计成“时间一到立即投递”。

建议拆成两个阶段：

#### 阶段 A：Ready

后台定期检查：

- 是否进入周报窗口
- 是否有足够活动
- 是否尚未生成本周周报

满足后，将本周周报状态置为：

- `Ready = true`

但先不立即推送。

#### 阶段 B：Delivered

当用户出现使用事件时，触发 interruptibility 评估。

如果：

- `Ready == true`
- `Delivered == false`
- 当前用户可打扰

则真正投递周报。

### 9.3 周报状态结构

```go
type WeeklyReportState struct {
    WeekKey      string     `json:"week_key"`
    ReportID     string     `json:"report_id"`
    Ready        bool       `json:"ready"`
    GeneratedAt  *time.Time `json:"generated_at,omitempty"`
    DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
    ExpireAt     *time.Time `json:"expire_at,omitempty"`
}
```

### 9.4 数据结构

```go
type CategoryStat struct {
    Name  string `json:"name"`
    Count int    `json:"count"`
}

type WeeklyReport struct {
    ReportID      string         `json:"report_id"`
    CharacterID   string         `json:"character_id"`
    PeriodStart   time.Time      `json:"period_start"`
    PeriodEnd     time.Time      `json:"period_end"`
    MessageCount  int            `json:"message_count"`
    TaskCount     int            `json:"task_count"`
    ToolCallCount int            `json:"tool_call_count"`
    FailureCount  int            `json:"failure_count"`
    TopCategories []CategoryStat `json:"top_categories"`
    Outputs       []string       `json:"outputs"`
    Unfinished    []string       `json:"unfinished"`
    PeakDay       string         `json:"peak_day"`
    PeakHour      int            `json:"peak_hour"`
    Summary       string         `json:"summary"`
}
```

### 9.5 聚合规则

- `message_count`
  - `user_message` 数量
- `task_count`
  - `task_result` 数量
- `tool_call_count`
  - `tool_call` 数量
- `failure_count`
  - `status == failed`
- `top_categories`
  - 按 `category` 聚合排序取 Top 3
- `outputs`
  - 从 `file_output` 和 `done task_result` 中提取
- `unfinished`
  - 从 `pending/failed` 中提取

### 9.6 渲染策略

第一期建议：

- 规则模板渲染为主
- LLM 润色为可选增强

输入：

- `WeeklyReport`
- `persona_type`
- `personality_tone`
- `dominant_emotion`

输出：

- 周报标题
- 周报摘要
- 关键块内容

### 9.7 Intent

```go
{
  "type": "weekly_report",
  "priority": "medium",
  "reason_codes": ["weekly_window", "enough_activity"],
  "payload": {
    "report_id": "...",
    "title": "本周陪跑回顾",
    "summary": "...",
    "report": { ... }
  }
}
```

---

## 10. progress_nudge 实现设计

### 10.1 触发条件

- `progress_nudge_enabled == true`
- 存在未完成事项
- 当前事项属于“一次性 deadline 型任务”，而不是 `every/cron` 周期任务
- 距离上次同事项提醒超过冷却
- 当前用户可打扰
- 最近没有刚收到其他主动事件

### 10.2 未完成事项识别

第一期使用规则法：

- `task_result` with `pending`
- `task_result` with `failed`
- 同主题在 24 小时内反复出现但没有 `done`
- `Meta.schedule_kind in {"at"}` 的任务会进入主动提醒池
- `Meta.schedule_kind in {"every","cron"}` 的任务不进入主动提醒池

### 10.3 首次提醒时机

不直接按“创建了多久”提醒，而是尽量按任务的目标时间距离决定首次提醒。

优先读取：

- `Meta.due_at`
- `Meta.due_at_ms`
- `Meta.at_ms`

如果没有 deadline 信息，才退回到 `CreatedAt` 做近似判断。

首次提醒延迟的第一版规则：

- 距离 deadline `>= 30 天`：首次提醒在 `7 天前`
- 距离 deadline `>= 7 天`：首次提醒在 `1 天前`
- 距离 deadline `>= 1 天`：首次提醒在 `6 小时前`
- 距离 deadline `>= 6 小时`：首次提醒在 `90 分钟前`
- 更近的任务：首次提醒在 `30 分钟前`

### 10.4 提醒次数与冷却

每个事件最多 3 次提醒：

1. 首次提醒：按时间距离规则触发
2. 第二次提醒：基于首次延迟的指数退避
3. 第三次提醒：在任务最后时间点兜底再提醒一次

如果用户在第一次提醒之后、deadline 之前有新的消息交互，则第二次中间提醒直接跳过，只保留最后一次 deadline 提醒。

### 10.5 数据结构

```go
type ProgressNudge struct {
    NudgeID       string    `json:"nudge_id"`
    CharacterID   string    `json:"character_id"`
    Topic         string    `json:"topic"`
    Reason        string    `json:"reason"`
    Suggestion    string    `json:"suggestion"`
    LastActiveAt  time.Time `json:"last_active_at"`
    UnfinishedCnt int       `json:"unfinished_count"`
}
```

### 10.6 文案策略

只输出两层信息：

- 一句提醒
- 一句建议

避免长篇说教。

### 10.7 Intent

```go
{
  "type": "progress_nudge",
  "priority": "low",
  "reason_codes": ["unfinished_tasks", "time_distance_ready"],
  "payload": {
    "nudge_id": "...",
    "event_id": "...",
    "topic": "明天讲 PPT",
    "summary": "明天讲 PPT 这件事还没有收口。",
    "suggestion": "要不要先把这件事收一下？",
    "reminder_count": 1
  }
}
```

---

## 11. Delivery 设计

### 11.1 PushType 扩展

在 [pkg/pet/types.go](/D:/study%20part/GoClawPet/pkg/pet/types.go) 中建议新增：

```go
const (
    PushTypeWeeklyReport  = "weekly_report_ready"
    PushTypeProgressNudge = "progress_nudge"
)
```

### 11.2 Push Payload

```go
type WeeklyReportPush struct {
    Title    string       `json:"title"`
    Summary  string       `json:"summary"`
    ReportID string       `json:"report_id"`
    Report   WeeklyReport `json:"report"`
}

type ProgressNudgePush struct {
    Topic      string `json:"topic"`
    Summary    string `json:"summary"`
    Suggestion string `json:"suggestion"`
    NudgeID    string `json:"nudge_id"`
}
```

### 11.3 UI 投递策略

- `weekly_report`
  - 优先聊天卡片
  - 同时可配桌宠气泡摘要
- `progress_nudge`
  - 优先桌宠气泡
  - 必要时再升级为卡片

---

## 12. Delivery History 与冷却

### 12.1 目的

主动事件必须有投递历史，否则无法可靠防重。

### 12.2 数据结构

```go
type DeliveryHistoryRecord struct {
    EventType   string    `json:"event_type"`
    EventID     string    `json:"event_id"`
    CharacterID string    `json:"character_id"`
    DeliveredAt time.Time `json:"delivered_at"`
}
```

### 12.3 存储

建议：

- `workspace/pet_proactive_history.json`

### 12.4 冷却规则

- 全局冷却：90 分钟
- `weekly_report`：每周一次
- `progress_nudge`：12 小时一次

---

## 13. 配置扩展

建议在 `AppConfig` 中增加：

```go
WeeklyReportEnabled   bool `json:"weekly_report_enabled"`
ProgressNudgeEnabled  bool `json:"progress_nudge_enabled"`
ProactiveCheckMinutes int  `json:"proactive_check_minutes"`
GlobalCooldownMinutes int  `json:"global_cooldown_minutes"`
```

默认值建议：

```go
WeeklyReportEnabled:   true
ProgressNudgeEnabled:  true
ProactiveCheckMinutes: 30
GlobalCooldownMinutes: 90
```

---

## 14. 服务启动流程

在 [pkg/pet/service.go](/D:/study%20part/GoClawPet/pkg/pet/service.go) 中建议：

### 新增成员

```go
activityStore    *activity.Store
proactiveManager *proactive.Manager
```

### 初始化

在 `NewPetService()` 中创建：

- `activityStore`
- `proactiveManager`

### 启动

在 `Start()` 中新增：

```go
go s.runEmotionDecay()
go s.proactiveManager.Start()
```

### 即时触发入口

除了后台定时检查，还建议在以下场景下调用：

- 用户发消息后
- 控制台面板打开后
- 桌宠窗口被激活后
- 任务完成后

这些入口不负责直接生成事件，只负责通知 proactive manager：

- “重新评估当前 ready 事件是否可投递”

---

## 15. 前端接入

### WebSocket 事件

前端需要识别两个新的 push type：

- `weekly_report_ready`
- `progress_nudge`

### 展示方式

#### weekly_report

- 气泡摘要
- 聊天区周报卡片

#### progress_nudge

- 优先气泡
- 点击后可展开到聊天区

---

## 16. 开发顺序

建议按以下顺序落地：

1. 定义 `activity` / `proactive` / `report` 基础类型
2. 接入 activity 记录
3. 实现 snapshot builder
4. 实现 interruptibility evaluator
5. 实现 history 与 cooldown
6. 实现 `weekly_report` 的 ready / delivered 状态
7. 实现 `progress_nudge`
8. 扩展 push types
9. 前端接收与展示

---

## 17. 测试建议

### 单元测试

- 活动聚合是否正确
- 周报分类统计是否正确
- 未完成事项识别是否正确
- interruptibility 判断是否正确
- 冷却规则是否生效

### 集成测试

- 构造 7 天活动记录 -> 周报进入 ready
- 用户触发使用事件 -> ready 周报被投递
- 构造未完成事项 -> 能触发 nudge
- 连续触发 -> 被冷却拦截

### 前端验证

- push 能正常进入 UI
- 卡片与气泡展示正确
- 不重复弹出

---

## 18. 实现结论

第一阶段实现设计建议采用：

- 结构化活动记录
- 统一状态快照
- 用户可打扰状态抽象
- 规则驱动准入
- provider 化事件实现
- 统一投递与冷却控制

优先完成两个主动事件：

- `weekly_report`
- `progress_nudge`

一句话总结：

**第一期实现的重点不是把主动性做花，而是先把“记录、可打扰判断、事件准入、投递”四条链路做稳。**
