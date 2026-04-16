package compression

import (
	"testing"

	"github.com/sipeed/picoclaw/pkg/pet/memory"
)

func TestDefaultCompressionConfig(t *testing.T) {
	cfg := DefaultCompressionConfig()

	if cfg.Model != "minimax" {
		t.Errorf("Model=%q, want %q", cfg.Model, "minimax")
	}
	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.Threshold != 20 {
		t.Errorf("Threshold=%d, want 20", cfg.Threshold)
	}
	if cfg.MergeIntervalMinutes != 5 {
		t.Errorf("MergeIntervalMinutes=%d, want 5", cfg.MergeIntervalMinutes)
	}
	if cfg.LowWeightThreshold != 10 {
		t.Errorf("LowWeightThreshold=%d, want 10", cfg.LowWeightThreshold)
	}
	if cfg.RetainCompressedDays != 30 {
		t.Errorf("RetainCompressedDays=%d, want 30", cfg.RetainCompressedDays)
	}
}

func TestCompressionConfig_Defaults(t *testing.T) {
	cfg := &CompressionConfig{}

	if cfg.Model != "" {
		t.Errorf("default Model=%q, want empty", cfg.Model)
	}
	if cfg.Enabled {
		t.Error("default Enabled should be false")
	}
	if cfg.Threshold != 0 {
		t.Errorf("default Threshold=%d, want 0", cfg.Threshold)
	}
}

func TestConversationEntry(t *testing.T) {
	entry := &ConversationEntry{
		ID:      123,
		Role:    "user",
		Content: "测试内容",
	}

	if entry.ID != 123 {
		t.Errorf("ID=%d, want 123", entry.ID)
	}
	if entry.Role != "user" {
		t.Errorf("Role=%q, want %q", entry.Role, "user")
	}
	if entry.Content != "测试内容" {
		t.Errorf("Content=%q, want %q", entry.Content, "测试内容")
	}
}

func TestCompressionResult(t *testing.T) {
	result := &CompressionResult{
		MemoryType: "conversation",
		Weight:     75,
		Content:    "测试记忆内容",
	}

	if result.MemoryType != "conversation" {
		t.Errorf("MemoryType=%q, want %q", result.MemoryType, "conversation")
	}
	if result.Weight != 75 {
		t.Errorf("Weight=%d, want 75", result.Weight)
	}
	if result.Content != "测试记忆内容" {
		t.Errorf("Content=%q, want %q", result.Content, "测试记忆内容")
	}
}

func TestParseCompressionResult(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	result := svc.parseCompressionResult("[memory_type:conversation-weight:85-content:测试摘要]")
	if result == nil {
		t.Fatal("parseCompressionResult returned nil")
	}

	if result.MemoryType != "conversation" {
		t.Errorf("MemoryType=%q, want %q", result.MemoryType, "conversation")
	}
	if result.Weight != 85 {
		t.Errorf("Weight=%d, want 85", result.Weight)
	}
	if result.Content != "测试摘要" {
		t.Errorf("Content=%q, want %q", result.Content, "测试摘要")
	}
}

func TestParseCompressionResult_Invalid(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	result := svc.parseCompressionResult("invalid format")
	if result != nil {
		t.Error("parseCompressionResult should return nil for invalid format")
	}
}

func TestParseMultipleCompressionResults(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	content := `[memory_type:conversation-weight:85-content:摘要1]
[memory_type:preference-weight:70-content:摘要2]
[memory_type:fact-weight:90-content:摘要3]`

	results := svc.parseMultipleCompressionResults(content)
	if len(results) != 3 {
		t.Errorf("len(results)=%d, want 3", len(results))
	}

	if results[0].MemoryType != "conversation" {
		t.Errorf("results[0].MemoryType=%q, want %q", results[0].MemoryType, "conversation")
	}
	if results[1].MemoryType != "preference" {
		t.Errorf("results[1].MemoryType=%q, want %q", results[1].MemoryType, "preference")
	}
	if results[2].MemoryType != "fact" {
		t.Errorf("results[2].MemoryType=%q, want %q", results[2].MemoryType, "fact")
	}
}

func TestCalculateStats(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	memories := []*memory.Memory{
		{Weight: 50},
		{Weight: 60},
		{Weight: 70},
		{Weight: 80},
		{Weight: 90},
	}

	avg, stddev := svc.calculateStats(memories)

	expectedAvg := 70.0
	if avg != expectedAvg {
		t.Errorf("avg=%f, want %f", avg, expectedAvg)
	}

	expectedStddev := 14.142
	if stddev < expectedStddev-0.1 || stddev > expectedStddev+0.1 {
		t.Errorf("stddev=%f, want ~%f", stddev, expectedStddev)
	}
}

func TestCalculateStats_Empty(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	avg, stddev := svc.calculateStats([]*memory.Memory{})
	if avg != 0 {
		t.Errorf("avg=%f, want 0", avg)
	}
	if stddev != 0 {
		t.Errorf("stddev=%f, want 0", stddev)
	}
}

func TestBuildCompressionPrompt(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	entries := []*ConversationEntry{
		{Role: "user", Content: "用户消息"},
		{Role: "pet", Content: "宠物回复"},
	}

	prompt := svc.buildCompressionPrompt(entries)

	if prompt == "" {
		t.Fatal("buildCompressionPrompt returned empty string")
	}

	if !contains(prompt, "用户消息") {
		t.Error("prompt should contain user message")
	}
	if !contains(prompt, "宠物回复") {
		t.Error("prompt should contain pet message")
	}
	if !contains(prompt, "[memory_type:") {
		t.Error("prompt should contain output format")
	}
}

func TestBuildMergePrompt(t *testing.T) {
	cfg := &CompressionConfig{}
	svc := &CompressionService{config: cfg}

	memories := []*memory.Memory{
		{MemoryType: "conversation", Weight: 30, Content: "碎片1"},
		{MemoryType: "conversation", Weight: 25, Content: "碎片2"},
	}

	prompt := svc.buildMergePrompt(memories)

	if prompt == "" {
		t.Fatal("buildMergePrompt returned empty string")
	}

	if !contains(prompt, "碎片1") {
		t.Error("prompt should contain memory content 1")
	}
	if !contains(prompt, "碎片2") {
		t.Error("prompt should contain memory content 2")
	}
	if !contains(prompt, "weight=30") {
		t.Error("prompt should contain weight")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
