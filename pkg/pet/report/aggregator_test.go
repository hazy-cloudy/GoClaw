package report

import (
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
)

func TestAggregateWeeklyReport(t *testing.T) {
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

	report := AggregateWeeklyReport("pet_001", events, now)
	if report.MessageCount != 1 {
		t.Fatalf("MessageCount = %d, want 1", report.MessageCount)
	}
	if report.TaskCount != 2 {
		t.Fatalf("TaskCount = %d, want 2", report.TaskCount)
	}
	if len(report.Unfinished) != 1 || report.Unfinished[0] != "补充联调文档" {
		t.Fatalf("Unfinished = %#v", report.Unfinished)
	}
	if len(report.Outputs) != 1 || report.Outputs[0] != "完成打包修复" {
		t.Fatalf("Outputs = %#v", report.Outputs)
	}
}
