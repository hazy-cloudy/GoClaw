package memory

import (
	"os"
	"testing"
)

func TestNewStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("NewStore returned nil")
	}
}

func TestStore_Add(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	memory, err := store.Add("pet_001", "测试记忆内容", MemoryTypeConversation, 75)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if memory.ID <= 0 {
		t.Errorf("Memory.ID=%d, want > 0", memory.ID)
	}
	if memory.Content != "测试记忆内容" {
		t.Errorf("Memory.Content=%q, want %q", memory.Content, "测试记忆内容")
	}
	if memory.MemoryType != MemoryTypeConversation {
		t.Errorf("Memory.MemoryType=%q, want %q", memory.MemoryType, MemoryTypeConversation)
	}
	if memory.Weight != 75 {
		t.Errorf("Memory.Weight=%d, want 75", memory.Weight)
	}
}

func TestStore_List(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	_, err = store.Add("pet_001", "记忆1", MemoryTypeConversation, 50)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	_, err = store.Add("pet_001", "记忆2", MemoryTypePreference, 70)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	memories, err := store.List("pet_001")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("len(memories)=%d, want 2", len(memories))
	}
}

func TestStore_ListByType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	_, err = store.Add("pet_001", "对话记忆", MemoryTypeConversation, 50)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	_, err = store.Add("pet_001", "偏好记忆", MemoryTypePreference, 70)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	memories, err := store.ListByType("pet_001", MemoryTypeConversation)
	if err != nil {
		t.Fatalf("ListByType failed: %v", err)
	}

	if len(memories) != 1 {
		t.Errorf("len(memories)=%d, want 1", len(memories))
	}
	if memories[0].Content != "对话记忆" {
		t.Errorf("memories[0].Content=%q, want %q", memories[0].Content, "对话记忆")
	}
}

func TestStore_ListByWeight(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	_, _ = store.Add("pet_001", "低权重", MemoryTypeConversation, 30)
	_, _ = store.Add("pet_001", "高权重", MemoryTypeConversation, 90)

	memories, err := store.ListByWeight("pet_001", 10)
	if err != nil {
		t.Fatalf("ListByWeight failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("len(memories)=%d, want 2", len(memories))
	}
	if memories[0].Weight < memories[1].Weight {
		t.Error("memories should be sorted by weight DESC")
	}
}

func TestStore_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	memory, _ := store.Add("pet_001", "待删除", MemoryTypeConversation, 50)

	err = store.Delete(memory.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	memories, _ := store.List("pet_001")
	if len(memories) != 0 {
		t.Errorf("len(memories)=%d, want 0", len(memories))
	}
}

func TestStore_DeleteByIDs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	m1, _ := store.Add("pet_001", "记忆1", MemoryTypeConversation, 50)
	m2, _ := store.Add("pet_001", "记忆2", MemoryTypeConversation, 50)

	err = store.DeleteByIDs([]int64{m1.ID, m2.ID})
	if err != nil {
		t.Fatalf("DeleteByIDs failed: %v", err)
	}

	memories, _ := store.List("pet_001")
	if len(memories) != 0 {
		t.Errorf("len(memories)=%d, want 0", len(memories))
	}
}

func TestStore_UpdateWeight(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	memory, _ := store.Add("pet_001", "测试", MemoryTypeConversation, 50)

	err = store.UpdateWeight(memory.ID, 85)
	if err != nil {
		t.Fatalf("UpdateWeight failed: %v", err)
	}

	memories, _ := store.List("pet_001")
	if memories[0].Weight != 85 {
		t.Errorf("memories[0].Weight=%d, want 85", memories[0].Weight)
	}
}

func TestStore_DeleteCharacter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	_, _ = store.Add("pet_001", "记忆1", MemoryTypeConversation, 50)
	_, _ = store.Add("pet_001", "记忆2", MemoryTypeConversation, 50)
	_, _ = store.Add("pet_002", "其他角色记忆", MemoryTypeConversation, 50)

	err = store.DeleteCharacter("pet_001")
	if err != nil {
		t.Fatalf("DeleteCharacter failed: %v", err)
	}

	memories, _ := store.List("pet_001")
	if len(memories) != 0 {
		t.Errorf("len(memories)=%d, want 0", len(memories))
	}

	otherMemories, _ := store.List("pet_002")
	if len(otherMemories) != 1 {
		t.Errorf("len(otherMemories)=%d, want 1", len(otherMemories))
	}
}

func TestStore_ListBelowWeight(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	_, _ = store.Add("pet_001", "低权重1", MemoryTypeConversation, 20)
	_, _ = store.Add("pet_001", "低权重2", MemoryTypeConversation, 30)
	_, _ = store.Add("pet_001", "高权重", MemoryTypeConversation, 80)

	memories, err := store.ListBelowWeight("pet_001", 50)
	if err != nil {
		t.Fatalf("ListBelowWeight failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("len(memories)=%d, want 2", len(memories))
	}
}
