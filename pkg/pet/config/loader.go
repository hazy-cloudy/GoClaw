package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sipeed/picoclaw/pkg/logger"
)

const (
	CharactersFile = "characters.json" // 角色配置文件名
	VoiceFile      = "voice.json"      // 语音配置文件名
)

// ConfigLoader 配置加载器
// 负责从工作区加载和管理characters.json与voice.json配置
type ConfigLoader struct {
	workspacePath string            // 工作区路径
	characters    *CharactersConfig // 角色配置
	voice         *VoiceConfig      // 语音配置
}

// NewConfigLoader 创建配置加载器实例
// workspacePath: 工作区目录路径，配置文件存放于此
func NewConfigLoader(workspacePath string) *ConfigLoader {
	return &ConfigLoader{
		workspacePath: workspacePath,
		characters:    &CharactersConfig{},
		voice:         &VoiceConfig{},
	}
}

// WorkspacePath 返回工作区路径
func (l *ConfigLoader) WorkspacePath() string {
	return l.workspacePath
}

// Load 加载所有配置文件
// 会加载characters.json和voice.json，如果文件不存在则返回错误
func (l *ConfigLoader) Load() error {
	if err := l.loadCharacters(); err != nil {
		return fmt.Errorf("failed to load characters: %w", err)
	}
	if err := l.loadVoice(); err != nil {
		return fmt.Errorf("failed to load voice: %w", err)
	}
	return nil
}

// loadCharacters 从characters.json加载角色配置
func (l *ConfigLoader) loadCharacters() error {
	path := filepath.Join(l.workspacePath, CharactersFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	var cfg CharactersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// 至少需要一个角色
	if len(cfg.Characters) == 0 {
		return fmt.Errorf("no characters found in %s", path)
	}

	// 如果没有设置激活角色，默认激活第一个
	if cfg.ActiveID == "" {
		cfg.ActiveID = cfg.Characters[0].ID
	}

	l.characters = &cfg
	logger.Infof("pet config: loaded %d characters, active_id=%s", len(cfg.Characters), cfg.ActiveID)
	return nil
}

// loadVoice 从voice.json加载语音配置
func (l *ConfigLoader) loadVoice() error {
	path := filepath.Join(l.workspacePath, VoiceFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	var cfg VoiceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	l.voice = &cfg
	logger.Infof("pet config: loaded voice config, default_model=%s, asr_enabled=%v", cfg.DefaultModel, cfg.ASREnabled)
	return nil
}

// GetCharacters 返回角色配置
func (l *ConfigLoader) GetCharacters() *CharactersConfig {
	return l.characters
}

// GetVoice 返回语音配置
func (l *ConfigLoader) GetVoice() *VoiceConfig {
	return l.voice
}

// GetActiveCharacter 返回当前激活的角色配置
func (l *ConfigLoader) GetActiveCharacter() *CharacterConfig {
	for _, c := range l.characters.Characters {
		if c.ID == l.characters.ActiveID {
			return c
		}
	}
	return l.characters.Characters[0]
}

// SaveCharacters 保存角色配置到文件
func (l *ConfigLoader) SaveCharacters() error {
	path := filepath.Join(l.workspacePath, CharactersFile)
	return l.saveJSON(path, l.characters)
}

// SaveVoice 保存语音配置到文件
func (l *ConfigLoader) SaveVoice() error {
	path := filepath.Join(l.workspacePath, VoiceFile)
	return l.saveJSON(path, l.voice)
}

// saveJSON 将数据以JSON格式保存到文件
func (l *ConfigLoader) saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

// GetVoiceModelConfig 根据名称获取语音模型配置
func (l *ConfigLoader) GetVoiceModelConfig(name string) *VoiceModelConfig {
	for _, m := range l.voice.ModelList {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// GetDefaultVoiceModel 获取默认语音模型配置
func (l *ConfigLoader) GetDefaultVoiceModel() *VoiceModelConfig {
	if l.voice.DefaultModel == "" {
		return nil
	}
	return l.GetVoiceModelConfig(l.voice.DefaultModel)
}

// EnsureDefaultConfig 确保配置目录中存在默认配置文件
// 如果文件不存在则创建默认配置
func EnsureDefaultConfig(workspacePath string) error {
	charactersPath := filepath.Join(workspacePath, CharactersFile)
	voicePath := filepath.Join(workspacePath, VoiceFile)

	// 创建默认characters.json（如果不存在）
	if _, err := os.Stat(charactersPath); os.IsNotExist(err) {
		defaultCfg := &CharactersConfig{
			Characters: []*CharacterConfig{DefaultCharacterConfig()},
			ActiveID:   "pet_001",
		}
		data, _ := json.MarshalIndent(defaultCfg, "", "  ")
		if err := os.WriteFile(charactersPath, data, 0644); err != nil {
			return fmt.Errorf("failed to create default %s: %w", charactersPath, err)
		}
		logger.Infof("pet config: created default %s", charactersPath)
	}

	// 创建默认voice.json（如果不存在）
	if _, err := os.Stat(voicePath); os.IsNotExist(err) {
		defaultCfg := DefaultVoiceConfig()
		data, _ := json.MarshalIndent(defaultCfg, "", "  ")
		if err := os.WriteFile(voicePath, data, 0644); err != nil {
			return fmt.Errorf("failed to create default %s: %w", voicePath, err)
		}
		logger.Infof("pet config: created default %s", voicePath)
	}

	return nil
}
