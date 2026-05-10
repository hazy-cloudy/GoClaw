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
	dueAt := now.Add(-5 * time.Minute)
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "补完提醒逻辑",
		CreatedAt:   now.Add(-45 * time.Minute),
		Meta: map[string]any{
			"schedule_kind": "at",
			"due_at":        dueAt.Format(time.RFC3339),
			"job_id":        "job-1",
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
	snapshot.Activity.DeadlineJobID = "job-1"
	snapshot.EvaluationReason = "task_deadline:job-1"

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

func TestProgressNudgeProviderStopsAfterSingleDeadlineReminder(t *testing.T) {
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
			"job_id":        "job-1",
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

	var snapshot Snapshot
	snapshot.Now = now
	snapshot.Preferences.ProgressNudgeEnabled = true
	snapshot.Pet.CharacterID = "pet_001"
	snapshot.Activity.ActiveSessionID = "session-1"
	snapshot.Activity.UnfinishedTaskCount = 1
	snapshot.Activity.DeadlineJobID = "job-1"
	snapshot.EvaluationReason = "task_deadline:job-1"

	intent, ok, err := NewProgressNudgeProvider(store, history).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("expected provider to stop after an existing deadline reminder history")
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
			"job_id":        "job-1",
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
	snapshot.Activity.DeadlineJobID = "job-1"
	snapshot.EvaluationReason = "task_deadline:job-1"

	intent, ok, err := NewProgressNudgeProvider(store, NewHistoryStore(workspace)).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("expected recurring cron task to be ignored by proactive reminder")
	}
}

func TestProgressNudgeProviderDoesNotRemindBeforeDeadline(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	now := time.Now()
	dueAt := now.Add(2 * time.Hour)
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "喝水",
		CreatedAt:   now.Add(-24 * time.Hour),
		Meta: map[string]any{
			"schedule_kind": "at",
			"due_at":        dueAt.Format(time.RFC3339),
			"job_id":        "job-1",
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
	snapshot.Activity.DeadlineJobID = "job-1"
	snapshot.EvaluationReason = "task_deadline:job-1"

	intent, ok, err := NewProgressNudgeProvider(store, NewHistoryStore(workspace)).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("expected no progress nudge before deadline")
	}
}

func TestProgressNudgeProviderMarksDeadlineDueOnFinalReminder(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	now := time.Now()
	dueAt := now.Add(-5 * time.Minute)
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "喝水",
		CreatedAt:   now.Add(-24 * time.Hour),
		Meta: map[string]any{
			"schedule_kind": "at",
			"due_at":        dueAt.Format(time.RFC3339),
			"job_id":        "job-1",
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
	snapshot.Activity.DeadlineJobID = "job-1"
	snapshot.EvaluationReason = "task_deadline:job-1"

	intent, ok, err := NewProgressNudgeProvider(store, NewHistoryStore(workspace)).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok || intent == nil {
		t.Fatal("expected final reminder intent")
	}
	deadlineDue, _ := intent.Payload["deadline_due"].(bool)
	if !deadlineDue {
		t.Fatal("expected deadline_due=true on final reminder")
	}
}

func TestProgressNudgeProviderAllowsThirdAttemptAfterTwoDelays(t *testing.T) {
	workspace := t.TempDir()
	store, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	now := time.Now()
	dueAt := now.Add(-5 * time.Minute)
	event := &activity.Event{
		ID:          "task-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        activity.EventTaskResult,
		Category:    activity.CategoryCode,
		Status:      activity.StatusPending,
		Title:       "第三次兜底",
		CreatedAt:   now.Add(-24 * time.Hour),
		Meta: map[string]any{
			"schedule_kind": "at",
			"due_at":        dueAt.Format(time.RFC3339),
			"job_id":        "job-1",
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
	snapshot.Activity.DeadlineJobID = "job-1"
	snapshot.Activity.DeadlineRetryCount = 2
	snapshot.EvaluationReason = "task_deadline:job-1:2"

	intent, ok, err := NewProgressNudgeProvider(store, NewHistoryStore(workspace)).Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok || intent == nil {
		t.Fatal("expected final fallback reminder after two delays")
	}
}
