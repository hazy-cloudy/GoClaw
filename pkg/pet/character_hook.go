package pet

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/providers"
)

// CharacterHook 角色人格拦截器
// 功能：
// 1. 在每次 LLM 调用前，将角色人格信息注入到 system prompt
// 2. 对所有 channel 生效（不仅是 pet channel）
// 3. 通过读取全局 CharacterStore 获取当前人格配置
type CharacterHook struct {
	characterStore *CharacterStore // 角色配置存储（全局单例）
}

// NewCharacterHook 创建 CharacterHook 实例
func NewCharacterHook(store *CharacterStore) *CharacterHook {
	return &CharacterHook{
		characterStore: store,
	}
}

// BeforeLLM LLM 调用前拦截
// 将角色人格信息作为 system message 注入到消息列表最前面
func (h *CharacterHook) BeforeLLM(ctx context.Context, req *agent.LLMHookRequest) (*agent.LLMHookRequest, agent.HookDecision, error) {
	if req == nil || req.Messages == nil {
		return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	// 获取角色人格 prompt
	personaPrompt := h.characterStore.GetPersonaPrompt()
	if personaPrompt == "" {
		return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	// 构建 system message
	personaMsg := providers.Message{
		Role:    "system",
		Content: personaPrompt,
	}

	// 将人格信息插入到消息列表最前面
	// 这样人格信息会在所有其他 system prompt 之前
	req.Messages = append([]providers.Message{personaMsg}, req.Messages...)

	return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// AfterLLM LLM 调用后拦截（暂未使用，保留扩展性）
func (h *CharacterHook) AfterLLM(ctx context.Context, resp *agent.LLMHookResponse) (*agent.LLMHookResponse, agent.HookDecision, error) {
	return resp, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// EmotionHook 情绪状态拦截器
// 功能：
// 1. 在 LLM 调用前，根据当前情绪状态注入情绪上下文
// 2. 在 LLM 调用后，解析 LLM 输出中的情绪标签并更新状态
type EmotionHook struct {
	engine *EmotionEngine // 情绪引擎
}

// NewEmotionHook 创建 EmotionHook 实例
func NewEmotionHook(engine *EmotionEngine) *EmotionHook {
	return &EmotionHook{
		engine: engine,
	}
}

// BeforeLLM LLM 调用前拦截
// 将情绪状态作为提示注入，帮助 LLM 生成符合当前情绪的回复
func (h *EmotionHook) BeforeLLM(ctx context.Context, req *agent.LLMHookRequest) (*agent.LLMHookRequest, agent.HookDecision, error) {
	if req == nil || req.Messages == nil {
		return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	// 获取情绪上下文提示
	emotionContext := h.engine.GetContextPrompt()
	if emotionContext == "" {
		return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	// 构建情绪 system message
	emotionMsg := providers.Message{
		Role:    "system",
		Content: emotionContext,
	}

	// 插入到第一个 system message 之后
	// 保持人格信息在最前，情绪信息在其后
	req.Messages = append(req.Messages[:1], append([]providers.Message{emotionMsg}, req.Messages[1:]...)...)

	return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// AfterLLM LLM 调用后拦截
// 从 LLM 响应中解析情绪标签并更新情绪引擎状态
func (h *EmotionHook) AfterLLM(ctx context.Context, resp *agent.LLMHookResponse) (*agent.LLMHookResponse, agent.HookDecision, error) {
	if resp == nil || resp.Response == nil {
		return resp, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	// 从 LLM 响应中解析情绪并更新引擎
	h.engine.UpdateFromLLMResponse(resp.Response)

	return resp, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// EmotionEngine 情绪状态机
// 功能：
// 1. 维护当前情绪状态（emotion + score）
// 2. 提供情绪上下文提示生成
// 3. 从 LLM 响应中解析情绪标签
type EmotionEngine struct {
	emotion string  // 当前情绪标签
	score   float64 // 情绪强度 (0.0-1.0)
	mu      any     // 简单的互斥锁（避免使用 sync.Mutex 引用问题）
}

// NewEmotionEngine 创建情绪引擎实例
func NewEmotionEngine() *EmotionEngine {
	return &EmotionEngine{
		emotion: "neutral",
		score:   0.5,
	}
}

// GetEmotion 获取当前情绪状态
func (e *EmotionEngine) GetEmotion() (string, float64) {
	return e.emotion, e.score
}

// SetEmotion 设置当前情绪状态
func (e *EmotionEngine) SetEmotion(emotion string, score float64) {
	e.emotion = emotion
	e.score = score
}

// GetContextPrompt 获取用于注入 LLM 的情绪上下文提示
// 根据当前情绪状态返回格式化文本
func (e *EmotionEngine) GetContextPrompt() string {
	emotion, score := e.GetEmotion()

	// 中性或低强度时不注入提示
	if emotion == "neutral" || score < 0.3 {
		return ""
	}

	// 情绪标签转中文描述
	desc := emotion
	switch emotion {
	case "happy":
		desc = "开心"
	case "sad":
		desc = "难过"
	case "angry":
		desc = "生气"
	case "worried":
		desc = "担心"
	case "love":
		desc = "喜爱"
	case "surprised":
		desc = "惊讶"
	}

	// 根据强度生成不同级别的提示
	if score > 0.8 {
		return "【当前状态】：非常" + desc
	} else if score > 0.6 {
		return "【当前状态】：有点" + desc
	}

	return ""
}

// UpdateFromLLMResponse 从 LLM 响应中解析情绪标签并更新状态
func (e *EmotionEngine) UpdateFromLLMResponse(resp *providers.LLMResponse) {
	if resp == nil || resp.Content == "" {
		return
	}

	// 解析情绪标签
	emotion := parseEmotionFromContent(resp.Content)
	if emotion != "" {
		// 设置情绪和默认强度
		e.SetEmotion(emotion, 0.6)
	}
}

// parseEmotionFromContent 从文本内容中解析情绪标签
// 支持的情绪标签：happy, sad, angry, worried, love, surprised, neutral
func parseEmotionFromContent(content string) string {
	emotions := []string{"happy", "sad", "angry", "worried", "love", "surprised", "neutral"}
	for _, emo := range emotions {
		if containsSubstring(content, emo) {
			return emo
		}
	}
	return ""
}

// containsSubstring 检查字符串是否包含子串
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
