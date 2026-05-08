package proactive

import (
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryStoreAppendAndList(t *testing.T) {
	store := NewHistoryStore(t.TempDir())
	rec := DeliveryHistoryRecord{
		EventType:   "weekly_report",
		EventID:     "report-1",
		CharacterID: "pet_001",
		DeliveredAt: time.Now(),
	}
	if err := store.Append(rec); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	items, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
}

func TestNewHistoryStorePath(t *testing.T) {
	base := t.TempDir()
	store := NewHistoryStore(base)
	want := filepath.Join(base, "pet_proactive_history.json")
	if store.path != want {
		t.Fatalf("path = %q, want %q", store.path, want)
	}
}
