package proactive

import (
	"testing"
	"time"
)

func TestPolicyDisabled(t *testing.T) {
	var snapshot Snapshot
	snapshot.Now = time.Now()
	snapshot.Preferences.ProactiveCare = false

	decision := NewPolicyEngine().Evaluate(snapshot, InterruptibilityDecision{
		Interruptible: true,
		Level:         InterruptibilityYes,
	})
	if decision.Allowed {
		t.Fatal("expected policy to block when proactive care is disabled")
	}
}

func TestPolicyAllowedBubble(t *testing.T) {
	var snapshot Snapshot
	snapshot.Now = time.Now()
	snapshot.Preferences.ProactiveCare = true
	snapshot.Preferences.GlobalCooldownMins = 90
	snapshot.Activity.UnfinishedTaskCount = 1

	decision := NewPolicyEngine().Evaluate(snapshot, InterruptibilityDecision{
		Interruptible: true,
		Level:         InterruptibilitySoft,
	})
	if !decision.Allowed {
		t.Fatal("expected policy to allow event")
	}
	if decision.DeliveryLevel != DeliveryBubble {
		t.Fatalf("DeliveryLevel = %q, want %q", decision.DeliveryLevel, DeliveryBubble)
	}
}

func TestPolicyAllowsFinalDeadlineFallbackDespiteCooldown(t *testing.T) {
	now := time.Now()
	var snapshot Snapshot
	snapshot.Now = now
	snapshot.Preferences.ProactiveCare = true
	snapshot.Preferences.GlobalCooldownMins = 10
	snapshot.Activity.UnfinishedTaskCount = 1
	snapshot.Activity.LastPushAt = now.Add(-1 * time.Minute)
	snapshot.Activity.DeadlineRetryCount = 2
	snapshot.Pet.DominantEmotion = "joy"

	decision := NewPolicyEngine().Evaluate(snapshot, InterruptibilityDecision{
		Interruptible: true,
		Level:         InterruptibilitySoft,
	})
	if !decision.Allowed {
		t.Fatal("expected final deadline fallback to bypass cooldown")
	}
}
