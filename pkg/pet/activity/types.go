package activity

import "time"

// EventType 表示一条“桌宠活动记录”的来源类型。
// 这层数据不是给用户直接展示的聊天记录，而是后续做周报/催办的原始素材。
type EventType string

const (
	EventUserMessage EventType = "user_message"
	EventToolCall    EventType = "tool_call"
	EventToolResult  EventType = "tool_result"
	EventFileOutput  EventType = "file_output"
	EventTaskResult  EventType = "task_result"
)

type Category string

const (
	CategoryCode   Category = "code"
	CategoryDoc    Category = "doc"
	CategoryDebug  Category = "debug"
	CategoryPPTX   Category = "pptx"
	CategoryConfig Category = "config"
	CategoryOther  Category = "other"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

// Event 是主动性系统使用的最小活动单元。
// 可以把它理解为“桌宠刚刚参与过的一件事”的结构化记录。
//
// 设计上刻意不追求一开始就把所有信息都存进去，而是先保留：
// 1. 这是什么事（Type / Category）
// 2. 结果怎么样（Status）
// 3. 跟谁、在哪个会话里发生（CharacterID / SessionID）
// 4. 一句人类可读的摘要（Title / Summary）
type Event struct {
	ID          string         `json:"id"`
	CharacterID string         `json:"character_id"`
	SessionID   string         `json:"session_id"`
	Type        EventType      `json:"type"`
	Category    Category       `json:"category"`
	Status      Status         `json:"status"`
	Title       string         `json:"title"`
	Summary     string         `json:"summary"`
	ToolName    string         `json:"tool_name,omitempty"`
	FilePaths   []string       `json:"file_paths,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}
