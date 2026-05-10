package proactive

import (
	"testing"
	"time"
)

func TestInterruptibilityBusy(t *testing.T) {
	var snapshot Snapshot
	snapshot.Now = time.Now()
	snapshot.Activity.CurrentSessionBusy = true

	decision := NewInterruptibilityEvaluator().Evaluate(snapshot)
	if decision.Interruptible {
		t.Fatal("expected not interruptible when session is busy")
	}
	if decision.Level != InterruptibilityNo {
		t.Fatalf("Level = %q, want %q", decision.Level, InterruptibilityNo)
	}
}

func TestInterruptibilityVisible(t *testing.T) {
	var snapshot Snapshot
	snapshot.Now = time.Now()
	snapshot.Activity.ConsoleVisible = true

	decision := NewInterruptibilityEvaluator().Evaluate(snapshot)
	if !decision.Interruptible {
		t.Fatal("expected interruptible when console is visible")
	}
	if decision.Level != InterruptibilityYes {
		t.Fatalf("Level = %q, want %q", decision.Level, InterruptibilityYes)
	}
}

func TestInterruptibilityRecentSingleMessageNotBusyForever(t *testing.T) {
	var snapshot Snapshot
	snapshot.Now = time.Now()
	snapshot.Activity.CurrentSessionBusy = true
	snapshot.Activity.LastUserMessageAt = snapshot.Now.Add(-50 * time.Second)

	decision := NewInterruptibilityEvaluator().Evaluate(snapshot)
	if !decision.Interruptible {
		t.Fatal("expected interruptible once last user message is not immediate")
	}
	if decision.Level != InterruptibilitySoft {
		t.Fatalf("Level = %q, want %q", decision.Level, InterruptibilitySoft)
	}
}
