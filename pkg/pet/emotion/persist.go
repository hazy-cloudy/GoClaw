package emotion

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// =============================================================================
// 持久化常量
// =============================================================================

const (
	// EmotionStateFile 情绪状态文件名
	// 存储位置: {workspacePath}/pet/emotion_state.json
	EmotionStateFile = "emotion_state.json"

	// Version 持久化版本号
	// 用于未来数据迁移兼容
	Version = 1
)

// =============================================================================
// 持久化数据结构
// =============================================================================

// EmotionSnapshot 情绪状态快照
// 用于将情绪状态持久化到JSON文件
// 包含完整的情绪状态和MBTI配置
type EmotionSnapshot struct {
	Version     int             `json:"version"`     // 格式版本号
	LastUpdate  int64           `json:"last_update"` // 最后更新时间（Unix时间戳）
	Emotions    SixEmotions     `json:"emotions"`    // 六大情绪状态
	Personality MBTIPersonality `json:"personality"` // MBTI性格配置
	Volatility  float64         `json:"volatility"`  // 情绪波动系数
}

// =============================================================================
// 持久化方法
// =============================================================================

// Save 将当前情绪状态保存到文件
// 保存路径: {persistPath}/pet/emotion_state.json
//
// 调用时机：
//   - 每次LLM响应后，情绪标签被解析并更新时
//   - 确保即使程序崩溃，最近的状态也被保存
//
// 持久化内容：
//   - 六大情绪当前值
//   - MBTI性格配置
//   - 情绪波动系数
//   - 最后更新时间（用于重启后计算衰减）
//
// 注意：
//   - 如果persistPath为空，则不进行保存
//   - 保存是异步安全的（读取时持有读锁）
func (e *EmotionEngine) Save() error {
	// 如果没有设置持久化路径，直接返回
	if e.persistPath == "" {
		return nil
	}

	// 读取当前状态（持有读锁）
	e.mu.RLock()
	snapshot := EmotionSnapshot{
		Version:     Version,
		LastUpdate:  e.lastUpdate.Unix(),
		Emotions:    e.emotions,
		Personality: e.personality,
		Volatility:  e.volatility,
	}
	e.mu.RUnlock()

	// 创建目录
	dir := filepath.Join(e.persistPath, "pet")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create emotion dir: %w", err)
	}

	// 写入文件
	filePath := filepath.Join(dir, EmotionStateFile)
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal emotion snapshot: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write emotion state: %w", err)
	}

	return nil
}

// Load 从文件加载情绪状态
// 加载路径: {persistPath}/pet/emotion_state.json
//
// 调用时机：
//   - PetService初始化时
//   - 确保重启后能恢复之前的情绪状态
//
// 恢复流程：
//  1. 读取文件中的last_update时间戳
//  2. 计算程序关闭的时长（elapsed）
//  3. 应用衰减算法，让情绪逐渐回归平静
//  4. 恢复所有情绪值、MBTI配置和波动系数
//
// 衰减计算：
//   - 假设关闭了N秒
//   - 情绪值会向中性50衰减
//   - 衰减速度由DecayRate和volatility控制
//
// 注意：
//   - 如果persistPath为空或文件不存在，使用默认状态
//   - 加载时持有写锁
func (e *EmotionEngine) Load() error {
	// 如果没有设置持久化路径，使用默认状态
	if e.persistPath == "" {
		return nil
	}

	filePath := filepath.Join(e.persistPath, "pet", EmotionStateFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，使用默认状态（正常首次运行）
			return nil
		}
		return fmt.Errorf("read emotion state: %w", err)
	}

	var snapshot EmotionSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("unmarshal emotion snapshot: %w", err)
	}

	// 恢复状态
	e.mu.Lock()
	defer e.mu.Unlock()

	// 计算衰减
	lastUpdate := time.Unix(snapshot.LastUpdate, 0)
	elapsed := time.Since(lastUpdate)

	// 恢复数据
	e.emotions = snapshot.Emotions
	e.personality = snapshot.Personality
	e.volatility = snapshot.Volatility
	e.lastUpdate = lastUpdate

	// 应用衰减（让情绪回归平静）
	e.applyDecayLocked(elapsed)

	return nil
}

// =============================================================================
// 内部衰减方法（持有写锁）
// =============================================================================

// applyDecayLocked 应用时间衰减（内部方法，调用前必须持有写锁）
// 参数：
//   - elapsed: 从lastUpdate到现在逝去的时间
//
// 算法：
//
//	对于每个情绪值，计算从lastUpdate到现在应该衰减的程度
//
//	decayFactor = DecayRate * elapsed_seconds * volatility
//	new_emotion = old_emotion + (50 - old_emotion) * decayFactor
//
// 示例：
//   - 上次保存时愤怒值为80
//   - 程序关闭了60秒
//   - 衰减因子 = 0.01 * 60 * 1.0 = 0.6
//   - 新愤怒值 = 80 + (50-80)*0.6 = 80 - 18 = 62
//
// 注意：
//   - 此方法假设调用者已经持有mu.Lock()
//   - 不负责更新lastUpdate时间
func (e *EmotionEngine) applyDecayLocked(elapsed time.Duration) {
	seconds := elapsed.Seconds()
	if seconds <= 0 {
		return
	}

	// 计算衰减因子
	decayFactor := DecayRate * seconds * e.volatility

	// 对每个情绪应用衰减
	// 公式: new = old + (50 - old) * decayFactor
	// 这会让值逐渐向50回归
	e.emotions.Joy = clamp(int(float64(e.emotions.Joy) + (float64(NeutralValue-e.emotions.Joy) * decayFactor)))
	e.emotions.Anger = clamp(int(float64(e.emotions.Anger) + (float64(NeutralValue-e.emotions.Anger) * decayFactor)))
	e.emotions.Sadness = clamp(int(float64(e.emotions.Sadness) + (float64(NeutralValue-e.emotions.Sadness) * decayFactor)))
	e.emotions.Disgust = clamp(int(float64(e.emotions.Disgust) + (float64(NeutralValue-e.emotions.Disgust) * decayFactor)))
	e.emotions.Surprise = clamp(int(float64(e.emotions.Surprise) + (float64(NeutralValue-e.emotions.Surprise) * decayFactor)))
	e.emotions.Fear = clamp(int(float64(e.emotions.Fear) + (float64(NeutralValue-e.emotions.Fear) * decayFactor)))
}
