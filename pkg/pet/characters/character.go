package characters

import (
	"github.com/sipeed/picoclaw/pkg/pet/emotion"
)

// Character 角色实体
// 代表一个独立的桌宠角色，拥有自己的情绪、性格和记忆
type Character struct {
	ID          string // 角色唯一标识
	Name        string // 角色名称
	Persona     string // 性格描述
	PersonaType string // 性格类型（如gentle/playful）
	Avatar      string // 头像/模型ID

	EmotionEngine *emotion.EmotionEngine // 角色专属的情绪引擎
	MBTI          MBTIPersonality        // MBTI性格配置
}

// MBTIPersonality MBTI性格结构
// 用于长期追踪角色的性格漂移
type MBTIPersonality struct {
	IE int // 内向(I)-外向(E)：<50偏内向，>50偏外向
	SN int // 实感(S)-直觉(N)：<50偏实感，>50偏直觉
	TF int // 理性(T)-感性(F)：<50偏理性，>50偏感性
	JP int // 判断(J)-感知(P)：<50偏判断，>50偏感知
}

// NewCharacter 创建新角色实例
// 所有新建角色都有默认的中性情绪和MBTI配置
func NewCharacter(id, name, persona, personaType, avatar string) *Character {
	return &Character{
		ID:            id,
		Name:          name,
		Persona:       persona,
		PersonaType:   personaType,
		Avatar:        avatar,
		EmotionEngine: emotion.NewEmotionEngine(""),
		MBTI: MBTIPersonality{
			IE: 50,
			SN: 50,
			TF: 50,
			JP: 50,
		},
	}
}

// GetEmotionEngine 获取角色的情绪引擎
func (c *Character) GetEmotionEngine() *emotion.EmotionEngine {
	return c.EmotionEngine
}

// GetMBTI 获取角色的MBTI配置
func (c *Character) GetMBTI() MBTIPersonality {
	return c.MBTI
}

// SetEmotions 设置角色的六维情绪值
func (c *Character) SetEmotions(e emotion.SixEmotions) {
	c.EmotionEngine.SetEmotion(emotion.EmotionJoy, e.Joy)
	c.EmotionEngine.SetEmotion(emotion.EmotionAnger, e.Anger)
	c.EmotionEngine.SetEmotion(emotion.EmotionSadness, e.Sadness)
	c.EmotionEngine.SetEmotion(emotion.EmotionDisgust, e.Disgust)
	c.EmotionEngine.SetEmotion(emotion.EmotionSurprise, e.Surprise)
	c.EmotionEngine.SetEmotion(emotion.EmotionFear, e.Fear)
}

// GetEmotions 获取角色的当前六维情绪值
func (c *Character) GetEmotions() emotion.SixEmotions {
	return c.EmotionEngine.GetEmotions()
}

// SetPersonality 设置角色的MBTI性格配置
func (c *Character) SetPersonality(m MBTIPersonality) {
	c.MBTI = m
}

// Clone 创建角色的深拷贝
// 情绪状态也会被复制一份
func (c *Character) Clone() *Character {
	clone := &Character{
		ID:            c.ID,
		Name:          c.Name,
		Persona:       c.Persona,
		PersonaType:   c.PersonaType,
		Avatar:        c.Avatar,
		EmotionEngine: emotion.NewEmotionEngine(""),
		MBTI:          c.MBTI,
	}
	clone.SetEmotions(c.GetEmotions())
	return clone
}
