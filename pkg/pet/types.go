package pet

import "encoding/json"

// =============================================================================
// 请求类型定义
// =============================================================================

// Request 客户端请求结构
// 客户端发送的所有请求都基于此结构
type Request struct {
	Action    string          `json:"action"`               // 操作类型，必填
	Data      json.RawMessage `json:"data"`                 // 请求数据
	RequestID string          `json:"request_id,omitempty"` // 请求ID，用于追踪响应
}

// =============================================================================
// Action 常量定义
// =============================================================================

// Action 请求动作类型
const (
	ActionChat             = "chat"              // 聊天交互
	ActionVoice            = "voice"             // 语音交互
	ActionOnboardingStart  = "onboarding_start"  // 查询是否需要初始化
	ActionOnboardingConfig = "onboarding_config" // 提交配置
	ActionCharacterGet     = "character_get"     // 获取角色配置
	ActionCharacterUpdate  = "character_update"  // 更新角色配置
	ActionConfigGet        = "config_get"        // 获取应用配置
	ActionConfigUpdate     = "config_update"     // 更新应用配置
	ActionEmotionGet       = "emotion_get"       // 获取情绪状态
	ActionEventAction      = "event_action"      // 触发事件动作
	ActionHealthCheck      = "health_check"      // 健康检查
	ActionMemoryGet        = "memory_get"        // 获取记忆记录
	ActionMemoryTypeAdd    = "memory_type_add"   // 添加记忆类型
	ActionMemoryOldGet     = "memory_old_get"    // 获取旧记忆摘要
	ActionDynamicGet       = "dynamic_get"       // 获取动态
)

// =============================================================================
// PushType 常量定义
// =============================================================================

// PushType 推送类型
const (
	PushTypeAIChat           = "ai_chat"           // AI 聊天回复
	PushTypeAIVoice          = "ai_voice"          // AI 语音回复
	PushTypeEmotionChange    = "emotion_change"    // 情绪变化
	PushTypeReminderTrigger  = "reminder_trigger"  // 提醒触发
	PushTypeOnboardingConfig = "onboarding_config" // 引导配置推送
	PushTypeHeartbeat        = "heartbeat"         // 心跳保活
	PushTypeSystemStatus     = "system_status"     // 系统状态
	PushTypeDynamicAdd       = "dynamic_add"       // 动态添加
)

// =============================================================================
// Status 常量定义
// =============================================================================

// Status 响应状态
const (
	StatusOK      = "ok"      // 成功
	StatusError   = "error"   // 错误
	StatusPending = "pending" // 处理中
)

// =============================================================================
// 响应结构定义
// =============================================================================

// Response 服务端响应结构
// 服务端对客户端请求的响应格式
type Response struct {
	Status    string          `json:"status"`               // 状态：ok/error/pending
	Action    string          `json:"action,omitempty"`     // 对应的 action
	Data      json.RawMessage `json:"data,omitempty"`       // 响应数据
	Error     string          `json:"error,omitempty"`      // 错误信息
	RequestID string          `json:"request_id,omitempty"` // 对应的请求ID
}

// Push 推送消息结构
// 服务端主动推送的消息格式
type Push struct {
	Type      string          `json:"type"`      // 固定为 "push"
	PushType  string          `json:"push_type"` // 推送类型
	Data      json.RawMessage `json:"data"`      // 推送数据
	Timestamp int64           `json:"timestamp"` // 时间戳
}

// =============================================================================
// 请求数据定义
// =============================================================================

// ChatRequest 聊天请求数据
// 用户发送文本消息给 AI
type ChatRequest struct {
	Text       string `json:"text"`        // 用户输入的文本内容
	SessionKey string `json:"session_key"` // 会话标识符，格式：pet:{pet_id}:{user_id}
}

// VoiceRequest 语音请求数据
// 用户发送语音消息（语音转文字后处理）
type VoiceRequest struct {
	Text       string `json:"text"`        // 语音识别后的文本内容
	SessionKey string `json:"session_key"` // 会话标识符
}

// OnboardingConfigRequest 引导配置请求数据
// 用户与桌宠的初始化配置
type OnboardingConfigRequest struct {
	UserName          string `json:"user_name"`           // 用户名称
	UserSex           string `json:"user_sex"`            // 用户性别
	PetName           string `json:"pet_name"`            // 桌宠名称
	PetSex            string `json:"pet_sex"`             // 桌宠性别
	PetPersonaContext string `json:"pet_persona_context"` // 详细性格描述
	PetPersonaType    string `json:"pet_persona_type"`    // 性格类型ID
}

// CharacterGetRequest 角色获取请求数据
type CharacterGetRequest struct {
	PetID string `json:"pet_id"` // 桌宠ID
}

// CharacterUpdateRequest 角色更新请求数据
type CharacterUpdateRequest struct {
	PetID          string      `json:"pet_id"`           // 桌宠ID
	PetName        string      `json:"pet_name"`         // 桌宠名称
	PetPersona     string      `json:"pet_persona"`      // 详细性格描述
	PetPersonaType string      `json:"pet_persona_type"` // 性格类型ID
	PetMBTI        *MBTIConfig `json:"pet_mbti"`         // MBTI 配置（可选）
}

// ConfigUpdateRequest 配置更新请求数据
type ConfigUpdateRequest struct {
	EmotionEnabled           *bool   `json:"emotion_enabled"`            // 是否启用情绪表情
	ReminderEnabled          *bool   `json:"reminder_enabled"`           // 是否启用提醒
	ProactiveCare            *bool   `json:"proactive_care"`             // 是否启用主动关怀
	ProactiveIntervalMinutes *int    `json:"proactive_interval_minutes"` // 主动关怀间隔（分钟）
	VoiceEnabled             *bool   `json:"voice_enabled"`              // 是否启用语音
	Language                 *string `json:"language"`                   // 语言设置
}

// EmotionGetRequest 情绪获取请求数据
type EmotionGetRequest struct {
	PetID string `json:"pet_id"` // 桌宠ID
}

// EventActionRequest 事件动作请求数据
type EventActionRequest struct {
	Event  string                 `json:"event"`  // 事件类型
	Params map[string]interface{} `json:"params"` // 事件参数
}

// MemoryGetRequest 记忆获取请求数据
type MemoryGetRequest struct {
	PetID string `json:"pet_id"` // 桌宠ID
	Limit int    `json:"limit"`  // 获取数量限制
}

// MemoryTypeAddRequest 记忆类型添加请求数据
type MemoryTypeAddRequest struct {
	PetID string `json:"pet_id"` // 桌宠ID
	Type  string `json:"type"`   // 记忆类型 (text/voice)
}

// MemoryOldGetRequest 旧记忆获取请求数据
type MemoryOldGetRequest struct {
	PetID string `json:"pet_id"` // 桌宠ID
	Limit int    `json:"limit"`  // 获取数量限制
}

// DynamicGetRequest 动态获取请求数据
type DynamicGetRequest struct {
	PetID string `json:"pet_id"` // 桌宠ID
	Limit int    `json:"limit"`  // 获取数量限制
}

// =============================================================================
// 响应数据定义
// =============================================================================

// OnboardingConfigResponse 引导配置响应数据
type OnboardingConfigResponse struct {
	PetID string `json:"pet_id"` // 桌宠ID
}

// OnboardingConfigPush 引导配置推送数据
type OnboardingConfigPush struct {
	Text string `json:"text"` // 欢迎文本
}

// CharacterConfig 角色配置
// 桌宠的完整角色配置信息
type CharacterConfig struct {
	PetID          string     `json:"pet_id"`           // 桌宠ID
	PetName        string     `json:"pet_name"`         // 桌宠名称
	PetPersona     string     `json:"pet_persona"`      // 详细性格描述
	PetPersonaType string     `json:"pet_persona_type"` // 性格类型ID
	PetMBTI        MBTIConfig `json:"pet_mbti"`         // MBTI 性格配置
	Avatar         string     `json:"avatar,omitempty"` // 头像/模型ID
	CreatedAt      string     `json:"created_at"`       // 创建时间 (RFC3339)
	UpdatedAt      string     `json:"updated_at"`       // 更新时间 (RFC3339)
}

// MBTIConfig MBTI 性格四维配置
// 用于定义角色的性格倾向
type MBTIConfig struct {
	IE int `json:"ie"` // 内向(I)-外向(E): 0-100, 50中性, >50偏外向, <50偏内向
	SN int `json:"sn"` // 实感(S)-直觉(N): 0-100, 50中性, >50偏实感, <50偏直觉
	TF int `json:"tf"` // 理性(T)-感性(F): 0-100, 50中性, >50偏理性, <50偏感性
	JP int `json:"jp"` // 判断(J)-感知(P): 0-100, 50中性, >50偏判断, <50偏感知
}

// DefaultMBTI 返回默认的中性 MBTI 配置
func DefaultMBTI() MBTIConfig {
	return MBTIConfig{IE: 50, SN: 50, TF: 50, JP: 50}
}

// AppConfig 应用配置
// 应用的功能开关和设置
type AppConfig struct {
	EmotionEnabled           bool   `json:"emotion_enabled"`            // 是否启用情绪表情
	ReminderEnabled          bool   `json:"reminder_enabled"`           // 是否启用提醒功能
	ProactiveCare            bool   `json:"proactive_care"`             // 是否启用主动关怀
	ProactiveIntervalMinutes int    `json:"proactive_interval_minutes"` // 主动关怀间隔（分钟）
	VoiceEnabled             bool   `json:"voice_enabled"`              // 是否启用语音
	Language                 string `json:"language"`                   // 语言设置
}

// ConfigGetResponse config_get 响应数据 (文档3.4.1，不含 emotion_enabled)
type ConfigGetResponse struct {
	ReminderEnabled          bool   `json:"reminder_enabled"`
	ProactiveCare            bool   `json:"proactive_care"`
	ProactiveIntervalMinutes int    `json:"proactive_interval_minutes"`
	VoiceEnabled             bool   `json:"voice_enabled"`
	Language                 string `json:"language"`
}

// EmotionState 情绪状态
// 桌宠当前的情绪状态
type EmotionState struct {
	PetID       string  `json:"pet_id"`      // 桌宠ID
	Emotion     string  `json:"emotion"`     // 情绪标签 (neutral/happy/sad/angry/worried/love/surprised)
	Score       float64 `json:"score"`       // 情绪强度 (0.0-1.0)
	Expression  string  `json:"expression"`  // Live2D 表情文件 ID
	Motion      string  `json:"motion"`      // Live2D 动作名称
	Description string  `json:"description"` // 情绪中文描述
}

// AIChatResponse AI聊天响应数据
// AI 对话回复的推送数据
type AIChatResponse struct {
	Text     string `json:"text"`      // AI 回复文本
	Emotion  string `json:"emotion"`   // 情绪标签
	Action   string `json:"action"`    // Live2D 动作名称
	VoiceURL string `json:"voice_url"` // 语音URL（仅语音回复有值）
}

// AIVoiceResponse AI语音响应数据
type AIVoiceResponse struct {
	Text     string `json:"text"`      // AI 回复文本
	Emotion  string `json:"emotion"`   // 情绪标签
	Action   string `json:"action"`    // Live2D 动作名称
	VoiceURL string `json:"voice_url"` // 语音URL
}

// ReminderTrigger 提醒触发推送数据
type ReminderTrigger struct {
	ReminderID string `json:"reminder_id"` // 提醒ID
	Content    string `json:"content"`     // 提醒内容
	Timestamp  int64  `json:"timestamp"`   // 触发时间戳
}

// DynamicItem 动态项
// 桌宠发布的动态
type DynamicItem struct {
	ID        string `json:"id"`        // 动态ID
	PetID     string `json:"pet_id"`    // 桌宠ID
	UserID    string `json:"user_id"`   // 用户ID
	Text      string `json:"text"`      // 动态文本
	Picture   string `json:"picture"`   // 动态图片URL
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// MemoryItem 记忆项
// 桌宠的记忆记录
type MemoryItem struct {
	ID        string `json:"id"`        // 记忆ID
	PetID     string `json:"pet_id"`    // 桌宠ID
	UserID    string `json:"user_id"`   // 用户ID
	Type      string `json:"type"`      // 记忆类型 (text/voice)
	Text      string `json:"text"`      // 记忆文本
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// MemorySummary 记忆摘要
// 旧记忆的摘要信息
type MemorySummary struct {
	PetID   string `json:"pet_id"`  // 桌宠ID
	Summary string `json:"summary"` // 记忆摘要文本
}

// Option 选项
// 用于引导流程中的选择项（旧版，保留兼容性）
type Option struct {
	ID   string `json:"id"`   // 选项ID
	Name string `json:"name"` // 选项名称
	Icon string `json:"icon"` // 选项图标
}

// OnboardingStep 引导步骤数据（旧版，保留兼容性）
type OnboardingStep struct {
	Step        int              `json:"step"`
	Total       int              `json:"total"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Type        string           `json:"type"`
	InputHint   string           `json:"input_hint,omitempty"`
	Options     []Option         `json:"options,omitempty"`
	Suggestions []string         `json:"suggestions,omitempty"`
	Character   *CharacterConfig `json:"character,omitempty"`
}

// OnboardingData 引导配置数据（旧版，保留兼容性）
type OnboardingData struct {
	Name        string `json:"name,omitempty"`
	Persona     string `json:"persona,omitempty"`
	PersonaType string `json:"persona_type,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

// DynamicAddRequest 动态添加请求数据 (文档3.9.2)
type DynamicAddRequest struct {
	PetID   string `json:"pet_id"`  // 桌宠ID
	UserID  string `json:"user_id"` // 用户ID
	Text    string `json:"text"`    // 动态文本
	Picture string `json:"picture"` // 动态图片URL
}

// HealthCheckResponse 健康检查响应数据
type HealthCheckResponse struct {
	Status    string `json:"status"`    // 状态
	Timestamp int64  `json:"timestamp"` // 时间戳 (数字类型)
}
