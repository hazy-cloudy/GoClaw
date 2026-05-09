package proactive

import (
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
)

func TestProgressNudgeProviderEvaluate(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	now := time.Now()
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "补完提醒逻辑",
		CreatedAt:   now.Add(-45 * time.Minute),
	}
	if err := store.Append(event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	var snapshot Snapshot
	snapshot.Now = now
	snapshot.Preferences.ProgressNudgeEnabled = true
	snapshot.Pet.CharacterID = "pet_001"
	snapshot.Activity.ActiveSessionID = "session-1"
	snapshot.Activity.UnfinishedTaskCount = 1

	intent, ok, err := NewProgressNudgeProvider(store, NewHistoryStore(workspace)).Evaluate(snapshot)
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

func TestProgressNudgeProviderRespectsThirdAndFinalReminderLimit(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	now := time.Now()
	dueAt := now.Add(-10 * time.Minute)
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "补完提醒逻辑",
		CreatedAt:   now.Add(-24 * time.Hour),
		Meta: map[string]any{
			"schedule_kind": "at",
			"due_at":        dueAt.Format(time.RFC3339),
		},
	}
	if err := store.Append(event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	history := NewHistoryStore(workspace)
	if err := history.Append(DeliveryHistoryRecord{
		EventType:   "progress_nudge",
		EventID:     "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		DeliveredAt: now.Add(-6 * time.Hour),
	}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := history.Append(DeliveryHistoryRecord{
		EventType:   "progress_nudge",
		EventID:     "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		DeliveredAt: now.Add(-2 * time.Hour),
	}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	var snapshot Snapshot
	snapshot.Now = now
	snapshot.Preferences.ProgressNudgeEnabled = true
	snapshot.Pet.CharacterID = "pet_001"
	snapshot.Activity.ActiveSessionID = "session-1"
	snapshot.Activity.UnfinishedTaskCount = 1

	intent, ok, err := NewProgressNudgeProvider(store, history).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok || intent == nil {
		t.Fatal("expected provider to allow the third deadline reminder")
	}

	if err := history.Append(DeliveryHistoryRecord{
		EventType:   "progress_nudge",
		EventID:     "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		DeliveredAt: now,
	}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	intent, ok, err = NewProgressNudgeProvider(store, history).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("expected provider to stop after the third reminder")
	}
}

func TestProgressNudgeProviderSkipsRecurringTasks(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	now := time.Now()
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "每周同步",
		CreatedAt:   now.Add(-7 * 24 * time.Hour),
		Meta: map[string]any{
			"schedule_kind": "cron",
			"cron_expr":     "0 9 * * 1",
		},
	}
	if err := store.Append(event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	var snapshot Snapshot
	snapshot.Now = now
	snapshot.Preferences.ProgressNudgeEnabled = true
	snapshot.Pet.CharacterID = "pet_001"
	snapshot.Activity.ActiveSessionID = "session-1"
	snapshot.Activity.UnfinishedTaskCount = 1

	intent, ok, err := NewProgressNudgeProvider(store, NewHistoryStore(workspace)).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("expected recurring cron task to be ignored by proactive reminder")
	}
}

func TestProgressNudgeProviderSkipsSecondReminderAfterUserReply(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	now := time.Now()
	dueAt := now.Add(24 * time.Hour)
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "补完提醒逻辑",
		CreatedAt:   now.Add(-24 * time.Hour),
		Meta: map[string]any{
			"schedule_kind": "at",
			"due_at":        dueAt.Format(time.RFC3339),
		},
	}
	if err := store.Append(event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	history := NewHistoryStore(workspace)
	firstDeliveredAt := now.Add(-8 * time.Hour)
	if err := history.Append(DeliveryHistoryRecord{
		EventType:   "progress_nudge",
		EventID:     "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		DeliveredAt: firstDeliveredAt,
	}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	var snapshot Snapshot
	snapshot.Now = now
	snapshot.Preferences.ProgressNudgeEnabled = true
	snapshot.Pet.CharacterID = "pet_001"
	snapshot.Activity.ActiveSessionID = "session-1"
	snapshot.Activity.UnfinishedTaskCount = 1
	snapshot.Activity.LastUserMessageAt = firstDeliveredAt.Add(30 * time.Minute)

	intent, ok, err := NewProgressNudgeProvider(store, history).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("expected second reminder to be skipped after user replied")
	}
}
