package activity

import (
	"testing"

	"github.com/sipeed/picoclaw/pkg/tools"
)

func TestClassifyText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Category
	}{
		{"ppt", "please build a presentation deck", CategoryPPTX},
		{"debug", "修复这个报错并排查原因", CategoryDebug},
		{"doc", "更新一下 README 文档", CategoryDoc},
		{"config", "帮我整理配置文件", CategoryConfig},
		{"code", "请修复这段 Go code", CategoryCode},
		{"other", "随便聊聊", CategoryOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyText(tt.text); got != tt.want {
				t.Fatalf("ClassifyText(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestBuildUserMessageEvent(t *testing.T) {
	event := BuildUserMessageEvent("pet_001", "session-1", "请帮我生成本周周报")
	if event.CharacterID != "pet_001" {
		t.Fatalf("CharacterID = %q", event.CharacterID)
	}
	if event.SessionID != "session-1" {
		t.Fatalf("SessionID = %q", event.SessionID)
	}
	if event.Type != EventUserMessage {
		t.Fatalf("Type = %q", event.Type)
	}
	if event.Category != CategoryOther {
		t.Fatalf("Category = %q, want %q", event.Category, CategoryOther)
	}
	if event.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set")
	}
}

func TestBuildToolResultEventMarksFailure(t *testing.T) {
	event := BuildToolResultEvent("pet_001", "session-1", "run_tests", &tools.ToolResult{
		ForLLM:  "tests failed",
		IsError: true,
	})
	if event.Type != EventToolResult {
		t.Fatalf("Type = %q", event.Type)
	}
	if event.Status != StatusFailed {
		t.Fatalf("Status = %q, want %q", event.Status, StatusFailed)
	}
	if event.Category != CategoryOther {
		t.Fatalf("Category = %q, want %q", event.Category, CategoryOther)
	}
}
