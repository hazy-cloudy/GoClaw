package proactive

type ProgressNudgeProvider struct{}

func NewProgressNudgeProvider() *ProgressNudgeProvider {
	return &ProgressNudgeProvider{}
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

	return &Intent{
		Type:        "progress_nudge",
		Priority:    "low",
		ReasonCodes: []string{"unfinished_tasks", "user_idle"},
		Payload: map[string]any{
			"nudge_id":    snapshot.Now.Format("20060102150405"),
			"topic":       "未完成事项",
			"summary":     "你上次推进的事情里还有没收口的部分。",
			"suggestion":  "要不要先把上一次卡住的部分收一下？",
			"pending_cnt": snapshot.Activity.UnfinishedTaskCount,
		},
	}, true, nil
}
