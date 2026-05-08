# 桌宠主动性系统架构设计（第一期）

## 1. 文档定位

本文档属于**架构设计**，重点描述：

- 桌宠主动性系统在当前项目中的组织方式
- 核心组件有哪些
- 组件之间如何协作
- `weekly_report` 和 `progress_nudge` 在整体系统中的位置
- “用户是否可打扰”这一能力在系统中处于什么层级

本文档不进入字段级、协议级、状态机级细节，这些内容属于实现设计。

---

## 2. 架构目标

第一期主动性系统的架构目标有五个：

1. **统一调度**
   - 所有主动事件都通过同一入口判断、调度与投递

2. **状态驱动**
   - 主动行为基于用户状态、桌宠状态和近期活动，而不是死板定时提醒

3. **可打扰判断独立化**
   - “用户现在是否适合被打扰”不能散落在各个事件内部，而必须成为公共能力

4. **可扩展**
   - 第一阶段只做两个事件，但后续事件类型可以平滑加入

5. **决策与渲染分离**
   - “是否打扰”“打扰什么”“怎么说”在架构层面必须解耦

---

## 3. 在当前项目中的归属

主动性系统建议作为 `pet` 模块的子系统引入，而不是散落在：

- Hook 链路中
- 前端页面逻辑中
- 各种零散定时器中

原因：

- 桌宠人格、情绪、用户画像、主动配置都属于 `pet` 领域
- `PetService` 已经承担了桌宠长生命周期能力的组织职责
- 主动事件最终也会通过 `pet` push 机制投递到前端

因此主动性系统的宿主建议是：

- [pkg/pet/service.go](/D:/study%20part/GoClawPet/pkg/pet/service.go)

也就是说：

- `PetService.Start()` 负责启动 proactive loop
- proactive loop 在后台周期性评估是否有事件进入 `ready`
- 用户使用事件负责触发“是否现在投递”的即时判定

---

## 4. 总体架构分层

建议整体拆成 6 层：

1. Activity Layer
2. Snapshot Layer
3. Interruptibility Layer
4. Policy Layer
5. Event Provider Layer
6. Delivery Layer

### 4.1 Activity Layer

负责记录“用户让桌宠做过什么”的结构化活动。

作用：

- 为 `weekly_report` 提供可聚合数据
- 为 `progress_nudge` 提供未完成事项线索

它不是给用户看的聊天记录，而是主动决策的基础数据层。

### 4.2 Snapshot Layer

负责把主动性系统评估所需的状态统一收敛成快照。

快照内容应包括：

- 桌宠人格与情绪
- 用户画像与实时状态
- 最近活跃程度
- 未完成事项数量
- 主动性配置
- 最近一次主动触达历史

Snapshot 是“当前时刻系统视角”的标准输入。

### 4.3 Interruptibility Layer

负责判断：

- 当前用户是否可打扰
- 当前时机适不适合真正把主动事件发出去

这一层不关心：

- 是周报还是催办
- 文案怎么写
- 事件是否有意义

它只回答：

- 这个用户现在适合被打扰吗？

### 4.4 Policy Layer

负责决定：

- 当前时刻某个事件是否允许进入触达候选
- 允许什么级别的触达
- 哪些事件更值得进入下一轮评估

Policy 是系统的“门禁层”，但它建立在 Interruptibility 之后。

### 4.5 Event Provider Layer

每个主动事件作为独立 provider 存在。

第一期 provider：

- `weekly_report`
- `progress_nudge`

以后新增事件也沿着同样模式扩展，例如：

- `care_checkin`
- `celebration_ping`
- `risk_alert`

### 4.6 Delivery Layer

负责把主动事件变成前端可展示的触达结果。

输出形式包括：

- 桌宠气泡
- 聊天卡片
- 会话消息

Delivery 同时负责：

- 投递历史
- 防重
- 降级
- 全局冷却

---

## 5. 核心架构原则

### 5.1 决策与表达必须分离

主动性系统需要同时处理三个问题：

- 要不要打扰
- 打扰什么
- 怎么说

这三件事在架构上应映射为：

- Interruptibility + Policy
- Event Provider
- Renderer / Delivery

这样做的好处是：

- 易调试
- 易测试
- 易替换
- 易复用

### 5.2 “可打扰判断”必须先于事件投递

无论事件本身多合理，如果用户当前不适合被打扰，这次主动触达都应该被推迟。

所以系统的顺序应该是：

1. 事件是否 ready
2. 用户是否可打扰
3. 当前是否允许这个事件投递
4. 再进行渲染和投递

这样可以避免：

- 周报在用户正忙时硬插入
- 催办在用户连续发消息时突然冒出来

### 5.3 周报采用“两阶段触发模型”

`weekly_report` 不应该等价于一个简单 cron。

它应该拆成两个阶段：

#### 阶段 A：Ready

时间窗满足后，后台判断是否有资格生成本周周报，并把状态置为 ready。

#### 阶段 B：Delivered

在用户出现使用行为时，再结合 interruptibility 决定是否真正发出去。

也就是说：

- 时间决定“周报是否准备好”
- 使用行为决定“周报什么时候发”

### 5.4 事件与调度分离

每个事件不应该自带自己的定时器。

统一由 `Proactive Manager` 周期性：

- 拉快照
- 跑 interruptibility
- 跑 policy
- 评估 provider
- 选择一个或多个 intent

这能保证：

- 冷却规则统一
- 静默规则统一
- ready 状态统一
- 事件之间不互相乱抢

### 5.5 第一阶段优先规则化，而不是智能化

架构应该允许未来引入更多智能策略，但第一阶段的主路径必须是规则驱动。

这是为了优先保障：

- 可解释
- 可验证
- 可控制

---

## 6. 核心组件划分

### 6.1 Proactive Manager

建议位置：

- `pkg/pet/proactive/manager.go`

职责：

- 周期性评估
- 组织 snapshot / interruptibility / policy / provider / delivery
- 协调整条主动事件链路

它是整个主动性系统的 orchestrator。

### 6.2 Activity Store

建议位置：

- `pkg/pet/activity/store.go`

职责：

- 持久化活动事件
- 按时间窗口查询
- 按角色、会话、类别筛选

### 6.3 Snapshot Builder

建议位置：

- `pkg/pet/proactive/snapshot.go`

职责：

- 从分散的状态源中构建统一快照

数据来源包括：

- `characters.Manager`
- `emotion.Engine`
- `userprofile.Manager`
- `config.Manager`
- `activity.Store`

### 6.4 Interruptibility Evaluator

建议位置：

- `pkg/pet/proactive/interruptibility.go`

职责：

- 判断用户当前是否可打扰
- 输出 `interruptible / soft / blocked`
- 为所有主动事件提供统一时机门控

建议它考虑：

- 当前会话是否 busy
- 最近消息活跃度
- 最近是否刚收到主动事件
- 当前是否命中静默时段
- 当前面板是否处于活跃使用态

### 6.5 Policy Engine

建议位置：

- `pkg/pet/proactive/policy.go`

职责：

- 对 event provider 输出结果做最终准入判断
- 控制触达等级
- 结合事件冷却做二次拦截

### 6.6 Event Providers

建议统一接口后，第一期至少实现：

- `WeeklyReportProvider`
- `ProgressNudgeProvider`

建议位置：

- `pkg/pet/report/weekly.go`
- `pkg/pet/proactive/provider_progress_nudge.go`

### 6.7 Delivery Service

建议位置：

- `pkg/pet/proactive/delivery.go`

职责：

- 将 intent 转为 push payload
- 写入投递历史
- 控制 UI 投递形式

---

## 7. 第一阶段事件在架构中的位置

### 7.1 weekly_report

在主动性系统中的位置：

- 输入：一周活动记录
- 决策：低频回顾型 provider
- 输出：结构化周报 + 人格化摘要
- 投递：优先卡片，其次气泡摘要

它更像一个“复盘类主动事件”。

而且它采用两阶段模型：

1. 后台进入 `ready`
2. 用户可打扰时再投递

### 7.2 progress_nudge

在主动性系统中的位置：

- 输入：未完成事项线索 + 最近活跃状态
- 决策：轻催办型 provider
- 输出：简短提醒与下一步建议
- 投递：优先气泡，必要时升级卡片

它更像一个“陪跑类主动事件”。

它也不应纯定时直接发，而应：

- 周期性评估是否有提醒资格
- 在用户可打扰时再轻量触发

---

## 8. 系统交互关系

整体关系建议如下：

```text
User / Session / Tool Activity
        ↓
Activity Collector
        ↓
Activity Store
        ↓
Snapshot Builder ← Character / Emotion / UserProfile / Config
        ↓
Interruptibility Evaluator
        ↓
Policy Engine
        ↓
Event Providers
   ├─ WeeklyReportProvider
   └─ ProgressNudgeProvider
        ↓
Renderer
        ↓
Delivery Layer
        ↓
Pet Push / Bubble / Chat Card
```

这条链的核心思想是：

- 先记录行为
- 再收敛状态
- 再判断当前是否可打扰
- 再判断事件是否允许进入投递
- 再挑选事件
- 最后才表达和投递

---

## 9. 当前项目接入点

### 服务端

主要接入点：

- [pkg/pet/service.go](/D:/study%20part/GoClawPet/pkg/pet/service.go)
- [pkg/pet/hooks.go](/D:/study%20part/GoClawPet/pkg/pet/hooks.go)
- [pkg/pet/config/types.go](/D:/study%20part/GoClawPet/pkg/pet/config/types.go)
- [pkg/pet/types.go](/D:/study%20part/GoClawPet/pkg/pet/types.go)
- [pkg/pet/userprofile/types.go](/D:/study%20part/GoClawPet/pkg/pet/userprofile/types.go)

### 前端

建议接入点：

- `clawpet-frontend/clawpet/lib/api/websocket.ts`
- `clawpet-frontend/clawpet/hooks/use-chat.ts`
- `clawpet-frontend/clawpet/app/desktop-pet/page.tsx`
- `clawpet-frontend/clawpet/components/chat-area.tsx`

前端只负责：

- 接收主动事件 push
- 展示气泡与卡片
- 提供点击/忽略交互

前端不负责主动决策。

---

## 10. 包结构建议

建议新增：

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

pkg/pet/report/
  weekly.go
  aggregator.go
  renderer.go
```

这样划分的原因：

- `activity` 负责记账
- `proactive` 负责调度、可打扰判断、准入和投递
- `report` 负责周报特有逻辑

边界清晰，便于后续继续扩展。

---

## 11. 演进路径

### 第一阶段

- 建立主动性基础框架
- 固化“可打扰判断”层
- 只接两个 provider
- 只做基础投递形态

### 第二阶段

- 增加更多主动事件
- 增强未完成事项识别
- 丰富前端承载形式

### 第三阶段

- 引入更复杂的时机学习
- 支持多事件排序与竞争
- 支持更强的跨周期分析

---

## 12. 架构结论

第一阶段的桌宠主动性不应实现为两个零散功能，而应作为 `pet` 模块下的统一主动事件子系统引入。

在这个架构下：

- `weekly_report` 负责低频高价值的回顾体验
- `progress_nudge` 负责中低频陪跑体验

两者共享：

- 行为记录
- 状态快照
- 可打扰判断
- 准入策略
- 投递与冷却机制

一句话总结：

**主动性第一期的重点不是做多少事件，而是先把“记录、时机、准入、投递”这套骨架搭起来。**
