package proactive

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
)

type ProgressNudgeProvider struct {
	activityStore *activity.Store
	history *HistoryStore
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

	task, reminderCount, err := p.pickReminderCandidate(snapshot)
	if err != nil || task == nil {
		return nil, false, err
	}
	dueAt := resolveTaskDueAt(task)

	nudgeID := fmt.Sprintf("%s-%d", task.ID, reminderCount+1)
	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = "未完成事项"
	}

	return &Intent{
		Type:        "progress_nudge",
		Priority:    "low",
		ReasonCodes: []string{"unfinished_tasks", "time_distance_ready"},
		Payload: map[string]any{
			"nudge_id":       nudgeID,
			"event_id":       task.ID,
			"topic":          title,
			"summary":        buildProgressSummary(task, snapshot.Now, dueAt, reminderCount),
			"suggestion":     buildProgressSuggestion(task, snapshot.Now, dueAt, reminderCount),
			"pending_cnt":    snapshot.Activity.UnfinishedTaskCount,
			"reminder_count": reminderCount + 1,
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
	selectedWait := time.Duration(math.MaxInt64)
	selectedEligibleAt := time.Time{}
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

		dueAt := resolveTaskDueAt(ev)
		firstDelay := firstReminderDelay(snapshot.Now.Sub(ev.CreatedAt), dueAt.Sub(ev.CreatedAt))
		if firstDelay <= 0 {
			firstDelay = 20 * time.Minute
		}
		reminderCount := 0
		var deliveries []DeliveryHistoryRecord
		if p.history != nil {
			records, err := p.history.FindByEvent("progress_nudge", ev.ID, snapshot.Pet.CharacterID, ev.SessionID)
			if err != nil {
				return nil, 0, err
			}
			deliveries = records
			sort.Slice(deliveries, func(i, j int) bool {
				return deliveries[i].DeliveredAt.Before(deliveries[j].DeliveredAt)
			})
			reminderCount = len(deliveries)
		}
		if reminderCount >= 3 {
			continue
		}
		if shouldSkipMiddleReminder(ev, dueAt, deliveries, snapshot.Activity.LastUserMessageAt) {
			continue
		}

		var eligibleAt time.Time
		if reminderCount == 0 {
			eligibleAt = ev.CreatedAt.Add(firstDelay)
		} else if reminderCount == 2 {
			eligibleAt = dueAt
		} else {
			lastDeliveredAt := deliveries[reminderCount-1].DeliveredAt
			backoff := firstDelay * time.Duration(1<<reminderCount)
			eligibleAt = lastDeliveredAt.Add(backoff)
		}
		if !dueAt.IsZero() && eligibleAt.After(dueAt) && reminderCount < 2 {
			eligibleAt = dueAt
		}
		wait := eligibleAt.Sub(snapshot.Now)
		if wait > 0 {
			continue
		}

		sameSession := snapshot.Activity.ActiveSessionID != "" && ev.SessionID == snapshot.Activity.ActiveSessionID
		if selected == nil ||
			(sameSession && !selectedSameSession) ||
			(sameSession == selectedSameSession && (eligibleAt.Before(selectedEligibleAt) || (eligibleAt.Equal(selectedEligibleAt) && wait > selectedWait))) {
			selected = ev
			selectedReminderCount = reminderCount
			selectedWait = wait
			selectedEligibleAt = eligibleAt
			selectedSameSession = sameSession
		}
	}
	return selected, selectedReminderCount, nil
}

func firstReminderDelay(age, leadTime time.Duration) time.Duration {
	if leadTime > 0 {
		switch {
		case leadTime >= 30*24*time.Hour:
			return 7 * 24 * time.Hour
		case leadTime >= 7*24*time.Hour:
			return 24 * time.Hour
		case leadTime >= 24*time.Hour:
			return 6 * time.Hour
		case leadTime >= 6*time.Hour:
			return 90 * time.Minute
		default:
			return 30 * time.Minute
		}
	}
	switch {
	case age >= 14*24*time.Hour:
		return 24 * time.Hour
	case age >= 7*24*time.Hour:
		return 12 * time.Hour
	case age >= 24*time.Hour:
		return 6 * time.Hour
	case age >= 6*time.Hour:
		return 90 * time.Minute
	default:
		return 30 * time.Minute
	}
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

func shouldSkipMiddleReminder(ev *activity.Event, dueAt time.Time, deliveries []DeliveryHistoryRecord, lastUserMessageAt time.Time) bool {
	if ev == nil || len(deliveries) != 1 || lastUserMessageAt.IsZero() {
		return false
	}
	firstDeliveredAt := deliveries[0].DeliveredAt
	if !lastUserMessageAt.After(firstDeliveredAt) {
		return false
	}
	if dueAt.IsZero() {
		return true
	}
	return lastUserMessageAt.Before(dueAt)
}

func buildProgressSummary(ev *activity.Event, now, dueAt time.Time, reminderCount int) string {
	title := progressTopic(ev)
	if !dueAt.IsZero() && dueAt.After(now) {
		return fmt.Sprintf("%s 还有 %s 就到时间点了。", title, formatHumanDuration(dueAt.Sub(now)))
	}
	if !dueAt.IsZero() && !dueAt.After(now) {
		return fmt.Sprintf("%s 现在已经到时间点了。", title)
	}
	return fmt.Sprintf("%s 还没有收口。", title)
}

func buildProgressSuggestion(ev *activity.Event, now, dueAt time.Time, reminderCount int) string {
	title := progressTopic(ev)
	if !dueAt.IsZero() && dueAt.After(now) {
		remaining := formatHumanDuration(dueAt.Sub(now))
		if reminderCount == 0 {
			return fmt.Sprintf("记得在 %s 内把 %s 提前准备好。", remaining, title)
		}
		return fmt.Sprintf("离 %s 只剩 %s 了，最后再检查一下吧。", title, remaining)
	}
	if reminderCount == 0 {
		return fmt.Sprintf("要不要先把 %s 收一下？", title)
	}
	return fmt.Sprintf("%s 我就再提醒这一次，要不要顺手把它收掉？", title)
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

func formatHumanDuration(d time.Duration) string {
	if d <= 0 {
		return "现在"
	}
	if d >= 24*time.Hour {
		days := int(d / (24 * time.Hour))
		if days == 1 {
			return "1 天"
		}
		return fmt.Sprintf("%d 天", days)
	}
	if d >= time.Hour {
		hours := int(d / time.Hour)
		mins := int((d % time.Hour) / time.Minute)
		if mins == 0 {
			return fmt.Sprintf("%d 小时", hours)
		}
		return fmt.Sprintf("%d 小时 %d 分钟", hours, mins)
	}
	mins := int(d / time.Minute)
	if mins <= 0 {
		return "不到 1 分钟"
	}
	return fmt.Sprintf("%d 分钟", mins)
}
