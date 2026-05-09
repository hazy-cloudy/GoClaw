package proactive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/report"
)

type WeeklyReportState struct {
	WeekKey     string             `json:"week_key"`
	ReportID    string             `json:"report_id"`
	Ready       bool               `json:"ready"`
	GeneratedAt *time.Time         `json:"generated_at,omitempty"`
	DeliveredAt *time.Time         `json:"delivered_at,omitempty"`
	ExpireAt    *time.Time         `json:"expire_at,omitempty"`
	Title       string             `json:"title,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	ReasonCodes []string           `json:"reason_codes,omitempty"`
	Report      *report.WeeklyReport `json:"report,omitempty"`
}

type WeeklyReportStateStore struct {
	path string
	mu   sync.Mutex
}

func NewWeeklyReportStateStore(workspacePath string) *WeeklyReportStateStore {
	return &WeeklyReportStateStore{
		path: filepath.Join(workspacePath, "pet_weekly_report_state.json"),
	}
}

func (s *WeeklyReportStateStore) Load() (*WeeklyReportState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &WeeklyReportState{}, nil
		}
		return nil, fmt.Errorf("read weekly report state: %w", err)
	}
	if len(data) == 0 {
		return &WeeklyReportState{}, nil
	}
	var state WeeklyReportState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal weekly report state: %w", err)
	}
	return &state, nil
}

func (s *WeeklyReportStateStore) Save(state *WeeklyReportState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state == nil {
		return fmt.Errorf("weekly report state is nil")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create weekly report state dir: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal weekly report state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write weekly report state: %w", err)
	}
	return nil
}
