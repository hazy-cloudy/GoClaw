package proactive

import (
	"fmt"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
)

type ProgressNudgeProvider struct {
	activityStore *activity.Store
	history       *HistoryStore
}

func NewProgressNudgeProvider(activityStore *activity.Store, history *HistoryStore) *ProgressNudgeProvider {
	return &ProgressNudgeProvider{
		activityStore: activityStore,
		history:       history,
	}
}

func (p *ProgressNudgeProvider) Name() string {
	return "progress_nudge"
}

func (p *ProgressNudgeProvider) Evaluate(snapshot Snapshot) (*Intent, bool, error) {
	if !snapshot.Preferences.ProgressNudgeEnabled {
		return nil, false, nil
	}
	if snapshot.Activity.UnfinishedTaskCount <= 0 {
		return nil, false, nil
	}
	if snapshot.EvaluationReason != "task_deadline" && !strings.HasPrefix(snapshot.EvaluationReason, "task_deadline:") {
		return nil, false, nil
	}

	task, reminderCount, err := p.pickReminderCandidate(snapshot)
	if err != nil || task == nil {
		return nil, false, err
	}
	dueAt := resolveTaskDueAt(task)
	deadlineDue := !dueAt.IsZero() && !dueAt.After(snapshot.Now)

	nudgeID := fmt.Sprintf("%s-%d", task.ID, reminderCount+1)
	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = "未完成事项"
	}

	return &Intent{
		Type:        "progress_nudge",
		Priority:    "low",
		ReasonCodes: []string{"unfinished_tasks", "deadline_due"},
		Payload: map[string]any{
			"nudge_id":       nudgeID,
			"event_id":       task.ID,
			"topic":          title,
			"summary":        buildProgressSummary(task, snapshot.Now, dueAt),
			"suggestion":     buildProgressSuggestion(task, snapshot.Now, dueAt),
			"pending_cnt":    snapshot.Activity.UnfinishedTaskCount,
			"reminder_count": reminderCount + 1,
			"deadline_due":   deadlineDue,
			"delay_count":    snapshot.Activity.DeadlineRetryCount,
		},
	}, true, nil
}

func (p *ProgressNudgeProvider) pickReminderCandidate(snapshot Snapshot) (*activity.Event, int, error) {
	if snapshot.Now.IsZero() || snapshot.Pet.CharacterID == "" {
		return nil, 0, nil
	}

	start := snapshot.Now.AddDate(0, 0, -14)
	if p == nil {
		return nil, 0, nil
	}
	return p.pickReminderCandidateFromStore(snapshot, p.activityStore, start)
}

func (p *ProgressNudgeProvider) pickReminderCandidateFromStore(snapshot Snapshot, activityStore *activity.Store, start time.Time) (*activity.Event, int, error) {
	if activityStore == nil {
		return nil, 0, nil
	}
	events, err := activityStore.ListRange(start, snapshot.Now)
	if err != nil {
		return nil, 0, err
	}

	var selected *activity.Event
	selectedReminderCount := 0
	selectedDueAt := time.Time{}
	selectedSameSession := false

	for _, ev := range events {
		if ev == nil || ev.Type != activity.EventTaskResult {
			continue
		}
		if ev.CharacterID != "" && ev.CharacterID != snapshot.Pet.CharacterID {
			continue
		}
		if ev.Status != activity.StatusPending && ev.Status != activity.StatusFailed {
			continue
		}
		if isRecurringTask(ev) {
			continue
		}
		if snapshot.Activity.DeadlineJobID != "" {
			jobID, _ := ev.Meta["job_id"].(string)
			if jobID != snapshot.Activity.DeadlineJobID {
				continue
			}
		}

		dueAt := resolveTaskDueAt(ev)
		if dueAt.IsZero() || dueAt.After(snapshot.Now) {
			continue
		}

		reminderCount := snapshot.Activity.DeadlineRetryCount
		if p.history != nil {
			records, err := p.history.FindByEvent("progress_nudge", ev.ID, snapshot.Pet.CharacterID, ev.SessionID)
			if err != nil {
				return nil, 0, err
			}
			if len(records) > 0 {
				continue
			}
		}
		if reminderCount > 2 {
			continue
		}

		sameSession := snapshot.Activity.ActiveSessionID != "" && ev.SessionID == snapshot.Activity.ActiveSessionID
		if selected == nil ||
			(sameSession && !selectedSameSession) ||
			(sameSession == selectedSameSession && (selectedDueAt.IsZero() || dueAt.Before(selectedDueAt))) {
			selected = ev
			selectedReminderCount = reminderCount
			selectedDueAt = dueAt
			selectedSameSession = sameSession
		}
	}
	return selected, selectedReminderCount, nil
}

func resolveTaskDueAt(ev *activity.Event) time.Time {
	if ev == nil || ev.Meta == nil {
		return ev.CreatedAt
	}
	if raw, ok := ev.Meta["due_at"].(string); ok && strings.TrimSpace(raw) != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return t
		}
	}
	if raw, ok := ev.Meta["due_at_ms"].(float64); ok && raw > 0 {
		return time.UnixMilli(int64(raw))
	}
	if raw, ok := ev.Meta["at_ms"].(float64); ok && raw > 0 {
		return time.UnixMilli(int64(raw))
	}
	return ev.CreatedAt
}

func isRecurringTask(ev *activity.Event) bool {
	if ev == nil || ev.Meta == nil {
		return false
	}
	if kind, ok := ev.Meta["schedule_kind"].(string); ok {
		kind = strings.TrimSpace(strings.ToLower(kind))
		return kind == "every" || kind == "cron"
	}
	return false
}

func buildProgressSummary(ev *activity.Event, now, dueAt time.Time) string {
	title := progressTopic(ev)
	if !dueAt.IsZero() && !dueAt.After(now) {
		return fmt.Sprintf("%s 已经到时间点了。", title)
	}
	return fmt.Sprintf("%s 该处理了。", title)
}

func buildProgressSuggestion(ev *activity.Event, now, dueAt time.Time) string {
	title := progressTopic(ev)
	return fmt.Sprintf("记得把 %s 收一下，别错过最后时间。", title)
}

func progressTopic(ev *activity.Event) string {
	if ev == nil {
		return "这件事"
	}
	title := strings.TrimSpace(ev.Title)
	if title != "" {
		return title
	}
	return "这件事"
}
