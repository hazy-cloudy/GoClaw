package report

import "time"

type CategoryStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type WeeklyReport struct {
	ReportID      string         `json:"report_id"`
	CharacterID   string         `json:"character_id"`
	PeriodStart   time.Time      `json:"period_start"`
	PeriodEnd     time.Time      `json:"period_end"`
	MessageCount  int            `json:"message_count"`
	TaskCount     int            `json:"task_count"`
	ToolCallCount int            `json:"tool_call_count"`
	FailureCount  int            `json:"failure_count"`
	TopCategories []CategoryStat `json:"top_categories"`
	Outputs       []string       `json:"outputs"`
	Unfinished    []string       `json:"unfinished"`
	PeakDay       string         `json:"peak_day"`
	PeakHour      int            `json:"peak_hour"`
	Summary       string         `json:"summary"`
}
