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
	ActiveDays    int            `json:"active_days"`
	SessionCount  int            `json:"session_count"`
	MessageCount  int            `json:"message_count"`
	TaskCount     int            `json:"task_count"`
	TaskDoneCount int            `json:"task_done_count"`
	ToolCallCount int            `json:"tool_call_count"`
	ToolErrorCount int           `json:"tool_error_count"`
	FailureCount  int            `json:"failure_count"`
	CompletionRate int           `json:"completion_rate"`
	TopCategories []CategoryStat `json:"top_categories"`
	Outputs       []string       `json:"outputs"`
	Unfinished    []string       `json:"unfinished"`
	PeakDay       string         `json:"peak_day"`
	PeakHour      int            `json:"peak_hour"`
	FirstActiveAt *time.Time     `json:"first_active_at,omitempty"`
	LastActiveAt  *time.Time     `json:"last_active_at,omitempty"`
	Summary       string         `json:"summary"`
}
