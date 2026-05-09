package proactive

import (
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/pet/activity"
	"github.com/sipeed/picoclaw/pkg/pet/characters"
	petconfig "github.com/sipeed/picoclaw/pkg/pet/config"
	"github.com/sipeed/picoclaw/pkg/pet/userprofile"
)

// SnapshotDependencies 把 BuildSnapshot 依赖到的外部能力显式列出来。
//
// 这样做的好处是：
// - 依赖关系清晰
// - 以后写测试时方便构造
// - 不需要把整个 PetService 都塞进来
type SnapshotDependencies struct {
	ActivityStore      *activity.Store
	ConfigManager      *petconfig.Manager
	UserProfileManager *userprofile.Manager
	CharacterProvider  *characters.Manager
	LastPushAt         func() time.Time
}

// BuildSnapshot 把“当前和主动性有关的状态”统一收敛成一个 Snapshot。
//
// 可以把它理解成：
// “如果桌宠现在想决定要不要主动一下，它眼里能看到哪些东西？”
func BuildSnapshot(now time.Time, deps SnapshotDependencies) Snapshot {
	if now.IsZero() {
		now = time.Now()
	}
	var snap Snapshot
	snap.Now = now

	// 先拿桌宠当前角色和情绪。
	charID := ""
	if deps.CharacterProvider != nil {
		charID = deps.CharacterProvider.GetCurrentID()
		char := deps.CharacterProvider.GetCurrent()
		if char != nil {
			snap.Pet.CharacterID = char.ID
			snap.Pet.PersonaType = char.PersonaType
			dominant, score := char.GetEmotionEngine().GetDominantEmotion()
			snap.Pet.DominantEmotion = strings.ToLower(dominant)
			snap.Pet.EmotionScore = score
		}
	}

	// 再补用户画像和用户状态。
	if deps.UserProfileManager != nil && charID != "" {
		profile := deps.UserProfileManager.LoadProfile()
		if profile != nil {
			snap.User.DisplayName = profile.DisplayName
			snap.User.Chronotype = profile.Chronotype
			snap.User.PersonalityTone = profile.PersonalityTone
			snap.User.PressureLevel = strings.ToLower(profile.PressureLevel)
			// 渲染层真正想要的是“阴阳怪气 / 抽象发疯 / 甜心夹子”这类语气标签。
			// 这个值应优先来自用户/产品层显式配置，而不是直接拿 pet_config 里的 persona 描述文本顶上。
			snap.Pet.PersonalityTone = profile.PersonalityTone
		}
		state := deps.UserProfileManager.LoadState(charID)
		if state != nil {
			snap.User.CurrentMood = state.CurrentMood
			snap.User.EnergyLevel = state.EnergyLevel
			snap.User.EngagementLevel = state.EngagementLevel
			snap.User.StressTrend = state.StressTrend
		}
	}

	// 再补主动性相关配置。
	if deps.ConfigManager != nil {
		app := deps.ConfigManager.GetApp()
		if app != nil {
			snap.Preferences.ProactiveCare = app.ProactiveCare
			snap.Preferences.ProactiveIntervalMins = app.ProactiveIntervalMinutes
			snap.Preferences.WeeklyReportEnabled = app.WeeklyReportEnabled
			snap.Preferences.ProgressNudgeEnabled = app.ProgressNudgeEnabled
			snap.Preferences.ProactiveCheckMinutes = app.ProactiveCheckMinutes
			snap.Preferences.GlobalCooldownMins = app.GlobalCooldownMinutes
		}
	}

	// 如果用户侧没有显式 personality tone，渲染层就只依赖 persona_type 做兜底风格分支。
	// 这里不再把 char.Persona 这种长文本描述塞进 PersonalityTone，避免混淆“角色设定”和“输出语气标签”。

	// 最后补最近一次主动触达时间，供冷却使用。
	if deps.LastPushAt != nil {
		snap.Activity.LastPushAt = deps.LastPushAt()
	}

	// 活动记录是周报和催办的事实基础。
	// 第一阶段先只统计最近 24 小时的消息数、任务数和未完成事项数。
	if deps.ActivityStore != nil {
		start := now.Add(-24 * time.Hour)
		events, err := deps.ActivityStore.ListRange(start, now)
		if err == nil {
			for _, ev := range events {
				if charID != "" && ev.CharacterID != "" && ev.CharacterID != charID {
					continue
				}
				switch ev.Type {
				case activity.EventUserMessage:
					snap.Activity.RecentMessageCount++
					if ev.CreatedAt.After(snap.Activity.LastUserMessageAt) {
						snap.Activity.LastUserMessageAt = ev.CreatedAt
					}
				case activity.EventToolCall:
					snap.Activity.RecentTaskCount++
				case activity.EventTaskResult:
					snap.Activity.RecentTaskCount++
					if ev.Status == activity.StatusPending || ev.Status == activity.StatusFailed {
						snap.Activity.UnfinishedTaskCount++
					}
				case activity.EventToolResult:
					if ev.Status == activity.StatusFailed {
						snap.Activity.RecentTaskCount++
						snap.Activity.UnfinishedTaskCount++
					}
				}
			}
		}
	}

	return snap
}
