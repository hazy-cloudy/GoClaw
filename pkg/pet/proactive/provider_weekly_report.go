package proactive

import (
	"fmt"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
	"github.com/sipeed/picoclaw/pkg/pet/report"
)

type WeeklyReportProvider struct {
	activityStore *activity.Store
	stateStore    *WeeklyReportStateStore
}

func NewWeeklyReportProvider(activityStore *activity.Store, stateStore *WeeklyReportStateStore) *WeeklyReportProvider {
	return &WeeklyReportProvider{
		activityStore: activityStore,
		stateStore:    stateStore,
	}
}

func (p *WeeklyReportProvider) Name() string {
	return "weekly_report"
}

func (p *WeeklyReportProvider) Evaluate(snapshot Snapshot) (*Intent, bool, error) {
	if !snapshot.Preferences.WeeklyReportEnabled || p.activityStore == nil || p.stateStore == nil {
		return nil, false, nil
	}

	now := snapshot.Now
	year, week := now.ISOWeek()
	weekKey := formatISOWeekKey(year, week)
	state, err := p.stateStore.Load()
	if err != nil {
		return nil, false, err
	}
	if state.WeekKey == weekKey && state.DeliveredAt != nil {
		return nil, false, nil
	}

	if state.Ready && state.WeekKey == weekKey {
		if state.ExpireAt != nil && state.ExpireAt.Before(now) {
			// Let the report regenerate if the ready window expired.
		} else if snapshot.EvaluationReason != "user_message" {
			return nil, false, nil
		} else {
			return buildWeeklyReportIntentFromState(state), true, nil
		}
	}

	// 第一阶段：周三之后都允许准备，方便本地调试与验证链路。
	if now.Weekday() < time.Wednesday {
		return nil, false, nil
	}

	start := now.AddDate(0, 0, -7)
	events, err := p.activityStore.ListRange(start, now)
	if err != nil {
		return nil, false, err
	}
	if len(events) < 3 {
		return nil, false, nil
	}

	rep := report.AggregateWeeklyReport(snapshot.Pet.CharacterID, events, now)
	rendered := report.RenderWeeklySummary(rep, snapshot.Pet.PersonaType, snapshot.Pet.PersonalityTone, snapshot.Pet.DominantEmotion)
	nowCopy := now
	expire := now.Add(7 * 24 * time.Hour)
	reasonCodes := []string{"weekly_window", "enough_activity"}
	state = &WeeklyReportState{
		WeekKey:     weekKey,
		ReportID:    rep.ReportID,
		Ready:       true,
		GeneratedAt: &nowCopy,
		ExpireAt:    &expire,
		Title:       "本周陪跑回顾",
		Summary:     rendered,
		ReasonCodes: reasonCodes,
		Report:      &rep,
	}
	if err := p.stateStore.Save(state); err != nil {
		return nil, false, err
	}

	if snapshot.EvaluationReason != "user_message" {
		return nil, false, nil
	}

	return buildWeeklyReportIntentFromState(state), true, nil
}

func buildWeeklyReportIntentFromState(state *WeeklyReportState) *Intent {
	if state == nil || !state.Ready {
		return nil
	}
	payload := map[string]any{
		"report_id": state.ReportID,
		"title":     state.Title,
		"summary":   state.Summary,
	}
	if state.Report != nil {
		payload["report"] = *state.Report
	}
	return &Intent{
		Type:        "weekly_report",
		Priority:    "medium",
		ReasonCodes: append([]string{}, state.ReasonCodes...),
		Payload:     payload,
	}
}

func formatISOWeekKey(year, week int) string {
	return fmt.Sprintf("%04d-W%02d", year, week)
}
