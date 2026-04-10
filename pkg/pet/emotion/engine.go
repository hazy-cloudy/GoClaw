package emotion

import (
	"sync"
	"time"
)

// =============================================================================
// 情绪常量定义
// =============================================================================

// 六大基础情绪标签
// 基于心理学基础情绪理论，每种情绪都有0-100的强度值，50为平静状态
const (
	EmotionJoy      = "joy"      // 快乐：积极正向情绪，表示愉悦、开心的状态
	EmotionAnger    = "anger"    // 愤怒：强烈负面情绪，通常由挫败或受威胁引发
	EmotionSadness  = "sadness"  // 悲伤：负面情绪，表示失落、难过的状态
	EmotionDisgust  = "disgust"  // 厌恶：强烈负面情绪，表示反感、讨厌
	EmotionSurprise = "surprise" // 惊讶：中性偏正向的强烈情绪，表示意外、震惊
	EmotionFear     = "fear"     // 恐惧：强烈负面情绪，表示害怕、担忧
	EmotionNeutral  = "neutral"  // 中性：无明显情绪，保持平静

	// 情绪阈值：当情绪值超过此范围时，视为"强烈情绪"
	// 用于判断是否需要推送到客户端触发特殊表情/动画
	ThresholdHigh = 80 // 高阈值：超过此值表示情绪非常强烈
	ThresholdLow  = 20 // 低阈值：低于此值表示情绪非常低落
	NeutralValue  = 50 // 中性值：情绪的平衡点，远离此值表示情绪明显

	// 衰减速率：每秒向中性回归的比例
	// 实现"时间抚平一切情绪"的生理规律
	// 每秒衰减1%，即情绪值每秒向50靠近1%
	DecayRate = 0.01
)

// =============================================================================
// 数据结构定义
// =============================================================================

// SixEmotions 六大基础情绪状态
// 每种情绪独立评分(0-100)，50为平静状态
// - 0:   完全负面极端（如极度悲伤）
// - 50:  平静/中性
// - 100: 完全正面极端（如极度快乐）
type SixEmotions struct {
	Joy      int `json:"joy"`      // 快乐
	Anger    int `json:"anger"`    // 愤怒
	Sadness  int `json:"sadness"`  // 悲伤
	Disgust  int `json:"disgust"`  // 厌恶
	Surprise int `json:"surprise"` // 惊讶
	Fear     int `json:"fear"`     // 恐惧
}

// MBTIPersonality MBTI性格四维配置
// 用于长期人格漂移的追踪
// 每个维度0-100，50为中心点
// - IE: 内向(I) <-> 外向(E)，50为中性
// - SN: 实感(S) <-> 直觉(N)，50为中性
// - TF: 理性(T) <-> 感性(F)，50为中性
// - JP: 判断(J) <-> 感知(P)，50为中性
type MBTIPersonality struct {
	IE int `json:"ie"` // 内向-外向维度
	SN int `json:"sn"` // 实感-直觉维度
	TF int `json:"tf"` // 理性-感性维度
	JP int `json:"jp"` // 判断-感知维度
}

// EmotionEngine 情绪状态机
// 核心组件，负责：
// 1. 维护六大情绪的实时状态
// 2. MBTI性格四轴的长期漂移
// 3. 情绪衰减计算（稳态回归）
// 4. 阈值检测与推送判断
type EmotionEngine struct {
	personality MBTIPersonality // MBTI性格配置（长期）
	emotions    SixEmotions     // 六大情绪状态（当前）
	volatility  float64         // 情绪波动系数，影响每次变化的幅度（越大变化越剧烈）
	lastUpdate  time.Time       // 最后更新时间，用于计算衰减

	persistPath string       // 持久化路径，保存到哪个目录
	mu          sync.RWMutex // 读写锁，保证并发安全
}

// EmotionPush 情绪变化推送数据结构
// 当情绪触发阈值时，推送给客户端
type EmotionPush struct {
	Emotion string `json:"emotion"`          // 情绪标签
	Score   int    `json:"score"`            // 情绪强度值
	Motion  string `json:"motion,omitempty"` // 关联的动作（可选）
	Prompt  string `json:"prompt,omitempty"` // 阈值触发时的Prompt，用于注入LLM
}

// =============================================================================
// 构造函数
// =============================================================================

// NewEmotionEngine 创建情绪引擎实例
// 参数：
//   - persistPath: 持久化存储路径，如 "/home/user/picoclaw"
//     如果为空，则不进行持久化存储
//
// 初始状态：
//   - 所有情绪值为50（中性）
//   - MBTI四维都为50（中性）
//   - 波动系数为1.0（默认）
//   - 最后更新为当前时间
func NewEmotionEngine(persistPath string) *EmotionEngine {
	return &EmotionEngine{
		personality: MBTIPersonality{IE: NeutralValue, SN: NeutralValue, TF: NeutralValue, JP: NeutralValue},
		emotions:    SixEmotions{Joy: NeutralValue, Anger: NeutralValue, Sadness: NeutralValue, Disgust: NeutralValue, Surprise: NeutralValue, Fear: NeutralValue},
		volatility:  1.0,
		lastUpdate:  time.Now(),
		persistPath: persistPath,
	}
}

// =============================================================================
// Getter方法（读操作）
// =============================================================================

// GetEmotions 获取当前六大情绪状态
// 返回当前情绪的拷贝，保证内部状态不被外部修改
func (e *EmotionEngine) GetEmotions() SixEmotions {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.emotions
}

// GetPersonality 获取当前MBTI性格配置
// 返回拷贝，防止外部修改
func (e *EmotionEngine) GetPersonality() MBTIPersonality {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.personality
}

// GetVolatility 获取当前波动系数
// 波动系数影响情绪变化的幅度
func (e *EmotionEngine) GetVolatility() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.volatility
}

// GetLastUpdate 获取最后更新时间
// 用于计算衰减
func (e *EmotionEngine) GetLastUpdate() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastUpdate
}

// =============================================================================
// 情绪设置与更新
// =============================================================================

// SetEmotion 设置单个情绪值
// 参数：
//   - emotion: 情绪标签（joy/anger/sadness/disgust/surprise/fear）
//   - score: 情绪强度值，会被clamp到0-100范围
//
// 注意：
//   - 自动更新lastUpdate为当前时间
//   - 线程安全
func (e *EmotionEngine) SetEmotion(emotion string, score int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastUpdate = time.Now()

	switch emotion {
	case EmotionJoy:
		e.emotions.Joy = clamp(score)
	case EmotionAnger:
		e.emotions.Anger = clamp(score)
	case EmotionSadness:
		e.emotions.Sadness = clamp(score)
	case EmotionDisgust:
		e.emotions.Disgust = clamp(score)
	case EmotionSurprise:
		e.emotions.Surprise = clamp(score)
	case EmotionFear:
		e.emotions.Fear = clamp(score)
	}
}

// UpdateFromLLMTags 从LLM标签批量更新情绪
// 参数：
//   - tags: LLM返回的情绪标签数组
//
// LLM标签格式：[emotion:joy:80]、[emotion:angry:30]等
//
// 逻辑：
//  1. 解析每个标签
//  2. 更新对应情绪值
//  3. 自动更新lastUpdate
func (e *EmotionEngine) UpdateFromLLMTags(tags []EmotionTag) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastUpdate = time.Now()

	for _, tag := range tags {
		switch tag.Emotion {
		case EmotionJoy:
			e.emotions.Joy = clamp(tag.Score)
		case EmotionAnger:
			e.emotions.Anger = clamp(tag.Score)
		case EmotionSadness:
			e.emotions.Sadness = clamp(tag.Score)
		case EmotionDisgust:
			e.emotions.Disgust = clamp(tag.Score)
		case EmotionSurprise:
			e.emotions.Surprise = clamp(tag.Score)
		case EmotionFear:
			e.emotions.Fear = clamp(tag.Score)
		}
	}
}

// UpdateFromLLMTagsDelta 从LLM标签更新情绪（支持增量格式）
// 参数：
//   - tags: 解析出的情绪标签，支持绝对值和增量值
//
// 增量格式：[emotion:joy:+5] 表示在当前值基础上+5
// 绝对格式：[emotion:joy:75] 表示设置为75
func (e *EmotionEngine) UpdateFromLLMTagsDelta(tags []EmotionTag) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastUpdate = time.Now()

	for _, tag := range tags {
		switch tag.Emotion {
		case EmotionJoy:
			if tag.IsDelta {
				e.emotions.Joy = clamp(e.emotions.Joy + tag.Delta)
			} else {
				e.emotions.Joy = clamp(tag.Score)
			}
		case EmotionAnger:
			if tag.IsDelta {
				e.emotions.Anger = clamp(e.emotions.Anger + tag.Delta)
			} else {
				e.emotions.Anger = clamp(tag.Score)
			}
		case EmotionSadness:
			if tag.IsDelta {
				e.emotions.Sadness = clamp(e.emotions.Sadness + tag.Delta)
			} else {
				e.emotions.Sadness = clamp(tag.Score)
			}
		case EmotionDisgust:
			if tag.IsDelta {
				e.emotions.Disgust = clamp(e.emotions.Disgust + tag.Delta)
			} else {
				e.emotions.Disgust = clamp(tag.Score)
			}
		case EmotionSurprise:
			if tag.IsDelta {
				e.emotions.Surprise = clamp(e.emotions.Surprise + tag.Delta)
			} else {
				e.emotions.Surprise = clamp(tag.Score)
			}
		case EmotionFear:
			if tag.IsDelta {
				e.emotions.Fear = clamp(e.emotions.Fear + tag.Delta)
			} else {
				e.emotions.Fear = clamp(tag.Score)
			}
		}
	}
}

// =============================================================================
// 情绪衰减（稳态回归）
// =============================================================================

// ApplyDecay 应用时间衰减，使情绪逐渐回归平静
// 参数：
//   - elapsed: 逝去的时间
//
// 算法：
//
//	每秒衰减公式：emotion_new = emotion_old + (50 - emotion_old) * DecayRate * volatility
//
// 特点：
//   - 离50越远，回归速度越快（指数衰减）
//   - 波动系数volatility影响衰减速度
//   - 50时完全不变
//
// 示例：
//   - 愤怒值80，经过10秒后：
//     衰减因子 = 0.01 * 10 * 1.0 = 0.1
//     新值 = 80 + (50-80)*0.1 = 77
func (e *EmotionEngine) ApplyDecay(elapsed time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	seconds := elapsed.Seconds()
	if seconds <= 0 {
		return
	}

	// 计算衰减因子 = 基础衰减率 * 时间(秒) * 波动系数
	decayFactor := DecayRate * seconds * e.volatility

	// 对每个情绪值应用衰减
	// 公式：new = old + (50 - old) * decayFactor
	// 这是一个指数衰减向50靠近的过程
	e.emotions.Joy = int(float64(e.emotions.Joy) + (float64(NeutralValue-e.emotions.Joy) * decayFactor))
	e.emotions.Anger = int(float64(e.emotions.Anger) + (float64(NeutralValue-e.emotions.Anger) * decayFactor))
	e.emotions.Sadness = int(float64(e.emotions.Sadness) + (float64(NeutralValue-e.emotions.Sadness) * decayFactor))
	e.emotions.Disgust = int(float64(e.emotions.Disgust) + (float64(NeutralValue-e.emotions.Disgust) * decayFactor))
	e.emotions.Surprise = int(float64(e.emotions.Surprise) + (float64(NeutralValue-e.emotions.Surprise) * decayFactor))
	e.emotions.Fear = int(float64(e.emotions.Fear) + (float64(NeutralValue-e.emotions.Fear) * decayFactor))

	// 确保值在有效范围内
	e.emotions.Joy = clamp(e.emotions.Joy)
	e.emotions.Anger = clamp(e.emotions.Anger)
	e.emotions.Sadness = clamp(e.emotions.Sadness)
	e.emotions.Disgust = clamp(e.emotions.Disgust)
	e.emotions.Surprise = clamp(e.emotions.Surprise)
	e.emotions.Fear = clamp(e.emotions.Fear)
}

// =============================================================================
// 情绪状态查询
// =============================================================================

// GetDominantEmotion 获取当前主导情绪
// 返回：
//   - string: 主导情绪标签（偏离50最远的情绪）
//   - int: 该情绪的强度值
//
// 算法：
//
//	找出六大情绪中偏离中性值50最远的情绪
//	即 abs(score - 50) 最大的那个
//
// 示例：
//   - joy=70, anger=50, others=50 → 返回 ("joy", 70)
//   - joy=60, sad=80, others=50 → 返回 ("sadness", 80) 因为|80-50|=30 > |60-50|=10
func (e *EmotionEngine) GetDominantEmotion() (string, int) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	emotions := map[string]int{
		EmotionJoy:      e.emotions.Joy,
		EmotionAnger:    e.emotions.Anger,
		EmotionSadness:  e.emotions.Sadness,
		EmotionDisgust:  e.emotions.Disgust,
		EmotionSurprise: e.emotions.Surprise,
		EmotionFear:     e.emotions.Fear,
	}

	dominant := EmotionNeutral
	maxScore := NeutralValue
	maxDeviation := 0

	// 遍历所有情绪，找出偏离中性最远的
	for emo, score := range emotions {
		deviation := abs(score - NeutralValue)
		if deviation > maxDeviation {
			maxDeviation = deviation
			dominant = emo
			maxScore = score
		}
	}

	return dominant, maxScore
}

// ShouldPush 判断是否应该推送情绪变化
// 返回：
//   - bool: 是否应该推送
//   - EmotionPush: 如果应该推送，包含推送数据
//
// 推送条件：
//   - 任意情绪值 > ThresholdHigh(80)
//   - 或任意情绪值 < ThresholdLow(20)
//
// 这个阈值用于触发客户端的特殊表情/动画效果
func (e *EmotionEngine) ShouldPush() (bool, EmotionPush) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	emotions := map[string]int{
		EmotionJoy:      e.emotions.Joy,
		EmotionAnger:    e.emotions.Anger,
		EmotionSadness:  e.emotions.Sadness,
		EmotionDisgust:  e.emotions.Disgust,
		EmotionSurprise: e.emotions.Surprise,
		EmotionFear:     e.emotions.Fear,
	}

	// 检查每个情绪是否触发阈值
	for emo, score := range emotions {
		if score > ThresholdHigh || score < ThresholdLow {
			return true, EmotionPush{
				Emotion: emo,
				Score:   score,
				Prompt:  e.getThresholdPrompt(emo, score),
			}
		}
	}

	return false, EmotionPush{}
}

// =============================================================================
// 辅助方法
// =============================================================================

// getThresholdPrompt 生成阈值触发时的Prompt
// 用于注入LLM，让LLM知道当前处于某种强烈情绪状态
//
// 返回格式：
//   - score > 80: "【当前状态】：非常{情绪中文}"
//   - score < 20: "【当前状态】：有点{情绪中文}"
func (e *EmotionEngine) getThresholdPrompt(emotion string, score int) string {
	desc := emotionToChinese(emotion)

	if score > ThresholdHigh {
		return "【当前状态】：非常" + desc
	}
	return "【当前状态】：有点" + desc
}

// emotionToChinese 情绪标签转中文描述
func emotionToChinese(emotion string) string {
	switch emotion {
	case EmotionJoy:
		return "开心"
	case EmotionAnger:
		return "生气"
	case EmotionSadness:
		return "难过"
	case EmotionDisgust:
		return "厌恶"
	case EmotionSurprise:
		return "惊讶"
	case EmotionFear:
		return "害怕"
	default:
		return "平静"
	}
}

// clamp 将值限制在0-100范围内
// 用于确保情绪值在有效范围内
func clamp(val int) int {
	if val < 0 {
		return 0
	}
	if val > 100 {
		return 100
	}
	return val
}

// abs 取绝对值
func abs(val int) int {
	if val < 0 {
		return -val
	}
	return val
}

// =============================================================================
// 标签数据结构
// =============================================================================

// EmotionTag LLM返回的情绪标签
// 格式：
//   - 绝对值：[emotion:joy:80] - 设置为80
//   - 增量值：[emotion:joy:+5] - 增加5
//   - 增量值：[emotion:joy:-3] - 减少3
type EmotionTag struct {
	Emotion string // 情绪名称
	Score   int    // 情绪强度（绝对值时使用）
	IsDelta bool   // 是否为增量值
	Delta   int    // 增量值（IsDelta=true时使用）
}
