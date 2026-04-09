package pet

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CharacterStore 角色配置存储
// 使用 Markdown 格式存储，支持以下结构：
// ---
// pet_name: 艾莉
// pet_persona: 温柔体贴
// pet_mbti:
//
//	ie: 50
//	sn: 50
//	tf: 50
//	jp: 50
//
// ---
// # Soul
// ## Personality
// [性格描述]
// ## Values
// [价值观描述]
type CharacterStore struct {
	mu            sync.RWMutex
	workspacePath string     // 工作区路径
	character     *Character // 当前角色配置
	soulContent   string     // Soul.md 内容（性格描述）
}

// Character 角色完整配置
type Character struct {
	PetID       string     // 桌宠ID
	Name        string     // 角色名称 (YAML: name)
	Persona     string     // 性格描述 (YAML: persona)
	PersonaType string     // 性格类型ID (YAML: persona_type)
	MBTI        MBTIConfig // MBTI 性格配置
	Avatar      string     // 头像/模型ID
	CreatedAt   time.Time  // 创建时间
	UpdatedAt   time.Time  // 更新时间
}

// 全局单例 CharacterStore
var (
	globalStore     *CharacterStore
	globalStoreOnce sync.Once
)

// GetCharacterStore 获取全局 CharacterStore 单例
// 如果尚未初始化，则使用默认工作区路径创建
func GetCharacterStore() *CharacterStore {
	globalStoreOnce.Do(func() {
		globalStore = newCharacterStore("")
	})
	return globalStore
}

// GetCharacterStoreWithPath 使用指定路径初始化全局单例
func GetCharacterStoreWithPath(workspacePath string) *CharacterStore {
	globalStoreOnce.Do(func() {
		globalStore = newCharacterStore(workspacePath)
	})
	return globalStore
}

// newCharacterStore 创建 CharacterStore 实例
func newCharacterStore(workspacePath string) *CharacterStore {
	cs := &CharacterStore{
		workspacePath: workspacePath,
		character: &Character{
			Name:        "小智",
			Persona:     "活泼可爱，精力充沛，喜欢开玩笑和撒娇，说话充满活力",
			PersonaType: "playful",
			MBTI:        DefaultMBTI(),
			Avatar:      "default",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	return cs
}

// Get 获取当前角色配置（线程安全）
func (s *CharacterStore) Get() *Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.character
}

// Update 更新角色配置并持久化到 MD 文件
func (s *CharacterStore) Update(char *Character) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	char.UpdatedAt = time.Now()
	if char.CreatedAt.IsZero() {
		char.CreatedAt = time.Now()
	}
	s.character = char

	return s.saveToFile()
}

// saveToFile 将角色配置保存到 MD 文件
// 格式：
// ---
// name: xxx
// persona: xxx
// persona_type: xxx
// mbti:
//
//	ie: 50
//	sn: 50
//	tf: 50
//	jp: 50
//
// avatar: xxx
// created_at: xxx
// updated_at: xxx
// ---
// # Soul
// ## Personality
// [性格描述]
func (s *CharacterStore) saveToFile() error {
	if s.workspacePath == "" {
		return nil // 没有设置工作区路径时不保存
	}

	dir := filepath.Join(s.workspacePath, "characters", "default")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create character dir: %w", err)
	}

	var buf bytes.Buffer

	// 写入 YAML frontmatter 头部
	buf.WriteString("---\n")

	// 基本信息
	fmt.Fprintf(&buf, "name: %s\n", s.character.Name)
	fmt.Fprintf(&buf, "persona: %s\n", s.character.Persona)
	fmt.Fprintf(&buf, "persona_type: %s\n", s.character.PersonaType)

	// MBTI 配置
	buf.WriteString("mbti:\n")
	fmt.Fprintf(&buf, "  ie: %d\n", s.character.MBTI.IE)
	fmt.Fprintf(&buf, "  sn: %d\n", s.character.MBTI.SN)
	fmt.Fprintf(&buf, "  tf: %d\n", s.character.MBTI.TF)
	fmt.Fprintf(&buf, "  jp: %d\n", s.character.MBTI.JP)

	// 其他字段
	fmt.Fprintf(&buf, "avatar: %s\n", s.character.Avatar)
	fmt.Fprintf(&buf, "created_at: %s\n", s.character.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&buf, "updated_at: %s\n", s.character.UpdatedAt.Format(time.RFC3339))

	// 写入 YAML frontmatter 结束
	buf.WriteString("---\n\n")

	// 写入 Soul 内容
	buf.WriteString("# Soul\n\n")
	buf.WriteString("## Personality\n\n")
	buf.WriteString(s.character.Persona + "\n")

	// 如果有额外的 soul 内容，追加
	if s.soulContent != "" {
		buf.WriteString("\n## Values\n\n")
		buf.WriteString(s.soulContent + "\n")
	}

	filePath := filepath.Join(dir, "config.md")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write character config: %w", err)
	}

	return nil
}

// Load 从 MD 文件加载角色配置
func (s *CharacterStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadFromFile()
}

func (s *CharacterStore) Change(petName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadFromOtherFile(petName)
}

// loadFromFile 从 MD 文件加载角色配置（内部方法）
func (s *CharacterStore) loadFromFile() error {
	if s.workspacePath == "" {
		return nil
	}

	filePath := filepath.Join(s.workspacePath, "characters", "default", "config.md")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在时使用默认配置
		}
		return fmt.Errorf("read character config: %w", err)
	}

	char, soul, err := parseCharacterFromMD(string(data))
	if err != nil {
		return fmt.Errorf("parse character config: %w", err)
	}

	s.character = char
	s.soulContent = soul
	return nil
}

// loadFromFile 从 MD 文件加载角色配置（内部方法）
func (s *CharacterStore) loadFromOtherFile(petName string) error {
	if s.workspacePath == "" {
		return nil
	}

	filePath := filepath.Join(s.workspacePath, "characters", petName, "config.md")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在时使用默认配置
		}
		return fmt.Errorf("read character config: %w", err)
	}

	char, soul, err := parseCharacterFromMD(string(data))
	if err != nil {
		return fmt.Errorf("parse character config: %w", err)
	}

	s.character = char
	s.soulContent = soul
	return nil
}

// parseCharacterFromMD 从 MD 格式解析角色配置
// 格式：
// ---
// name: xxx
// persona: xxx
// ...
// ---
// # Soul
// ## Personality
// xxx
func parseCharacterFromMD(content string) (*Character, string, error) {
	char := &Character{}

	// 分割 frontmatter 和 body
	parts := strings.Split(content, "---")
	if len(parts) < 3 {
		return char, "", nil
	}

	// 解析 frontmatter
	frontmatter := strings.TrimSpace(parts[1])
	lines := strings.Split(frontmatter, "\n")

	var inMBTI bool
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否在 mbti 块内
		if inMBTI {
			if strings.HasPrefix(line, "ie:") {
				char.MBTI.IE = parseInt(strings.TrimPrefix(line, "ie:"))
				inMBTI = false
			} else if strings.HasPrefix(line, "sn:") {
				char.MBTI.SN = parseInt(strings.TrimPrefix(line, "sn:"))
			} else if strings.HasPrefix(line, "tf:") {
				char.MBTI.TF = parseInt(strings.TrimPrefix(line, "tf:"))
			} else if strings.HasPrefix(line, "jp:") {
				char.MBTI.JP = parseInt(strings.TrimPrefix(line, "jp:"))
			}
			continue
		}

		// 解析各字段
		if strings.HasPrefix(line, "name:") {
			char.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "persona:") {
			char.Persona = strings.TrimSpace(strings.TrimPrefix(line, "persona:"))
		} else if strings.HasPrefix(line, "persona_type:") {
			char.PersonaType = strings.TrimSpace(strings.TrimPrefix(line, "persona_type:"))
		} else if strings.HasPrefix(line, "mbti:") {
			inMBTI = true
		} else if strings.HasPrefix(line, "avatar:") {
			char.Avatar = strings.TrimSpace(strings.TrimPrefix(line, "avatar:"))
		} else if strings.HasPrefix(line, "created_at:") {
			if t, err := time.Parse(time.RFC3339, strings.TrimSpace(strings.TrimPrefix(line, "created_at:"))); err == nil {
				char.CreatedAt = t
			}
		} else if strings.HasPrefix(line, "updated_at:") {
			if t, err := time.Parse(time.RFC3339, strings.TrimSpace(strings.TrimPrefix(line, "updated_at:"))); err == nil {
				char.UpdatedAt = t
			}
		}
	}

	// 解析 Soul body
	body := strings.Join(parts[2:], "---")
	soulContent := extractSoulContent(body)
	if char.Persona == "" {
		char.Persona = soulContent
	}

	return char, soulContent, nil
}

// extractSoulContent 从 body 中提取 Soul 内容
func extractSoulContent(body string) string {
	lines := strings.Split(body, "\n")
	var inPersonality, inValues bool
	var content []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Personality") {
			inPersonality = true
			inValues = false
			continue
		} else if strings.HasPrefix(line, "## Values") {
			inPersonality = false
			inValues = true
			continue
		} else if strings.HasPrefix(line, "#") {
			inPersonality = false
			inValues = false
		}

		if inPersonality || inValues {
			content = append(content, line)
		}
	}

	return strings.TrimSpace(strings.Join(content, "\n"))
}

// parseInt 解析字符串为整数
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	var val int
	fmt.Sscanf(s, "%d", &val)
	return val
}

// GetPersonaPrompt 获取用于注入 LLM 的 persona prompt
// 返回格式化的 Markdown 文本，用于拼接到 system prompt
func (s *CharacterStore) GetPersonaPrompt() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.character == nil {
		return ""
	}

	// 生成人格描述，包含 MBTI 特征
	mbtiDesc := describeMBTI(s.character.MBTI)

	return fmt.Sprintf(`【角色信息】
姓名：%s
性格类型：%s

人格特征：
%s

回复风格应体现以上人格特征,字数尽量少一些，是和用户聊天。`, s.character.Name, s.character.Persona, mbtiDesc)
}

// describeMBTI 将 MBTI 配置转换为文字描述
func describeMBTI(mbti MBTIConfig) string {
	var parts []string

	// IE 维度
	if mbti.IE > 60 {
		parts = append(parts, "偏外向，健谈热情")
	} else if mbti.IE < 40 {
		parts = append(parts, "偏内向，沉稳内敛")
	} else {
		parts = append(parts, "内外向平衡")
	}

	// SN 维度
	if mbti.SN > 60 {
		parts = append(parts, "重实际，重细节")
	} else if mbti.SN < 40 {
		parts = append(parts, "重直觉，爱想象")
	} else {
		parts = append(parts, "实际与直觉并存")
	}

	// TF 维度
	if mbti.TF > 60 {
		parts = append(parts, "偏理性，善分析")
	} else if mbti.TF < 40 {
		parts = append(parts, "偏感性，重情感")
	} else {
		parts = append(parts, "理性与感性平衡")
	}

	// JP 维度
	if mbti.JP > 60 {
		parts = append(parts, "有计划，善决策")
	} else if mbti.JP < 40 {
		parts = append(parts, "随性灵活，善适应")
	} else {
		parts = append(parts, "计划与灵活并存")
	}

	return strings.Join(parts, "，")
}

// IsOnboardingComplete 检查引导是否完成
func (s *CharacterStore) IsOnboardingComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.character != nil && s.character.Name != "" && s.character.Persona != ""
}

// SetSoulContent 设置额外的 Soul 内容（Values 等）
func (s *CharacterStore) SetSoulContent(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.soulContent = content
}

// GetSoulContent 获取 Soul 内容
func (s *CharacterStore) GetSoulContent() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.soulContent
}

// AddField 添加或更新自定义字段到 frontmatter
func (s *CharacterStore) AddField(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 目前简单实现，实际可以使用更复杂的 frontmatter 解析库
	// 这里暂时不支持动态字段，需要时可以扩展
	return nil
}
