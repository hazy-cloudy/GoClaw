package config

// CharactersConfig 多角色配置结构
// 包含所有角色列表和当前激活的角色ID
type CharactersConfig struct {
	Characters []*CharacterConfig `json:"characters"` // 角色列表
	ActiveID   string             `json:"active_id"`  // 当前激活的角色ID
}

// CharacterConfig 单个角色的配置
// 包含角色的基本信息、情绪状态和MBTI性格配置
type CharacterConfig struct {
	ID           string        `json:"id"`                      // 角色唯一标识
	Name         string        `json:"name"`                    // 角色名称
	Persona      string        `json:"persona"`                 // 性格描述
	PersonaType  string        `json:"persona_type"`            // 性格类型（如gentle/playful等）
	Avatar       string        `json:"avatar"`                  // 头像/模型ID
	EmotionState *EmotionState `json:"emotion_state,omitempty"` // 情绪状态（可选）
	MBTI         *MBTIConfig   `json:"mbti,omitempty"`          // MBTI配置（可选）
}

// EmotionState 情绪状态配置
// 六大基础情绪的强度值，50为中性
type EmotionState struct {
	Joy      int `json:"joy"`      // 快乐值 0-100
	Anger    int `json:"anger"`    // 愤怒值 0-100
	Sadness  int `json:"sadness"`  // 悲伤值 0-100
	Disgust  int `json:"disgust"`  // 厌恶值 0-100
	Surprise int `json:"surprise"` // 惊讶值 0-100
	Fear     int `json:"fear"`     // 恐惧值 0-100
}

// ToSixEmotions 转换为emotion包的SixEmotions结构
func (e *EmotionState) ToSixEmotions() SixEmotions {
	return SixEmotions{
		Joy:      e.Joy,
		Anger:    e.Anger,
		Sadness:  e.Sadness,
		Disgust:  e.Disgust,
		Surprise: e.Surprise,
		Fear:     e.Fear,
	}
}

// SixEmotions 六维情绪结构（用于内部计算）
type SixEmotions struct {
	Joy      int // 快乐
	Anger    int // 愤怒
	Sadness  int // 悲伤
	Disgust  int // 厌恶
	Surprise int // 惊讶
	Fear     int // 恐惧
}

// MBTIConfig MBTI性格四维配置
// 每个维度0-100，50为中心点
type MBTIConfig struct {
	IE int `json:"ie"` // 内向(I)-外向(E)：<50偏内向，>50偏外向
	SN int `json:"sn"` // 实感(S)-直觉(N)：<50偏实感，>50偏直觉
	TF int `json:"tf"` // 理性(T)-感性(F)：<50偏理性，>50偏感性
	JP int `json:"jp"` // 判断(J)-感知(P)：<50偏判断，>50偏感知
}

// VoiceConfig 语音配置结构
// 包含TTS模型列表和ASR设置
type VoiceConfig struct {
	ModelList    []*VoiceModelConfig `json:"model_list"`    // 可用的语音模型列表
	DefaultModel string              `json:"default_model"` // 默认使用的模型名称
	ASREnabled   bool                `json:"asr_enabled"`   // 是否启用语音识别
}

// VoiceModelConfig 语音模型配置
// 支持多种TTS服务商（如Minimax、OpenAI等）
type VoiceModelConfig struct {
	Name    string `json:"name"`     // 模型标识名称
	Model   string `json:"model"`    // 实际使用的模型ID（如speech-2.8-hd）
	APIKey  string `json:"api_key"`  // API密钥（支持${ENV_VAR}格式）
	APIBase string `json:"api_base"` // API地址
	Enabled bool   `json:"enabled"`  // 是否启用
}

// DefaultEmotionState 返回默认的中性情绪状态
// 所有情绪值为50（平静状态）
func DefaultEmotionState() *EmotionState {
	return &EmotionState{
		Joy:      50,
		Anger:    50,
		Sadness:  50,
		Disgust:  50,
		Surprise: 50,
		Fear:     50,
	}
}

// DefaultMBTI 返回默认的MBTI配置
// 所有维度为50（完全中性）
func DefaultMBTI() *MBTIConfig {
	return &MBTIConfig{
		IE: 50,
		SN: 50,
		TF: 50,
		JP: 50,
	}
}

// DefaultCharacterConfig 返回默认的角色配置
// 用于新角色或配置缺失时
func DefaultCharacterConfig() *CharacterConfig {
	return &CharacterConfig{
		ID:           "pet_001",
		Name:         "艾莉",
		Persona:      "温柔体贴，善于关心他人",
		PersonaType:  "gentle",
		Avatar:       "default",
		EmotionState: DefaultEmotionState(),
		MBTI:         DefaultMBTI(),
	}
}

// DefaultVoiceConfig 返回默认的语音配置
func DefaultVoiceConfig() *VoiceConfig {
	return &VoiceConfig{
		ModelList:    []*VoiceModelConfig{},
		DefaultModel: "",
		ASREnabled:   false,
	}
}
