package pet

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
)

type PushHandler func(push any)

type PetService struct {
	msgBus      *bus.MessageBus
	config      PetServiceConfig
	pushHandler PushHandler

	characterStore *CharacterStore
	emotionEngine  *EmotionEngine

	appConfig    AppConfig
	connSessions map[string]string

	mu sync.RWMutex
}

type PetServiceConfig struct {
	WorkspacePath string
}

func NewPetService(msgBus *bus.MessageBus, cfg PetServiceConfig) *PetService {
	s := &PetService{
		msgBus:        msgBus,
		config:        cfg,
		emotionEngine: NewEmotionEngine(),
		appConfig:     AppConfig{EmotionEnabled: true, ReminderEnabled: true, ProactiveCare: true, Language: "zh-CN"},
		connSessions:  make(map[string]string),
	}

	if cfg.WorkspacePath != "" {
		s.characterStore = GetCharacterStoreWithPath(cfg.WorkspacePath)
		if err := s.characterStore.Load(); err != nil {
			fmt.Printf("pet: failed to load character: %v\n", err)
		}
	}

	return s
}

func (s *PetService) CharacterStore() *CharacterStore {
	return s.characterStore
}

func (s *PetService) EmotionEngine() *EmotionEngine {
	return s.emotionEngine
}

func (s *PetService) SetPushHandler(handler PushHandler) {
	s.pushHandler = handler
}

func (s *PetService) RegisterSession(connID, sessionID string) {
	s.mu.Lock()
	s.connSessions[connID] = sessionID
	s.mu.Unlock()
}

func (s *PetService) UnregisterSession(connID string) {
	s.mu.Lock()
	delete(s.connSessions, connID)
	s.mu.Unlock()
}

func (s *PetService) GetSessionByConnID(connID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connSessions[connID]
}

func (s *PetService) HandleRequest(connID string, req Request) error {
	sessionID := s.GetSessionByConnID(connID)

	switch req.Action {
	case ActionChat:
		return s.handleChat(sessionID, req)
	case ActionVoice:
		return s.handleVoice(sessionID, req)
	case ActionOnboardingStart:
		return s.handleOnboardingStart(sessionID)
	case ActionOnboardingConfig:
		return s.handleOnboardingConfig(sessionID, req)
	case ActionCharacterGet:
		return s.handleCharacterGet(sessionID, req)
	case ActionCharacterUpdate:
		return s.handleCharacterUpdate(sessionID, req)
	case ActionConfigGet:
		return s.handleConfigGet(sessionID, req)
	case ActionConfigUpdate:
		return s.handleConfigUpdate(sessionID, req)
	case ActionEmotionGet:
		return s.handleEmotionGet(sessionID, req)
	case ActionEventAction:
		return s.handleEventAction(sessionID, req)
	case ActionHealthCheck:
		return s.handleHealthCheck(sessionID, req)
	case ActionMemoryGet:
		return s.handleMemoryGet(sessionID, req)
	case ActionMemoryTypeAdd:
		return s.handleMemoryTypeAdd(sessionID, req)
	case ActionMemoryOldGet:
		return s.handleMemoryOldGet(sessionID, req)
	case ActionDynamicGet:
		return s.handleDynamicGet(sessionID, req)
	default:
		return s.sendError(sessionID, req.Action, fmt.Sprintf("unknown action: %s", req.Action))
	}
}

func (s *PetService) handleChat(sessionID string, req Request) error {
	var chatReq ChatRequest
	if err := json.Unmarshal(req.Data, &chatReq); err != nil {
		return s.sendError(sessionID, req.Action, "invalid chat data")
	}

	inbound := bus.InboundMessage{
		Channel:  "pet",
		ChatID:   sessionID,
		Content:  chatReq.Text,
		Metadata: map[string]string{"type": "chat", "conn_id": req.RequestID},
	}

	if err := s.msgBus.PublishInbound(context.Background(), inbound); err != nil {
		return s.sendError(sessionID, req.Action, err.Error())
	}

	return s.sendResponse(sessionID, req.Action, map[string]string{"session_key": chatReq.SessionKey})
}

func (s *PetService) handleOnboardingStart(sessionID string) error {
	// 检查是否已有角色配置
	if s.characterStore != nil && s.characterStore.IsOnboardingComplete() {
		// 已有配置，返回 error 表示不用配置了
		return s.sendError(sessionID, "onboarding_start", "onboarding already completed")
	}
	// 没有配置，返回 OK 表示需要初始化
	return s.sendResponse(sessionID, "onboarding_start", map[string]string{})
}

func (s *PetService) handleOnboardingConfig(sessionID string, req Request) error {
	var data OnboardingConfigRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid onboarding config data")
	}

	// 保存角色配置
	if s.characterStore != nil {
		mbti := DefaultMBTI()
		char := &Character{
			Name:        data.PetName,
			Persona:     data.PetPersonaContext,
			PersonaType: data.PetPersonaType,
			MBTI:        mbti,
			Avatar:      "default",
		}
		if err := s.characterStore.Update(char); err != nil {
			return s.sendError(sessionID, req.Action, err.Error())
		}
	}

	s.sendResponse(sessionID, req.Action, OnboardingConfigResponse{PetID: "pet_001"})

	s.sendPush(sessionID, PushTypeOnboardingConfig, OnboardingConfigPush{Text: "你好，我是桌宠，你可以和我聊天吗？"})

	return nil
}

func (s *PetService) handleCharacterGet(sessionID string, req Request) error {
	if s.characterStore == nil {
		return s.sendError(sessionID, req.Action, "character store not available")
	}

	char := s.characterStore.Get()
	charConfig := CharacterConfig{
		PetID:          char.PetID,
		PetName:        char.Name,
		PetPersona:     char.Persona,
		PetPersonaType: char.PersonaType,
		PetMBTI:        char.MBTI,
		Avatar:         char.Avatar,
		CreatedAt:      char.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      char.UpdatedAt.Format(time.RFC3339),
	}
	return s.sendResponse(sessionID, req.Action, charConfig)
}

func (s *PetService) handleCharacterUpdate(sessionID string, req Request) error {
	var data CharacterConfig
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid character data")
	}

	if s.characterStore != nil {
		mbti := data.PetMBTI
		if mbti.IE == 0 && mbti.SN == 0 && mbti.TF == 0 && mbti.JP == 0 {
			mbti = DefaultMBTI()
		}
		char := &Character{
			Name:        data.PetName,
			Persona:     data.PetPersona,
			PersonaType: data.PetPersonaType,
			MBTI:        mbti,
			Avatar:      data.Avatar,
		}
		if err := s.characterStore.Update(char); err != nil {
			return s.sendError(sessionID, req.Action, err.Error())
		}
		char = s.characterStore.Get()
		data.PetID = char.PetID
		data.PetMBTI = char.MBTI
		data.CreatedAt = char.CreatedAt.Format(time.RFC3339)
		data.UpdatedAt = char.UpdatedAt.Format(time.RFC3339)
	}

	return s.sendResponse(sessionID, req.Action, data)
}

func (s *PetService) handleConfigGet(sessionID string, req Request) error {
	s.mu.RLock()
	cfg := ConfigGetResponse{
		ReminderEnabled:          s.appConfig.ReminderEnabled,
		ProactiveCare:            s.appConfig.ProactiveCare,
		ProactiveIntervalMinutes: s.appConfig.ProactiveIntervalMinutes,
		VoiceEnabled:             s.appConfig.VoiceEnabled,
		Language:                 s.appConfig.Language,
	}
	s.mu.RUnlock()

	return s.sendResponse(sessionID, req.Action, cfg)
}

func (s *PetService) handleConfigUpdate(sessionID string, req Request) error {
	var data ConfigUpdateRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid config data")
	}

	s.mu.Lock()
	if data.EmotionEnabled != nil {
		s.appConfig.EmotionEnabled = *data.EmotionEnabled
	}
	if data.ReminderEnabled != nil {
		s.appConfig.ReminderEnabled = *data.ReminderEnabled
	}
	if data.ProactiveCare != nil {
		s.appConfig.ProactiveCare = *data.ProactiveCare
	}
	if data.ProactiveIntervalMinutes != nil {
		s.appConfig.ProactiveIntervalMinutes = *data.ProactiveIntervalMinutes
	}
	if data.VoiceEnabled != nil {
		s.appConfig.VoiceEnabled = *data.VoiceEnabled
	}
	if data.Language != nil {
		s.appConfig.Language = *data.Language
	}

	cfg := ConfigGetResponse{
		ReminderEnabled:          s.appConfig.ReminderEnabled,
		ProactiveCare:            s.appConfig.ProactiveCare,
		ProactiveIntervalMinutes: s.appConfig.ProactiveIntervalMinutes,
		VoiceEnabled:             s.appConfig.VoiceEnabled,
		Language:                 s.appConfig.Language,
	}
	s.mu.Unlock()

	return s.sendResponse(sessionID, req.Action, cfg)
}

func (s *PetService) handleEmotionGet(sessionID string, req Request) error {
	emotion, score := s.emotionEngine.GetEmotion()

	var petID string
	if s.characterStore != nil {
		petID = s.characterStore.Get().PetID
	}
	if petID == "" {
		petID = "pet_001"
	}

	emo := EmotionState{
		PetID:       petID,
		Emotion:     emotion,
		Score:       score,
		Expression:  "",
		Motion:      "Idle",
		Description: emotionToDescription(emotion),
	}
	return s.sendResponse(sessionID, req.Action, emo)
}

func emotionToDescription(emotion string) string {
	descriptions := map[string]string{
		"neutral":   "平静",
		"happy":     "开心",
		"sad":       "悲伤",
		"angry":     "生气",
		"worried":   "担心",
		"love":      "喜爱",
		"surprised": "惊讶",
	}
	if desc, ok := descriptions[emotion]; ok {
		return desc
	}
	return "平静"
}

func (s *PetService) handleEventAction(sessionID string, req Request) error {
	var data EventActionRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid event data")
	}

	emotion, _ := s.emotionEngine.GetEmotion()

	content := fmt.Sprintf("[事件: %s]", data.Event)
	inbound := bus.InboundMessage{
		Channel: "pet",
		ChatID:  sessionID,
		Content: content,
		Metadata: map[string]string{
			"type":       "event_action",
			"event":      data.Event,
			"params":     fmt.Sprintf("%v", data.Params),
			"emotion":    emotion,
			"expression": "",
			"motion":     "TapBody",
		},
	}

	if err := s.msgBus.PublishInbound(context.Background(), inbound); err != nil {
		return s.sendError(sessionID, req.Action, err.Error())
	}

	return s.sendResponse(sessionID, req.Action, map[string]bool{"handled": true})
}

func (s *PetService) handleHealthCheck(sessionID string, req Request) error {
	return s.sendResponse(sessionID, req.Action, map[string]string{
		"status":    "ok",
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	})
}

func (s *PetService) handleVoice(sessionID string, req Request) error {
	var voiceReq VoiceRequest
	if err := json.Unmarshal(req.Data, &voiceReq); err != nil {
		return s.sendError(sessionID, req.Action, "invalid voice data")
	}

	inbound := bus.InboundMessage{
		Channel:  "pet",
		ChatID:   sessionID,
		Content:  voiceReq.Text,
		Metadata: map[string]string{"type": "voice", "conn_id": req.RequestID},
	}

	if err := s.msgBus.PublishInbound(context.Background(), inbound); err != nil {
		return s.sendError(sessionID, req.Action, err.Error())
	}

	return s.sendResponse(sessionID, req.Action, map[string]string{"session_key": voiceReq.SessionKey})
}

func (s *PetService) handleMemoryGet(sessionID string, req Request) error {
	var memReq MemoryGetRequest
	if err := json.Unmarshal(req.Data, &memReq); err != nil {
		return s.sendError(sessionID, req.Action, "invalid memory get data")
	}

	// TODO: 实现记忆获取业务逻辑
	// 占位返回空数组，业务实现时替换
	return s.sendResponse(sessionID, req.Action, []MemoryItem{})
}

func (s *PetService) handleMemoryTypeAdd(sessionID string, req Request) error {
	var memReq MemoryTypeAddRequest
	if err := json.Unmarshal(req.Data, &memReq); err != nil {
		return s.sendError(sessionID, req.Action, "invalid memory type add data")
	}

	// TODO: 实现记忆类型添加业务逻辑
	// 占位返回空对象，业务实现时替换
	return s.sendResponse(sessionID, req.Action, map[string]string{})
}

func (s *PetService) handleMemoryOldGet(sessionID string, req Request) error {
	var memReq MemoryOldGetRequest
	if err := json.Unmarshal(req.Data, &memReq); err != nil {
		return s.sendError(sessionID, req.Action, "invalid memory old get data")
	}

	// TODO: 实现旧记忆获取业务逻辑
	// 占位返回空数组，业务实现时替换
	return s.sendResponse(sessionID, req.Action, []MemorySummary{})
}

func (s *PetService) handleDynamicGet(sessionID string, req Request) error {
	var dynReq DynamicGetRequest
	if err := json.Unmarshal(req.Data, &dynReq); err != nil {
		return s.sendError(sessionID, req.Action, "invalid dynamic get data")
	}

	// TODO: 实现动态获取业务逻辑
	// 占位返回空数组，业务实现时替换
	return s.sendResponse(sessionID, req.Action, []DynamicItem{})
}

func (s *PetService) sendResponse(sessionID, action string, data interface{}) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp := Response{
		Status: StatusOK,
		Action: action,
		Data:   rawData,
	}

	if s.pushHandler == nil {
		return nil
	}
	s.pushHandler(resp)
	return nil
}

func (s *PetService) sendError(sessionID, action, errMsg string) error {
	data := map[string]string{"error": errMsg}
	resp := Response{
		Status: StatusError,
		Action: action,
		Data:   mustMarshal(data),
	}

	if s.pushHandler == nil {
		return fmt.Errorf("%s", errMsg)
	}
	s.pushHandler(resp)
	return fmt.Errorf("%s", errMsg)
}

func (s *PetService) sendPush(sessionID, pushType string, data interface{}) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	push := Push{
		Type:      "push",
		PushType:  pushType,
		Data:      rawData,
		Timestamp: time.Now().Unix(),
	}

	if s.pushHandler == nil {
		return nil
	}
	s.pushHandler(push)
	return nil
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{"error": "marshal error"}`)
	}
	return data
}
