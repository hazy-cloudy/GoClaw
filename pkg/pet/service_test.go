package pet

import (
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/memory"
)

// TestContainsIgnoreCase 测试 containsIgnoreCase 函数
func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty substr", "hello", "", true},
		{"exact match", "hello", "hello", true},
		{"case insensitive", "Hello", "hello", true},
		{"partial match", "Hello World", "world", true},
		{"no match", "hello", "xyz", false},
		{"shorter than substr", "hi", "hello", false},
		{"chinese content", "我喜欢音乐", "音乐", true},
		{"chinese case insensitive", "我喜欢音乐", "MUSIC", false}, // 中文没有大小写
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsIgnoreCase(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

// TestSortMemoriesByWeight 测试 sortMemoriesByWeight 函数
func TestSortMemoriesByWeight(t *testing.T) {
	memories := []*memory.Memory{
		{ID: 1, Weight: 30},
		{ID: 2, Weight: 90},
		{ID: 3, Weight: 50},
		{ID: 4, Weight: 80},
	}

	sortMemoriesByWeight(memories)

	expected := []int{90, 80, 50, 30}
	for i, m := range memories {
		if m.Weight != expected[i] {
			t.Errorf("memories[%d].Weight = %d, want %d", i, m.Weight, expected[i])
		}
	}
}

// TestSortMemoriesByWeight_Empty 测试空切片
func TestSortMemoriesByWeight_Empty(t *testing.T) {
	memories := []*memory.Memory{}
	sortMemoriesByWeight(memories)
	if len(memories) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(memories))
	}
}

// TestMemorySearchRequest 测试 MemorySearchRequest 类型
func TestMemorySearchRequest_Defaults(t *testing.T) {
	req := MemorySearchRequest{}

	if req.CharacterID != "" {
		t.Errorf("CharacterID should be empty by default")
	}
	if req.Keyword != "" {
		t.Errorf("Keyword should be empty by default")
	}
	if req.Type != "" {
		t.Errorf("Type should be empty by default")
	}
	if req.MinWeight != 0 {
		t.Errorf("MinWeight should be 0 by default")
	}
	if req.Limit != 0 {
		t.Errorf("Limit should be 0 by default")
	}
	if req.Offset != 0 {
		t.Errorf("Offset should be 0 by default")
	}
}

// TestMemoryItem 测试 MemoryItem 类型
func TestMemoryItem(t *testing.T) {
	item := MemoryItem{
		ID:        1,
		Type:      "conversation",
		Weight:    75,
		Content:   "测试内容",
		CreatedAt: "2024-01-15T10:30:00Z",
	}

	if item.ID != 1 {
		t.Errorf("ID = %d, want 1", item.ID)
	}
	if item.Type != "conversation" {
		t.Errorf("Type = %q, want %q", item.Type, "conversation")
	}
	if item.Weight != 75 {
		t.Errorf("Weight = %d, want 75", item.Weight)
	}
	if item.Content != "测试内容" {
		t.Errorf("Content = %q, want %q", item.Content, "测试内容")
	}
}

// TestMemorySearchResponse 测试 MemorySearchResponse 类型
func TestMemorySearchResponse(t *testing.T) {
	resp := MemorySearchResponse{
		Memories: []MemoryItem{
			{ID: 1, Type: "conversation", Weight: 75, Content: "测试"},
		},
		Total:   1,
		HasMore: false,
	}

	if len(resp.Memories) != 1 {
		t.Errorf("len(Memories) = %d, want 1", len(resp.Memories))
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
	if resp.HasMore != false {
		t.Error("HasMore should be false")
	}
}

// TestConversationListRequest 测试 ConversationListRequest 类型
func TestConversationListRequest_Defaults(t *testing.T) {
	req := ConversationListRequest{}

	if req.CharacterID != "" {
		t.Errorf("CharacterID should be empty by default")
	}
	if req.Limit != 0 {
		t.Errorf("Limit should be 0 by default")
	}
	if req.Offset != 0 {
		t.Errorf("Offset should be 0 by default")
	}
}

// TestConversationItem 测试 ConversationItem 类型
func TestConversationItem(t *testing.T) {
	item := ConversationItem{
		ID:         1,
		Role:       "user",
		Content:    "你好",
		Timestamp:  "2024-01-15T10:30:00Z",
		Compressed: false,
	}

	if item.ID != 1 {
		t.Errorf("ID = %d, want 1", item.ID)
	}
	if item.Role != "user" {
		t.Errorf("Role = %q, want %q", item.Role, "user")
	}
	if item.Compressed != false {
		t.Error("Compressed should be false")
	}
}

// TestConversationListResponse 测试 ConversationListResponse 类型
func TestConversationListResponse(t *testing.T) {
	resp := ConversationListResponse{
		Conversations: []ConversationItem{
			{ID: 1, Role: "user", Content: "你好"},
			{ID: 2, Role: "pet", Content: "你好！"},
		},
		Total:   2,
		HasMore: true,
	}

	if len(resp.Conversations) != 2 {
		t.Errorf("len(Conversations) = %d, want 2", len(resp.Conversations))
	}
	if resp.Total != 2 {
		t.Errorf("Total = %d, want 2", resp.Total)
	}
	if resp.HasMore != true {
		t.Error("HasMore should be true")
	}
}

// TestActionConstants 测试 Action 常量
func TestActionConstants(t *testing.T) {
	if ActionMemorySearch != "memory_search" {
		t.Errorf("ActionMemorySearch = %q, want %q", ActionMemorySearch, "memory_search")
	}
	if ActionConversationList != "conversation_list" {
		t.Errorf("ActionConversationList = %q, want %q", ActionConversationList, "conversation_list")
	}
}

// TestMemoryItemTimeFormat 测试时间格式
func TestMemoryItemTimeFormat(t *testing.T) {
	now := time.Now()
	item := MemoryItem{
		ID:        1,
		CreatedAt: now.Format("2006-01-02T15:04:05Z"),
	}

	// 验证时间格式可以被解析
	_, err := time.Parse("2006-01-02T15:04:05Z", item.CreatedAt)
	if err != nil {
		t.Errorf("CreatedAt time format is invalid: %v", err)
	}
}
