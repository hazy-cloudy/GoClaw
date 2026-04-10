package pet

import (
	"context"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/logger"
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

	// 记录人格注入日志
	logger.InfoCF("pet", "CharacterHook: 已注入人格Prompt", map[string]any{
		"persona_length": len(personaPrompt),
		"persona_preview": func() string {
			if len(personaPrompt) > 200 {
				return personaPrompt[:200] + "..."
			}
			return personaPrompt
		}(),
		"total_messages": len(req.Messages),
	})

	return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

// AfterLLM LLM 调用后拦截（暂未使用，保留扩展性）
func (h *CharacterHook) AfterLLM(ctx context.Context, resp *agent.LLMHookResponse) (*agent.LLMHookResponse, agent.HookDecision, error) {
	return resp, agent.HookDecision{Action: agent.HookActionContinue}, nil
}
