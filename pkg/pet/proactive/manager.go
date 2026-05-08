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
	buildSnapshot    func() Snapshot
	providers        []Provider
	deliver          func(Intent, DeliveryLevel) error
	lastPushAt       time.Time
}

// NewManager 创建主动性系统的总协调器。
//
// 这里先只接：
// - buildSnapshot: 每次评估时如何获取当前状态
// - history: 已投递记录
//
// 后续真正接入 weekly_report / progress_nudge provider 时，
// 也是从这里继续往下扩。
func NewManager(history *HistoryStore, buildSnapshot func() Snapshot, providers []Provider, deliver func(Intent, DeliveryLevel) error) *Manager {
	return &Manager{
		interruptibility: NewInterruptibilityEvaluator(),
		policy:           NewPolicyEngine(),
		history:          history,
		buildSnapshot:    buildSnapshot,
		providers:        providers,
		deliver:          deliver,
	}
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

	initialSnapshot := m.buildSnapshot()
	interval := time.Duration(initialSnapshot.Preferences.ProactiveCheckMinutes) * time.Minute
	if interval <= 0 {
		interval = 30 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evaluate("scheduled_tick")
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
	m.evaluate(reason)
}

// evaluate 是主动性底座里的核心评估入口。
// 这里会依次跑：snapshot -> interruptibility -> policy -> providers -> delivery。
func (m *Manager) evaluate(reason string) {
	snapshot := m.buildSnapshot()
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
	if !policy.Allowed || m.deliver == nil {
		return
	}

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
		if err := m.deliver(*intent, policy.DeliveryLevel); err != nil {
			logger.WarnCF("pet", "proactive delivery failed", map[string]any{
				"provider": provider.Name(),
				"error":    err.Error(),
			})
		}
		return
	}
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
