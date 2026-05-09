package proactive

import (
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
)

func TestWeeklyReportProviderEvaluate(t *testing.T) {
	workspace := t.TempDir()
	activityStore, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	now := time.Date(2026, 5, 8, 10, 0, 0, 0, time.Local)
	events := []*activity.Event{
		{
			ID:          "1",
			CharacterID: "pet_001",
			SessionID:   "s1",
			Type:        activity.EventUserMessage,
			Category:    activity.CategoryCode,
			Status:      activity.StatusDone,
			Title:       "修复问题",
			CreatedAt:   now.Add(-time.Hour),
		},
		{
			ID:          "2",
			CharacterID: "pet_001",
			SessionID:   "s1",
			Type:        activity.EventTaskResult,
			Category:    activity.CategoryCode,
			Status:      activity.StatusDone,
			Title:       "完成打包修复",
			CreatedAt:   now.Add(-30 * time.Minute),
		},
		{
			ID:          "3",
			CharacterID: "pet_001",
			SessionID:   "s1",
			Type:        activity.EventTaskResult,
			Category:    activity.CategoryDoc,
			Status:      activity.StatusPending,
			Title:       "补充联调文档",
			CreatedAt:   now.Add(-10 * time.Minute),
		},
	}
	for _, ev := range events {
		if err := activityStore.Append(ev); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	var snapshot Snapshot
	snapshot.Now = now
	snapshot.EvaluationReason = "user_message"
	snapshot.Pet.CharacterID = "pet_001"
	snapshot.Pet.PersonaType = "gentle"
	snapshot.Pet.PersonalityTone = "甜心夹子"
	snapshot.Pet.DominantEmotion = "joy"
	snapshot.Preferences.WeeklyReportEnabled = true

	provider := NewWeeklyReportProvider(activityStore, NewWeeklyReportStateStore(workspace))
	intent, ok, err := provider.Evaluate(snapshot)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !ok || intent == nil {
		t.Fatal("expected weekly report intent")
	}
	if intent.Type != "weekly_report" {
		t.Fatalf("Type = %q", intent.Type)
	}
	if _, ok := intent.Payload["report_id"]; !ok {
		t.Fatal("expected report_id in payload")
	}
}

func TestWeeklyReportProviderPrepareThenDeliverOnUserMessage(t *testing.T) {
	workspace := t.TempDir()
	activityStore, err := activity.NewStore(workspace)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	now := time.Date(2026, 5, 8, 10, 0, 0, 0, time.Local)
	events := []*activity.Event{
		{ID: "1", CharacterID: "pet_001", SessionID: "s1", Type: activity.EventUserMessage, Category: activity.CategoryCode, Status: activity.StatusDone, Title: "修复问题", CreatedAt: now.Add(-time.Hour)},
		{ID: "2", CharacterID: "pet_001", SessionID: "s1", Type: activity.EventTaskResult, Category: activity.CategoryCode, Status: activity.StatusDone, Title: "完成打包修复", CreatedAt: now.Add(-30 * time.Minute)},
		{ID: "3", CharacterID: "pet_001", SessionID: "s1", Type: activity.EventTaskResult, Category: activity.CategoryDoc, Status: activity.StatusPending, Title: "补充联调文档", CreatedAt: now.Add(-10 * time.Minute)},
	}
	for _, ev := range events {
		if err := activityStore.Append(ev); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	provider := NewWeeklyReportProvider(activityStore, NewWeeklyReportStateStore(workspace))

	var scheduled Snapshot
	scheduled.Now = now
	scheduled.EvaluationReason = "scheduled_tick"
	scheduled.Pet.CharacterID = "pet_001"
	scheduled.Pet.PersonaType = "gentle"
	scheduled.Preferences.WeeklyReportEnabled = true

	intent, ok, err := provider.Evaluate(scheduled)
	if err != nil {
		t.Fatalf("scheduled Evaluate() error = %v", err)
	}
	if ok || intent != nil {
		t.Fatal("scheduled evaluation should only prepare state, not deliver")
	}

	var interactive Snapshot
	interactive = scheduled
	interactive.EvaluationReason = "user_message"
	intent, ok, err = provider.Evaluate(interactive)
	if err != nil {
		t.Fatalf("interactive Evaluate() error = %v", err)
	}
	if !ok || intent == nil {
		t.Fatal("interactive evaluation should deliver prepared report")
	}
	if intent.Type != "weekly_report" {
		t.Fatalf("Type = %q", intent.Type)
	}
}
