package characters

import (
	"fmt"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/pet/config"
	"github.com/sipeed/picoclaw/pkg/pet/emotion"
)

// Manager 角色管理器
// 负责管理所有角色，支持多角色切换
type Manager struct {
	characters map[string]*Character // 所有角色映射
	activeID   string                // 当前激活的角色ID
}

// NewManager 从配置创建角色管理器
// 会根据配置初始化所有角色及其情绪、MBTI状态
func NewManager(cfg *config.CharactersConfig) (*Manager, error) {
	if cfg == nil || len(cfg.Characters) == 0 {
		return nil, fmt.Errorf("no characters provided")
	}

	m := &Manager{
		characters: make(map[string]*Character),
	}

	// 初始化所有角色
	for _, cc := range cfg.Characters {
		char := NewCharacter(cc.ID, cc.Name, cc.Persona, cc.PersonaType, cc.Avatar)

		// 设置情绪状态
		if cc.EmotionState != nil {
			char.SetEmotions(emotion.SixEmotions{
				Joy:      cc.EmotionState.Joy,
				Anger:    cc.EmotionState.Anger,
				Sadness:  cc.EmotionState.Sadness,
				Disgust:  cc.EmotionState.Disgust,
				Surprise: cc.EmotionState.Surprise,
				Fear:     cc.EmotionState.Fear,
			})
		}

		// 设置MBTI性格
		if cc.MBTI != nil {
			char.SetPersonality(MBTIPersonality{
				IE: cc.MBTI.IE,
				SN: cc.MBTI.SN,
				TF: cc.MBTI.TF,
				JP: cc.MBTI.JP,
			})
		}

		m.characters[char.ID] = char
		logger.Infof("characters: loaded character id=%s, name=%s", char.ID, char.Name)
	}

	// 验证并设置激活角色
	if err := m.validateActiveID(cfg.ActiveID); err != nil {
		return nil, err
	}
	m.activeID = cfg.ActiveID

	logger.Infof("characters: manager created with %d characters, active=%s", len(m.characters), m.activeID)
	return m, nil
}

// validateActiveID 验证激活角色ID是否有效
func (m *Manager) validateActiveID(id string) error {
	if id == "" {
		return fmt.Errorf("active character ID is empty")
	}
	if _, ok := m.characters[id]; !ok {
		return fmt.Errorf("active character ID %s not found", id)
	}
	return nil
}

// GetCurrent 获取当前激活的角色
func (m *Manager) GetCurrent() *Character {
	if c, ok := m.characters[m.activeID]; ok {
		return c
	}
	return nil
}

// GetCurrentID 获取当前激活角色的ID
func (m *Manager) GetCurrentID() string {
	return m.activeID
}

// Switch 切换到指定角色
// 切换是即时完成的，新角色拥有自己独立的状态
func (m *Manager) Switch(id string) error {
	if id == m.activeID {
		return nil // 已经是目标角色，无需切换
	}

	if _, ok := m.characters[id]; !ok {
		return fmt.Errorf("character %s not found", id)
	}

	logger.Infof("characters: switching from %s to %s", m.activeID, id)
	m.activeID = id
	return nil
}

// List 返回所有角色列表
func (m *Manager) List() []*Character {
	result := make([]*Character, 0, len(m.characters))
	for _, c := range m.characters {
		result = append(result, c)
	}
	return result
}

// Get 根据ID获取指定角色
func (m *Manager) Get(id string) *Character {
	return m.characters[id]
}

// Add 添加新角色到管理器
func (m *Manager) Add(c *Character) {
	m.characters[c.ID] = c
}

// Remove 从管理器移除角色
// 不能移除当前激活的角色
func (m *Manager) Remove(id string) error {
	if id == m.activeID {
		return fmt.Errorf("cannot remove active character")
	}
	if _, ok := m.characters[id]; !ok {
		return fmt.Errorf("character %s not found", id)
	}
	delete(m.characters, id)
	return nil
}

// ToConfig 将当前状态转换为配置结构
// 用于持久化保存
func (m *Manager) ToConfig() *config.CharactersConfig {
	characters := make([]*config.CharacterConfig, 0, len(m.characters))
	for _, c := range m.characters {
		emo := c.GetEmotions()
		characters = append(characters, &config.CharacterConfig{
			ID:          c.ID,
			Name:        c.Name,
			Persona:     c.Persona,
			PersonaType: c.PersonaType,
			Avatar:      c.Avatar,
			EmotionState: &config.EmotionState{
				Joy:      emo.Joy,
				Anger:    emo.Anger,
				Sadness:  emo.Sadness,
				Disgust:  emo.Disgust,
				Surprise: emo.Surprise,
				Fear:     emo.Fear,
			},
			MBTI: &config.MBTIConfig{
				IE: c.MBTI.IE,
				SN: c.MBTI.SN,
				TF: c.MBTI.TF,
				JP: c.MBTI.JP,
			},
		})
	}
	return &config.CharactersConfig{
		Characters: characters,
		ActiveID:   m.activeID,
	}
}
