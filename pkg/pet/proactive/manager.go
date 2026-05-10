package proactive

import (
	"context"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

type Manager struct {
	mu               sync.RWMutex
	interruptibility *InterruptibilityEvaluator
	policy           *PolicyEngine
	history          *HistoryStore
	buildSnapshot    func(string) Snapshot
	providers        []Provider
	deliver          func(Intent, DeliveryLevel) error
	lastPushAt       time.Time
	retry            func(string, time.Duration)
}

type EvaluationResult struct {
	IntentType     string
	Delivered      bool
	DeliveryLevel  DeliveryLevel
	RetryAfter     time.Duration
	BlockedByPolicy bool
	Err            error
}

// NewManager 创建主动性系统的总协调器。
//
// 这里先只接：
// - buildSnapshot: 每次评估时如何获取当前状态
// - history: 已投递记录
//
// 后续真正接入 weekly_report / progress_nudge provider 时，
// 也是从这里继续往下扩。
func NewManager(history *HistoryStore, buildSnapshot func(string) Snapshot, providers []Provider, deliver func(Intent, DeliveryLevel) error) *Manager {
	mgr := &Manager{
		interruptibility: NewInterruptibilityEvaluator(),
		policy:           NewPolicyEngine(),
		history:          history,
		buildSnapshot:    buildSnapshot,
		providers:        providers,
		deliver:          deliver,
	}
	if history != nil {
		if records, err := history.List(); err == nil && len(records) > 0 {
			last := records[len(records)-1].DeliveredAt
			mgr.lastPushAt = last
		}
	}
	return mgr
}

func (m *Manager) SetRetryScheduler(fn func(string, time.Duration)) {
	if m == nil {
		return
	}
	m.retry = fn
}

// Start 启动后台周期性检查。
//
// 第一阶段它做的事很克制：
// - 定时拉一次 snapshot
// - 跑 interruptibility
// - 跑 policy
// - 打日志验证链路闭环
//
// 还不会真正创建周报或催办事件。
func (m *Manager) Start(ctx context.Context) {
	if m == nil || m.buildSnapshot == nil {
		return
	}

	initialSnapshot := m.buildSnapshot("scheduled_tick")
	interval := time.Duration(initialSnapshot.Preferences.ProactiveCheckMinutes) * time.Minute
	if interval <= 0 {
		interval = 30 * time.Minute
	}

	m.evaluate("scheduled_tick", true)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evaluate("scheduled_tick", true)
		}
	}
}

// Trigger 用于“用户使用时”的即时触发。
//
// 这是后面实现“周报 ready 了，但要等用户回来时再发”的关键入口。
func (m *Manager) Trigger(reason string) {
	if m == nil {
		return
	}
	m.evaluate(reason, true)
}

func (m *Manager) EvaluateNow(reason string) EvaluationResult {
	if m == nil {
		return EvaluationResult{}
	}
	return m.evaluate(reason, false)
}

// evaluate 是主动性底座里的核心评估入口。
// 这里会依次跑：snapshot -> interruptibility -> policy -> providers -> delivery。
func (m *Manager) evaluate(reason string, allowRetrySchedule bool) EvaluationResult {
	result := EvaluationResult{}
	snapshot := m.buildSnapshot(reason)
	interruptibility := m.interruptibility.Evaluate(snapshot)
	policy := m.policy.Evaluate(snapshot, interruptibility)
	logger.DebugCF("pet", "proactive evaluated", map[string]any{
		"reason":                 reason,
		"interruptible":          interruptibility.Interruptible,
		"interruptibility_level": interruptibility.Level,
		"policy_allowed":         policy.Allowed,
		"policy_score":           policy.Score,
		"policy_delivery":        policy.DeliveryLevel,
	})
	for _, provider := range m.providers {
		intent, ok, err := provider.Evaluate(snapshot)
		if err != nil {
			logger.WarnCF("pet", "proactive provider evaluate failed", map[string]any{
				"provider": provider.Name(),
				"error":    err.Error(),
			})
			continue
		}
		if !ok || intent == nil {
			continue
		}

		result.IntentType = intent.Type
		deliveryLevel, allowed, retryAfter := m.resolveDelivery(*intent, policy)
		result.DeliveryLevel = deliveryLevel
		result.RetryAfter = retryAfter
		result.BlockedByPolicy = !allowed
		if !allowed || m.deliver == nil {
			if retryAfter > 0 && allowRetrySchedule && m.retry != nil {
				m.retry(snapshot.EvaluationReason, retryAfter)
			}
			return result
		}

		if err := m.deliver(*intent, deliveryLevel); err != nil {
			logger.WarnCF("pet", "proactive delivery failed", map[string]any{
				"provider": provider.Name(),
				"error":    err.Error(),
			})
			result.Err = err
		}
		result.Delivered = result.Err == nil
		return result
	}
	return result
}

func (m *Manager) resolveDelivery(intent Intent, policy PolicyDecision) (DeliveryLevel, bool, time.Duration) {
	if policy.Allowed {
		return policy.DeliveryLevel, true, 0
	}
	if intent.Type != "progress_nudge" || intent.Payload == nil {
		return DeliveryBlocked, false, 0
	}
	deadlineDue, _ := intent.Payload["deadline_due"].(bool)
	if !deadlineDue {
		return DeliveryBlocked, false, 0
	}
	delayCount := 0
	switch v := intent.Payload["delay_count"].(type) {
	case int:
		delayCount = v
	case int64:
		delayCount = int(v)
	case float64:
		delayCount = int(v)
	}
	if delayCount >= 2 {
		return DeliveryBubble, true, 0
	}
	return DeliveryBlocked, false, 2 * time.Minute
}

// RecordDelivery 目前还没被具体事件使用，但这一步先放进来，
// 是为了让全局冷却和防重有统一落点。
func (m *Manager) RecordDelivery(eventType, eventID, characterID string) error {
	if m == nil || m.history == nil {
		return nil
	}
	now := time.Now()
	m.mu.Lock()
	m.lastPushAt = now
	m.mu.Unlock()
	return m.history.Append(DeliveryHistoryRecord{
		EventType:   eventType,
		EventID:     eventID,
		CharacterID: characterID,
		DeliveredAt: now,
	})
}

func (m *Manager) LastPushAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastPushAt
}
