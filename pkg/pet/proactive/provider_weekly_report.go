package proactive

import (
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
	weekKey := now.Format("2006-W01")
	state, err := p.stateStore.Load()
	if err != nil {
		return nil, false, err
	}
	if state.WeekKey == weekKey && state.DeliveredAt != nil {
		return nil, false, nil
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
	_ = p.stateStore.Save(&WeeklyReportState{
		WeekKey:     weekKey,
		ReportID:    rep.ReportID,
		Ready:       true,
		GeneratedAt: &nowCopy,
		ExpireAt:    &expire,
	})

	return &Intent{
		Type:        "weekly_report",
		Priority:    "medium",
		ReasonCodes: []string{"weekly_window", "enough_activity"},
		Payload: map[string]any{
			"report_id": rep.ReportID,
			"title":     "本周陪跑回顾",
			"summary":   rendered,
			"report":    rep,
		},
	}, true, nil
}
