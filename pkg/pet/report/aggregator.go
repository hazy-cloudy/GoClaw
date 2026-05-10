package report

import (
	"fmt"
	"sort"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
)

func AggregateWeeklyReport(characterID string, events []*activity.Event, now time.Time) WeeklyReport {
	end := now
	start := now.AddDate(0, 0, -7)

	report := WeeklyReport{
		ReportID:    fmt.Sprintf("weekly-%s", now.Format("20060102")),
		CharacterID: characterID,
		PeriodStart: start,
		PeriodEnd:   end,
	}

	categoryCounts := map[string]int{}
	outputs := make([]string, 0, 5)
	unfinished := make([]string, 0, 5)
	peakByDay := map[string]int{}
	peakByHour := map[int]int{}
	activeDays := map[string]struct{}{}
	sessions := map[string]struct{}{}

	for _, ev := range events {
		if ev == nil {
			continue
		}
		categoryCounts[string(ev.Category)]++
		dayKey := ev.CreatedAt.Format("2006-01-02")
		peakByDay[dayKey]++
		peakByHour[ev.CreatedAt.Hour()]++
		activeDays[dayKey] = struct{}{}
		if ev.SessionID != "" {
			sessions[ev.SessionID] = struct{}{}
		}
		if report.FirstActiveAt == nil || ev.CreatedAt.Before(*report.FirstActiveAt) {
			t := ev.CreatedAt
			report.FirstActiveAt = &t
		}
		if report.LastActiveAt == nil || ev.CreatedAt.After(*report.LastActiveAt) {
			t := ev.CreatedAt
			report.LastActiveAt = &t
		}

		switch ev.Type {
		case activity.EventUserMessage:
			report.MessageCount++
		case activity.EventToolCall:
			report.ToolCallCount++
		case activity.EventTaskResult:
			report.TaskCount++
			if ev.Status == activity.StatusDone {
				report.TaskDoneCount++
			}
			if ev.Status == activity.StatusFailed {
				report.FailureCount++
			}
			if ev.Status == activity.StatusPending || ev.Status == activity.StatusFailed {
				if ev.Title != "" {
					unfinished = appendIfMissing(unfinished, ev.Title, 5)
				}
			}
			if ev.Status == activity.StatusDone && ev.Title != "" {
				outputs = appendIfMissing(outputs, ev.Title, 5)
			}
		case activity.EventFileOutput:
			if ev.Title != "" {
				outputs = appendIfMissing(outputs, ev.Title, 5)
			}
		case activity.EventToolResult:
			if ev.Status == activity.StatusFailed {
				report.FailureCount++
				report.ToolErrorCount++
			}
		}
	}

	report.ActiveDays = len(activeDays)
	report.SessionCount = len(sessions)
	if report.TaskCount > 0 {
		report.CompletionRate = int(float64(report.TaskDoneCount) / float64(report.TaskCount) * 100)
	}
	report.TopCategories = topCategories(categoryCounts, 3)
	report.Outputs = outputs
	report.Unfinished = unfinished
	report.PeakDay = topDay(peakByDay)
	report.PeakHour = topHour(peakByHour)
	report.Summary = buildSummary(report)
	return report
}

func topCategories(counts map[string]int, limit int) []CategoryStat {
	stats := make([]CategoryStat, 0, len(counts))
	for name, count := range counts {
		if name == "" || count == 0 {
			continue
		}
		stats = append(stats, CategoryStat{Name: name, Count: count})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Name < stats[j].Name
		}
		return stats[i].Count > stats[j].Count
	})
	if len(stats) > limit {
		return stats[:limit]
	}
	return stats
}

func appendIfMissing(items []string, value string, limit int) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	if len(items) >= limit {
		return items
	}
	return append(items, value)
}

func topDay(counts map[string]int) string {
	best := ""
	bestCount := 0
	for day, count := range counts {
		if count > bestCount || (count == bestCount && day < best) {
			best = day
			bestCount = count
		}
	}
	return best
}

func topHour(counts map[int]int) int {
	bestHour := 0
	bestCount := -1
	for hour, count := range counts {
		if count > bestCount || (count == bestCount && hour < bestHour) {
			bestHour = hour
			bestCount = count
		}
	}
	return bestHour
}

func buildSummary(report WeeklyReport) string {
	if len(report.TopCategories) == 0 {
		return "这周活动还不多，不过已经开始积累记录啦，接下来我们可以一起把节奏慢慢跑起来。"
	}
	secondCategory := report.TopCategories[0].Name
	if len(report.TopCategories) > 1 {
		secondCategory = report.TopCategories[1].Name
	}
	return fmt.Sprintf(
		"这周一共推进了 %d 项任务，完成率约 %d%%，活跃了 %d 天，重点内容主要集中在 %s 和 %s。",
		report.TaskCount,
		report.CompletionRate,
		report.ActiveDays,
		report.TopCategories[0].Name,
		secondCategory,
	)
}
