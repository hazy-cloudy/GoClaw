package proactive

import "testing"

func TestManagerResolveDeliveryAllowsImmediateWhenPolicyAllows(t *testing.T) {
	mgr := &Manager{}
	intent := Intent{Type: "progress_nudge", Payload: map[string]any{"deadline_due": true, "delay_count": 0}}
	policy := PolicyDecision{Allowed: true, DeliveryLevel: DeliveryCard}

	level, allowed, retryAfter := mgr.resolveDelivery(intent, policy)
	if !allowed {
		t.Fatal("expected delivery to be allowed")
	}
	if level != DeliveryCard {
		t.Fatalf("level = %q, want %q", level, DeliveryCard)
	}
	if retryAfter != 0 {
		t.Fatalf("retryAfter = %v, want 0", retryAfter)
	}
}

func TestManagerResolveDeliveryRetriesDeadlineReminderTwice(t *testing.T) {
	mgr := &Manager{}
	intent := Intent{Type: "progress_nudge", Payload: map[string]any{"deadline_due": true, "delay_count": 1}}
	policy := PolicyDecision{Allowed: false, DeliveryLevel: DeliveryBlocked}

	level, allowed, retryAfter := mgr.resolveDelivery(intent, policy)
	if allowed {
		t.Fatal("expected delivery to be delayed")
	}
	if level != DeliveryBlocked {
		t.Fatalf("level = %q, want %q", level, DeliveryBlocked)
	}
	if retryAfter <= 0 {
		t.Fatalf("retryAfter = %v, want positive delay", retryAfter)
	}
}

func TestManagerResolveDeliveryForcesAfterTwoRetries(t *testing.T) {
	mgr := &Manager{}
	intent := Intent{Type: "progress_nudge", Payload: map[string]any{"deadline_due": true, "delay_count": 2}}
	policy := PolicyDecision{Allowed: false, DeliveryLevel: DeliveryBlocked}

	level, allowed, retryAfter := mgr.resolveDelivery(intent, policy)
	if !allowed {
		t.Fatal("expected final fallback delivery")
	}
	if level != DeliveryBubble {
		t.Fatalf("level = %q, want %q", level, DeliveryBubble)
	}
	if retryAfter != 0 {
		t.Fatalf("retryAfter = %v, want 0", retryAfter)
	}
}
