package proactive

import (
	"testing"
	"time"
)

func TestProgressNudgeProviderEvaluate(t *testing.T) {
	var snapshot Snapshot
	snapshot.Now = time.Now()
	snapshot.Preferences.ProgressNudgeEnabled = true
	snapshot.Activity.UnfinishedTaskCount = 2

	intent, ok, err := NewProgressNudgeProvider().Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok || intent == nil {
		t.Fatal("expected progress nudge intent")
	}
	if intent.Type != "progress_nudge" {
		t.Fatalf("Type = %q", intent.Type)
	}
}
