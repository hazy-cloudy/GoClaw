package activity

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreAppendAndListRange(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	now := time.Date(2026, 5, 8, 10, 0, 0, 0, time.Local)
	event := &Event{
		ID:          "ev-1",
		CharacterID: "pet_001",
		SessionID:   "session-1",
		Type:        EventUserMessage,
		Category:    CategoryCode,
		Status:      StatusDone,
		Title:       "修复打包问题",
		Summary:     "修复打包问题",
		CreatedAt:   now,
	}

	if err := store.Append(event); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	events, err := store.ListRange(now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("ListRange() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].ID != "ev-1" {
		t.Fatalf("events[0].ID = %q", events[0].ID)
	}
}

func TestStoreMonthFile(t *testing.T) {
	base := t.TempDir()
	store := &Store{baseDir: base}
	ts := time.Date(2026, 5, 8, 0, 0, 0, 0, time.Local)
	want := filepath.Join(base, "2026-05.jsonl")
	if got := store.monthFile(ts); got != want {
		t.Fatalf("monthFile() = %q, want %q", got, want)
	}
}
