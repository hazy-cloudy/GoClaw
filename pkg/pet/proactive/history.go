package proactive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type HistoryStore struct {
	path string
	mu   sync.Mutex
}

func NewHistoryStore(workspacePath string) *HistoryStore {
	return &HistoryStore{
		path: filepath.Join(workspacePath, "pet_proactive_history.json"),
	}
}

func (s *HistoryStore) loadLocked() ([]DeliveryHistoryRecord, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read proactive history: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	var records []DeliveryHistoryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("unmarshal proactive history: %w", err)
	}
	return records, nil
}

func (s *HistoryStore) List() ([]DeliveryHistoryRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *HistoryStore) Append(record DeliveryHistoryRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, err := s.loadLocked()
	if err != nil {
		return err
	}
	records = append(records, record)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal proactive history: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write proactive history: %w", err)
	}
	return nil
}

func (s *HistoryStore) FindByEvent(eventType, eventID, characterID, sessionID string) ([]DeliveryHistoryRecord, error) {
	records, err := s.List()
	if err != nil {
		return nil, err
	}
	filtered := make([]DeliveryHistoryRecord, 0, len(records))
	for _, record := range records {
		if eventType != "" && record.EventType != eventType {
			continue
		}
		if eventID != "" && record.EventID != eventID {
			continue
		}
		if characterID != "" && record.CharacterID != characterID {
			continue
		}
		if sessionID != "" && record.SessionID != "" && record.SessionID != sessionID {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered, nil
}
