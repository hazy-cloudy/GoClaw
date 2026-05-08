package proactive

import (
	"strings"
	"time"
)

type PolicyEngine struct{}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{}
}

// Evaluate 在 interruptibility 之后做第二层规则判断。
//
// 这层更偏“系统规则”：
// - 是否总开关关闭
// - 是否命中全局冷却
// - 当前更适合气泡还是卡片
func (e *PolicyEngine) Evaluate(snapshot Snapshot, interruptibility InterruptibilityDecision) PolicyDecision {
	if !snapshot.Preferences.ProactiveCare {
		return PolicyDecision{
			Allowed:       false,
			DeliveryLevel: DeliveryBlocked,
			Score:         0,
			ReasonCodes:   []string{"proactive_disabled"},
		}
	}

	if !interruptibility.Interruptible || interruptibility.Level == InterruptibilityNo {
		return PolicyDecision{
			Allowed:       false,
			DeliveryLevel: DeliveryBlocked,
			Score:         0,
			ReasonCodes:   append([]string{"not_interruptible"}, interruptibility.ReasonCodes...),
		}
	}

	now := snapshot.Now
	if now.IsZero() {
		now = time.Now()
	}

	cooldown := time.Duration(snapshot.Preferences.GlobalCooldownMins) * time.Minute
	if cooldown <= 0 {
		cooldown = 90 * time.Minute
	}
	if !snapshot.Activity.LastPushAt.IsZero() && now.Sub(snapshot.Activity.LastPushAt) < cooldown {
		return PolicyDecision{
			Allowed:       false,
			DeliveryLevel: DeliveryBlocked,
			Score:         0,
			ReasonCodes:   []string{"global_cooldown"},
		}
	}

	// score 先用简单规则打分，后面接 event provider 时再继续细化。
	score := 0
	reasons := append([]string{}, interruptibility.ReasonCodes...)
	if snapshot.Activity.UnfinishedTaskCount > 0 {
		score += 20
		reasons = append(reasons, "unfinished_tasks")
	}
	if snapshot.User.PressureLevel == "high" || snapshot.User.PressureLevel == "critical" {
		score += 20
		reasons = append(reasons, "high_pressure")
	}
	if snapshot.User.EnergyLevel > 0 && snapshot.User.EnergyLevel <= 35 {
		score += 10
		reasons = append(reasons, "low_energy")
	}
	if snapshot.Activity.RecentMessageCount == 0 {
		score += 10
		reasons = append(reasons, "low_recent_activity")
	}
	dominantEmotion := strings.ToLower(snapshot.Pet.DominantEmotion)
	if dominantEmotion == "fear" || dominantEmotion == "joy" {
		score += 5
		reasons = append(reasons, "pet_emotion_bias")
	}

	level := DeliveryBlocked
	allowed := false
	switch {
	case score >= 40:
		level = DeliveryCard
		allowed = true
	case score >= 20:
		level = DeliveryBubble
		allowed = true
	}

	return PolicyDecision{
		Allowed:       allowed,
		DeliveryLevel: level,
		Score:         score,
		ReasonCodes:   reasons,
	}
}
