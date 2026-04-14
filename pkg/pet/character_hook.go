package pet

import (
	"context"
	"fmt"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/pet/characters"
	"github.com/sipeed/picoclaw/pkg/providers"
)

type CharacterHook struct {
	charManager *characters.Manager
}

func NewCharacterHook(charManager *characters.Manager) *CharacterHook {
	return &CharacterHook{
		charManager: charManager,
	}
}

func (h *CharacterHook) BeforeLLM(ctx context.Context, req *agent.LLMHookRequest) (*agent.LLMHookRequest, agent.HookDecision, error) {
	if req == nil || req.Messages == nil {
		return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	char := h.charManager.GetCurrent()
	if char == nil {
		return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
	}

	personaPrompt := h.buildPersonaPrompt(char)

	personaMsg := providers.Message{
		Role:    "system",
		Content: personaPrompt,
	}

	req.Messages = append([]providers.Message{personaMsg}, req.Messages...)

	logger.InfoCF("pet", "CharacterHook: injected persona prompt", map[string]any{
		"character":      char.Name,
		"persona_length": len(personaPrompt),
	})

	return req, agent.HookDecision{Action: agent.HookActionContinue}, nil
}

func (h *CharacterHook) buildPersonaPrompt(char *characters.Character) string {
	return fmt.Sprintf(`【桌宠角色信息】
姓名：%s
性格类型：%s

性格描述：
%s

回复风格应体现以上人格特征，现在你作为一位桌宠角色，正在和用户聊天，聊天需要语言简单，就几个字好，在其他事情上就可以正常回答十几个字。`, char.Name, char.PersonaType, char.Persona)
}

func (h *CharacterHook) AfterLLM(ctx context.Context, resp *agent.LLMHookResponse) (*agent.LLMHookResponse, agent.HookDecision, error) {
	return resp, agent.HookDecision{Action: agent.HookActionContinue}, nil
}
