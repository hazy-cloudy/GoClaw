package pet

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sipeed/picoclaw/pkg/logger"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/pet/action"
	"github.com/sipeed/picoclaw/pkg/pet/characters"
	"github.com/sipeed/picoclaw/pkg/pet/config"
	"github.com/sipeed/picoclaw/pkg/pet/memory"
	"github.com/sipeed/picoclaw/pkg/pet/voice"
)

type PushHandler func(push any)

type PetService struct {
	msgBus      *bus.MessageBus
	config      PetServiceConfig
	pushHandler PushHandler

	configManager *config.Manager
	charManager   *characters.Manager
	actionManager *action.ActionManager
	memoryStore   *memory.Store
	voiceLoader   *voice.Loader

	connSessions map[string]string

	mu sync.RWMutex

	ctx         context.Context
	cancel      context.CancelFunc
	decayTicker *time.Ticker
}

type PetServiceConfig struct {
	WorkspacePath string
}

func NewPetService(msgBus *bus.MessageBus, cfg PetServiceConfig) (*PetService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &PetService{
		msgBus:        msgBus,
		config:        cfg,
		actionManager: action.NewActionManager(cfg.WorkspacePath),
		connSessions:  make(map[string]string),
		ctx:           ctx,
		cancel:        cancel,
	}
	workspacePath := cfg.WorkspacePath
	if workspacePath != "" {
		logger.Debugf("pet: workspacePath=%s", workspacePath)
		s.configManager = config.NewManager(workspacePath)
		if s.configManager == nil {
			return nil, fmt.Errorf("failed to create config manager")
		}

		var err error
		s.charManager, err = characters.NewManager(s.configManager.GetCharacters(), s.configManager)
		if err != nil {
			fmt.Printf("pet: failed to create character manager: %v\n", err)
			return nil, err
		}

		s.voiceLoader = voice.NewLoader(s.configManager.GetVoice())
		if err := s.voiceLoader.Load(); err != nil {
			fmt.Printf("pet: failed to load voice: %v\n", err)
		}
		if s.voiceLoader.GetProvider() != nil {
			fmt.Println("pet: voice provider loaded successfully")
		} else {
			fmt.Println("pet: voice provider is nil after loading")
		}

		s.memoryStore, err = memory.NewStore(cfg.WorkspacePath)
		if err != nil {
			fmt.Printf("pet: failed to create memory store: %v\n", err)
		}

		if err := s.actionManager.Load(); err != nil {
			fmt.Printf("pet: failed to load actions: %v\n", err)
		}
	}

	return s, nil
}

func (s *PetService) Start() {
	s.decayTicker = time.NewTicker(5 * time.Second)
	go s.runEmotionDecay()
	fmt.Println("pet: PetService started, emotion decay ticker running")
}

func (s *PetService) runEmotionDecay() {
	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("pet: emotion decay ticker stopped")
			return
		case <-s.decayTicker.C:
			if char := s.charManager.GetCurrent(); char != nil {
				char.GetEmotionEngine().ApplyDecay(5 * time.Second)
				if shouldPush, push := char.GetEmotionEngine().ShouldPush(); shouldPush {
					s.Push(push)
				}
			}
		}
	}
}

func (s *PetService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.decayTicker != nil {
		s.decayTicker.Stop()
	}
	if s.memoryStore != nil {
		s.memoryStore.Close()
	}
	s.Shutdown()
	fmt.Println("pet: PetService stopped")
}

func (s *PetService) Shutdown() {
	if s.charManager != nil {
		if err := s.charManager.SavePrivateConfig(); err != nil {
			fmt.Printf("pet: failed to save private config: %v\n", err)
		}
	}
	if s.configManager != nil {
		if err := s.configManager.Save(); err != nil {
			fmt.Printf("pet: failed to save config: %v\n", err)
		}
	}
}

func (s *PetService) ConfigManager() *config.Manager {
	return s.configManager
}

func (s *PetService) CharManager() *characters.Manager {
	return s.charManager
}

func (s *PetService) ActionManager() *action.ActionManager {
	return s.actionManager
}

func (s *PetService) MemoryStore() *memory.Store {
	return s.memoryStore
}

func (s *PetService) VoiceLoader() *voice.Loader {
	return s.voiceLoader
}

func (s *PetService) SetPushHandler(handler PushHandler) {
	s.pushHandler = handler
}

func (s *PetService) Push(push any) {
	if s.pushHandler != nil {
		s.pushHandler(push)
	}
}

func (s *PetService) PushToolStart(tool string, data json.RawMessage) {
	if s.pushHandler == nil {
		return
	}

	streamData := map[string]interface{}{
		"type":  "tool",
		"text":  "正在调用 " + tool,
		"tool":  tool,
		"phase": "start",
	}

	push := map[string]interface{}{
		"type":      "push",
		"push_type": "ai_chat",
		"data":      streamData,
		"timestamp": time.Now().Unix(),
		"is_final":  true,
	}

	s.pushHandler(push)
}

func (s *PetService) PushToolEnd(tool string, data json.RawMessage) {
	if s.pushHandler == nil {
		return
	}

	streamData := map[string]interface{}{
		"type":  "tool",
		"text":  tool + " 执行完成",
		"tool":  tool,
		"phase": "end",
	}

	push := map[string]interface{}{
		"type":      "push",
		"push_type": "ai_chat",
		"data":      streamData,
		"timestamp": time.Now().Unix(),
		"is_final":  true,
	}

	s.pushHandler(push)
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

func (s *PetService) PushInitStatus(sessionID string) {
	char := s.charManager.GetCurrent()

	var character *CharacterConfig
	var mbti MBTIConfig
	var emotionState EmotionState

	if char != nil {
		emotions := char.GetEmotions()
		mbtiCfg := char.GetMBTI()
		emoEngine := char.GetEmotionEngine()
		dominantEmotion, emotionScore := emoEngine.GetDominantEmotion()

		character = &CharacterConfig{
			PetID:          char.ID,
			PetName:        char.Name,
			PetPersona:     char.Persona,
			PetPersonaType: char.PersonaType,
			Avatar:         char.Avatar,
		}
		mbti = MBTIConfig{
			IE: mbtiCfg.IE,
			SN: mbtiCfg.SN,
			TF: mbtiCfg.TF,
			JP: mbtiCfg.JP,
		}
		emotionState = EmotionState{
			PetID:       char.ID,
			Emotion:     dominantEmotion,
			Joy:         emotions.Joy,
			Anger:       emotions.Anger,
			Sadness:     emotions.Sadness,
			Disgust:     emotions.Disgust,
			Surprise:    emotions.Surprise,
			Fear:        emotions.Fear,
			Description: GetEmotionDescription(dominantEmotion, emotionScore),
		}
	} else {
		mbti = DefaultMBTI()
		emotionState = EmotionState{
			PetID:       "pet_001",
			Emotion:     "neutral",
			Joy:         50,
			Anger:       50,
			Sadness:     50,
			Disgust:     50,
			Surprise:    50,
			Fear:        50,
			Description: "平静",
		}
	}

	s.sendPush(sessionID, PushTypeInitStatus, InitStatusPush{
		NeedConfig:   char == nil,
		HasCharacter: char != nil,
		Character:    character,
		MBTI:         mbti,
		EmotionState: emotionState,
	})
}

func (s *PetService) HandleRequest(connID string, req Request) error {
	sessionID := s.GetSessionByConnID(connID)

	switch req.Action {
	case ActionChat:
		return s.handleChat(sessionID, req)
	case ActionOnboardingConfig:
		return s.handleOnboardingConfig(sessionID, req)
	case ActionCharacterGet:
		return s.handleCharacterGet(sessionID, req)
	case ActionCharacterUpdate:
		return s.handleCharacterUpdate(sessionID, req)
	case ActionCharacterSwitch:
		return s.handleCharacterSwitch(sessionID, req)
	case ActionConfigGet:
		return s.handleConfigGet(sessionID, req)
	case ActionConfigUpdate:
		return s.handleConfigUpdate(sessionID, req)
	case ActionEmotionGet:
		return s.handleEmotionGet(sessionID, req)
	case ActionHealthCheck:
		return s.handleHealthCheck(sessionID, req)
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

func (s *PetService) handleOnboardingConfig(sessionID string, req Request) error {
	var data OnboardingConfigRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid onboarding config data")
	}

	char := s.charManager.GetCurrent()
	if char != nil {
		char.Name = data.PetName
		char.Persona = data.PetPersona
		char.PersonaType = data.PetPersonaType
		s.charManager.UpdateCharacter(char.ID, data.PetName, data.PetPersona, data.PetPersonaType)
		// 保存会在 shutdown 时统一进行
	}

	return s.sendResponse(sessionID, req.Action, OnboardingConfigResponse{PetID: char.ID})
}

func (s *PetService) handleCharacterGet(sessionID string, req Request) error {
	char := s.charManager.GetCurrent()
	if char == nil {
		return s.sendError(sessionID, req.Action, "no active character")
	}

	charConfig := CharacterConfig{
		PetID:          char.ID,
		PetName:        char.Name,
		PetPersona:     char.Persona,
		PetPersonaType: char.PersonaType,
		Avatar:         char.Avatar,
	}
	return s.sendResponse(sessionID, req.Action, charConfig)
}

func (s *PetService) handleCharacterUpdate(sessionID string, req Request) error {
	var data CharacterUpdateRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid character data")
	}

	char := s.charManager.GetCurrent()
	if char != nil {
		if data.PetName != "" {
			char.Name = data.PetName
		}
		if data.PetPersona != "" {
			char.Persona = data.PetPersona
		}
		if data.PetPersonaType != "" {
			char.PersonaType = data.PetPersonaType
		}
		s.charManager.UpdateCharacter(char.ID, data.PetName, data.PetPersona, data.PetPersonaType)
		data.PetID = char.ID
	}

	return s.sendResponse(sessionID, req.Action, data)
}

func (s *PetService) handleCharacterSwitch(sessionID string, req Request) error {
	var data CharacterSwitchRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid character switch data")
	}

	if err := s.charManager.Switch(data.CharacterID); err != nil {
		return s.sendError(sessionID, req.Action, err.Error())
	}

	s.sendPush(sessionID, PushTypeCharacterSwitch, CharacterSwitchPush{
		CharacterID: s.charManager.GetCurrentID(),
	})

	return s.sendResponse(sessionID, req.Action, map[string]string{"character_id": data.CharacterID})
}

func (s *PetService) handleConfigGet(sessionID string, req Request) error {
	cfg := s.configManager.GetApp()
	return s.sendResponse(sessionID, req.Action, cfg)
}

func (s *PetService) handleConfigUpdate(sessionID string, req Request) error {
	var data ConfigUpdateRequest
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return s.sendError(sessionID, req.Action, "invalid config data")
	}

	cfg := s.configManager.GetApp()

	if data.EmotionEnabled != nil {
		cfg.EmotionEnabled = *data.EmotionEnabled
	}
	if data.ReminderEnabled != nil {
		cfg.ReminderEnabled = *data.ReminderEnabled
	}
	if data.ProactiveCare != nil {
		cfg.ProactiveCare = *data.ProactiveCare
	}
	if data.ProactiveIntervalMinutes != nil {
		cfg.ProactiveIntervalMinutes = *data.ProactiveIntervalMinutes
	}
	if data.VoiceEnabled != nil {
		cfg.VoiceEnabled = *data.VoiceEnabled
	}
	if data.Language != nil {
		cfg.Language = *data.Language
	}

	s.configManager.SetAppConfig(cfg)

	return s.sendResponse(sessionID, req.Action, cfg)
}

func (s *PetService) handleEmotionGet(sessionID string, req Request) error {
	char := s.charManager.GetCurrent()
	if char == nil {
		return s.sendError(sessionID, req.Action, "no active character")
	}

	emotions := char.GetEmotions()
	emoEngine := char.GetEmotionEngine()
	dominantEmotion, emotionScore := emoEngine.GetDominantEmotion()

	emo := EmotionState{
		PetID:       char.ID,
		Emotion:     dominantEmotion,
		Joy:         emotions.Joy,
		Anger:       emotions.Anger,
		Sadness:     emotions.Sadness,
		Disgust:     emotions.Disgust,
		Surprise:    emotions.Surprise,
		Fear:        emotions.Fear,
		Description: GetEmotionDescription(dominantEmotion, emotionScore),
	}
	return s.sendResponse(sessionID, req.Action, emo)
}

func (s *PetService) handleHealthCheck(sessionID string, req Request) error {
	return s.sendResponse(sessionID, req.Action, HealthCheckResponse{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
	})
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

func (s *PetService) AppConfig() *config.AppConfig {
	return s.configManager.GetApp()
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{"error": "marshal error"}`)
	}
	return data
}

var EmotionDescriptions = map[string][]string{
	"joy":      {"很不开心", "有点不开心", "有点开心", "很开心"},
	"sadness":  {"很平静", "有点难过", "比较难过", "非常难过"},
	"anger":    {"很平静", "有点生气", "比较生气", "非常生气"},
	"fear":     {"很平静", "有点害怕", "比较害怕", "非常害怕"},
	"disgust":  {"很平静", "有点厌恶", "比较厌恶", "非常厌恶"},
	"surprise": {"很平静", "有点惊讶", "比较惊讶", "非常惊讶"},
}

func scoreToIndex(score int) int {
	switch {
	case score < 20:
		return 0
	case score < 50:
		return 1
	case score < 80:
		return 2
	default:
		return 3
	}
}

func GetEmotionDescription(emotion string, score int) string {
	if emotion == "neutral" || score == 50 {
		return "平静"
	}

	descriptions, ok := EmotionDescriptions[emotion]
	if !ok {
		return "平静"
	}

	idx := scoreToIndex(score)
	return descriptions[idx]
}
